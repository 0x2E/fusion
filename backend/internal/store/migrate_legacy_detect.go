package store

import (
	"database/sql"
	"fmt"
)

func shouldMigrateLegacyDB(db *sql.DB) (bool, error) {
	// Checking only schema_migrations is not enough:
	// 1) brand-new empty DB has no migration table but is not a legacy DB;
	// 2) interrupted startup may leave schema_migrations present but empty;
	// 3) partially-migrated/corrupted legacy DB may still contain marker columns.
	// We therefore prioritize legacy marker columns for decision-making.

	for _, table := range []string{"groups", "feeds", "items"} {
		exists, err := tableExistsDB(db, table)
		if err != nil {
			return false, fmt.Errorf("check table %s: %w", table, err)
		}
		if !exists {
			return false, nil
		}
	}

	legacyMarkers := []struct {
		table  string
		column string
	}{
		{table: "groups", column: "deleted_at"},
		{table: "feeds", column: "consecutive_failures"},
		{table: "feeds", column: "req_proxy"},
		{table: "items", column: "bookmark"},
		{table: "items", column: "deleted_at"},
	}

	for _, marker := range legacyMarkers {
		exists, err := columnExistsDB(db, marker.table, marker.column)
		if err != nil {
			return false, fmt.Errorf("check legacy column %s.%s: %w", marker.table, marker.column, err)
		}
		if exists {
			return true, nil
		}
	}

	hasMigrationsTable, err := tableExistsDB(db, "schema_migrations")
	if err != nil {
		return false, fmt.Errorf("check schema_migrations table: %w", err)
	}
	if hasMigrationsTable {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
			return false, fmt.Errorf("count schema_migrations rows: %w", err)
		}
		if count > 0 {
			return false, nil
		}
	}

	return false, nil
}

func tableExistsDB(db *sql.DB, table string) (bool, error) {
	var count int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = :table`,
		sql.Named("table", table),
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func columnExistsDB(db *sql.DB, table, column string) (bool, error) {
	var count int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM pragma_table_info(:table) WHERE name = :column`,
		sql.Named("table", table),
		sql.Named("column", column),
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
