package store

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// setupTestDB creates a test database in the system temp directory
func setupTestDB(t *testing.T) (*Store, string) {
	t.Helper()

	dbPath := filepath.Join(os.TempDir(), fmt.Sprintf("fusion_test_%d.db", os.Getpid()))

	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	return store, dbPath
}

// teardownTestDB closes the database and removes the test file
func teardownTestDB(t *testing.T, store *Store, dbPath string) {
	t.Helper()

	if err := store.Close(); err != nil {
		t.Errorf("failed to close database: %v", err)
	}

	if err := os.Remove(dbPath); err != nil {
		t.Errorf("failed to remove test database: %v", err)
	}
}

func TestNew(t *testing.T) {
	dbPath := filepath.Join(os.TempDir(), fmt.Sprintf("fusion_test_new_%d.db", os.Getpid()))
	defer os.Remove(dbPath)

	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer store.Close()

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
	store, dbPath := setupTestDB(t)
	defer os.Remove(dbPath)

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
