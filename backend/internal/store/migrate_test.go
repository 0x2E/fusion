package store

import (
	"database/sql"
	"testing"
)

func TestMigrate(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	// Verify all expected tables exist
	tables := []string{"groups", "feeds", "items", "bookmarks", "schema_migrations"}
	for _, table := range tables {
		var count int
		query := "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
		err := store.db.QueryRow(query, table).Scan(&count)
		if err != nil {
			t.Errorf("failed to check table %s: %v", table, err)
			continue
		}
		if count != 1 {
			t.Errorf("expected table %s to exist, but it doesn't", table)
		}
	}
}

func TestMigrationVersionTracking(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	// Check that schema_migrations table has entries
	var count int
	err := store.db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query schema_migrations: %v", err)
	}

	if count == 0 {
		t.Error("schema_migrations table is empty, but migrations should have been applied")
	}

	// Verify all expected versions were applied
	versions := []int{1, 2, 3}
	for _, version := range versions {
		var applied bool
		err = store.db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = :version)", sql.Named("version", version)).Scan(&applied)
		if err != nil {
			t.Fatalf("failed to check version %d: %v", version, err)
		}
		if !applied {
			t.Errorf("migration version %d was not applied", version)
		}
	}
}

func TestMigrateIdempotent(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	// Get initial migration count
	var initialCount int
	err := store.db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&initialCount)
	if err != nil {
		t.Fatalf("failed to query initial migration count: %v", err)
	}

	// Run migrate again
	if err := store.migrate(); err != nil {
		t.Fatalf("second migrate() call failed: %v", err)
	}

	// Verify count hasn't changed
	var finalCount int
	err = store.db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&finalCount)
	if err != nil {
		t.Fatalf("failed to query final migration count: %v", err)
	}

	if finalCount != initialCount {
		t.Errorf("migrations were re-applied: initial=%d, final=%d", initialCount, finalCount)
	}
}
