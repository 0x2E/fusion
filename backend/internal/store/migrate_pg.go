package store

import (
	"embed"
	"fmt"
	"log/slog"
	"sort"
	"time"
)

//go:embed migrations/pg/*.sql
var pgMigrationFiles embed.FS

func (s *PGStore) migrate() error {
	startedAt := time.Now()
	slog.Info("database migration started")

	if err := s.createMigrationsTable(); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	applied, err := s.getAppliedVersions()
	if err != nil {
		return fmt.Errorf("get applied versions: %w", err)
	}

	entries, err := pgMigrationFiles.ReadDir("migrations/pg")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	appliedCount := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		version, err := extractVersion(entry.Name())
		if err != nil {
			return fmt.Errorf("invalid migration filename %s: %w", entry.Name(), err)
		}

		if applied[version] {
			slog.Debug("migration already applied", "version", version, "file", entry.Name())
			continue
		}

		slog.Info("applying migration", "version", version, "file", entry.Name())
		if err := s.applyPGMigration(version, entry.Name()); err != nil {
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}

		appliedCount++
		slog.Info("migration applied", "version", version, "file", entry.Name())
	}

	slog.Info(
		"database migration finished",
		"applied", appliedCount,
		"duration", time.Since(startedAt),
	)

	return nil
}

func (s *PGStore) createMigrationsTable() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    BIGINT PRIMARY KEY,
			applied_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::bigint
		)
	`)
	return err
}

func (s *PGStore) getAppliedVersions() (map[int]bool, error) {
	rows, err := s.db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

func (s *PGStore) applyPGMigration(version int, filename string) error {
	content, err := pgMigrationFiles.ReadFile("migrations/pg/" + filename)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(string(content)); err != nil {
		return fmt.Errorf("exec migration: %w", err)
	}

	if _, err := tx.Exec(
		"INSERT INTO schema_migrations (version) VALUES ($1)",
		version,
	); err != nil {
		return fmt.Errorf("record version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
