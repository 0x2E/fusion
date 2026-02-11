package store

import (
	"database/sql"
	"fmt"
)

func createCurrentSchema(tx *sql.Tx) error {
	content, err := migrationFiles.ReadFile("migrations/001_initial.sql")
	if err != nil {
		return fmt.Errorf("read baseline schema: %w", err)
	}

	if _, err := tx.Exec(string(content)); err != nil {
		return fmt.Errorf("create current schema: %w", err)
	}

	return nil
}

func createMigrationsTableTx(tx *sql.Tx) error {
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at INTEGER NOT NULL DEFAULT (unixepoch())
		)
	`); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	return nil
}

func prepareLegacyFeedMapTx(tx *sql.Tx, feedMap map[int64]int64) error {
	if _, err := tx.Exec(`
		CREATE TEMP TABLE IF NOT EXISTS legacy_feed_map (
			old_id INTEGER PRIMARY KEY,
			new_id INTEGER NOT NULL
		);
		DELETE FROM legacy_feed_map;
	`); err != nil {
		return fmt.Errorf("create temporary legacy feed map: %w", err)
	}

	if len(feedMap) == 0 {
		return nil
	}

	insertStmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO legacy_feed_map (old_id, new_id)
		VALUES (:old_id, :new_id)
	`)
	if err != nil {
		return fmt.Errorf("prepare temporary legacy feed map insert: %w", err)
	}
	defer insertStmt.Close()

	for oldID, newID := range feedMap {
		if _, err := insertStmt.Exec(
			sql.Named("old_id", oldID),
			sql.Named("new_id", newID),
		); err != nil {
			return fmt.Errorf("insert temporary legacy feed map %d->%d: %w", oldID, newID, err)
		}
	}

	return nil
}

func prepareLegacyItemMapTx(tx *sql.Tx) error {
	if _, err := tx.Exec(`
		CREATE TEMP TABLE IF NOT EXISTS legacy_item_map (
			old_id INTEGER PRIMARY KEY,
			new_id INTEGER NOT NULL
		);
		DELETE FROM legacy_item_map;
	`); err != nil {
		return fmt.Errorf("create temporary legacy item map: %w", err)
	}

	return nil
}
