package store

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// baselineMigrationVersion001 is the integer version parsed from
// migration file name `001_initial.sql`.
const baselineMigrationVersion001 = 1

func prepareLegacyDatabase(dbPath string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	shouldMigrate, err := shouldMigrateLegacyDB(db)
	if err != nil {
		return fmt.Errorf("detect legacy schema: %w", err)
	}
	if !shouldMigrate {
		slog.Debug("legacy database migration skipped", "reason", "schema already current")
		return nil
	}

	slog.Info("legacy database migration started", "db_path", dbPath)

	// Flush WAL pages back into the main DB file before file-copy backup.
	// Without this, copying only *.db may miss recent committed data in *.db-wal.
	if _, err := db.Exec(`PRAGMA wal_checkpoint(TRUNCATE)`); err != nil {
		return fmt.Errorf("checkpoint wal before legacy migration: %w", err)
	}

	backupPath := fmt.Sprintf("%s.bak.%d", dbPath, time.Now().Unix())
	if err := copyFile(dbPath, backupPath); err != nil {
		return fmt.Errorf("backup legacy database: %w", err)
	}
	slog.Info("legacy database backup created", "backup_path", backupPath)

	if err := db.Close(); err != nil {
		return fmt.Errorf("close source database before migration: %w", err)
	}

	tempPath := fmt.Sprintf("%s.migrating.%d", dbPath, time.Now().UnixNano())
	if err := buildMigratedDatabase(dbPath, tempPath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("build migrated database: %w", err)
	}

	if err := replaceDatabaseFile(dbPath, tempPath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("replace database file: %w", err)
	}

	slog.Info("legacy database migration finished", "db_path", dbPath, "backup_path", backupPath)

	return nil
}

func buildMigratedDatabase(legacyPath, targetPath string) error {
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}

	targetDB, err := sql.Open("sqlite", targetPath)
	if err != nil {
		return fmt.Errorf("open target database: %w", err)
	}
	defer targetDB.Close()

	if err := targetDB.Ping(); err != nil {
		return fmt.Errorf("ping target database: %w", err)
	}

	// Attach the old DB as a secondary schema so we can copy data with
	// SQL like INSERT INTO main_table SELECT ... FROM legacy.old_table.
	if _, err := targetDB.Exec(`ATTACH DATABASE :legacy_path AS legacy`, sql.Named("legacy_path", legacyPath)); err != nil {
		return fmt.Errorf("attach legacy database: %w", err)
	}
	defer targetDB.Exec(`DETACH DATABASE legacy`)

	tx, err := targetDB.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := createCurrentSchema(tx); err != nil {
		return err
	}
	groupMap, err := copyLegacyGroups(tx)
	if err != nil {
		return err
	}
	if err := restoreDefaultGroupFromLegacy(tx); err != nil {
		return err
	}
	feedMap, err := copyLegacyFeeds(tx, groupMap)
	if err != nil {
		return err
	}
	if err := prepareLegacyFeedMapTx(tx, feedMap); err != nil {
		return err
	}
	if err := copyLegacyItems(tx); err != nil {
		return err
	}
	if err := copyLegacyBookmarks(tx); err != nil {
		return err
	}
	if err := createMigrationsTableTx(tx); err != nil {
		return err
	}

	if _, err := tx.Exec(
		`INSERT OR IGNORE INTO schema_migrations (version) VALUES (:version)`,
		sql.Named("version", baselineMigrationVersion001),
	); err != nil {
		return fmt.Errorf("record baseline migration version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	if _, err := targetDB.Exec(`PRAGMA wal_checkpoint(TRUNCATE)`); err != nil {
		return fmt.Errorf("checkpoint migrated database wal: %w", err)
	}

	var integrity string
	// Validate the newly built DB before replacing the original file.
	if err := targetDB.QueryRow(`PRAGMA integrity_check`).Scan(&integrity); err != nil {
		return fmt.Errorf("integrity check: %w", err)
	}
	if integrity != "ok" {
		return fmt.Errorf("integrity check failed: %s", integrity)
	}

	return nil
}
