package store

import (
	"path/filepath"
	"testing"
)

func closeStore(t *testing.T, store *Store) {
	t.Helper()
	if err := store.Close(); err != nil {
		t.Errorf("failed to close database: %v", err)
	}
}

func setupTestDB(t *testing.T) (*Store, string) {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	return store, dbPath
}

func TestNew(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer closeStore(t, store)

	// Verify database connection is alive
	if err := store.db.Ping(); err != nil {
		t.Errorf("database ping failed: %v", err)
	}

	// Verify migrations were executed by checking schema_migrations table exists
	var count int
	err = store.db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	if err != nil {
		t.Errorf("schema_migrations table not found: %v", err)
	}

	if count == 0 {
		t.Error("no migrations were applied")
	}
}

func TestClose(t *testing.T) {
	store, _ := setupTestDB(t)

	if err := store.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	// Verify connection is closed by attempting to ping
	if err := store.db.Ping(); err == nil {
		t.Error("expected ping to fail after Close(), but it succeeded")
	}
}

func TestNewInvalidPath(t *testing.T) {
	// Use an invalid path that should fail
	invalidPath := "/nonexistent/directory/test.db"

	_, err := New(invalidPath)
	if err == nil {
		t.Error("expected New() to fail with invalid path, but it succeeded")
	}
}
