package store

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func (s *Store) migrate() error {
	if err := s.createMigrationsTable(); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	applied, err := s.getAppliedVersions()
	if err != nil {
		return fmt.Errorf("get applied versions: %w", err)
	}

	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		version, err := extractVersion(entry.Name())
		if err != nil {
			return fmt.Errorf("invalid migration filename %s: %w", entry.Name(), err)
		}

		if applied[version] {
			continue
		}

		if err := s.applyMigration(version, entry.Name()); err != nil {
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}
	}

	return nil
}

func (s *Store) createMigrationsTable() error {
	schema := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at INTEGER NOT NULL DEFAULT (unixepoch())
	);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *Store) getAppliedVersions() (map[int]bool, error) {
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

func (s *Store) applyMigration(version int, filename string) error {
	content, err := migrationFiles.ReadFile("migrations/" + filename)
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
		"INSERT INTO schema_migrations (version) VALUES (:version)",
		sql.Named("version", version),
	); err != nil {
		return fmt.Errorf("record version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func extractVersion(filename string) (int, error) {
	if !strings.HasSuffix(filename, ".sql") {
		return 0, fmt.Errorf("not a .sql file")
	}

	parts := strings.SplitN(filename, "_", 2)
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid format, expected NNN_description.sql")
	}

	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid version number: %w", err)
	}

	return version, nil
}
