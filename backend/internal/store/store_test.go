package store

import (
	"path/filepath"
	"testing"

	"github.com/patrickjmcd/reedme/internal/model"
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

func mustCreateGroup(t *testing.T, store *Store, name string) *model.Group {
	t.Helper()

	group, err := store.CreateGroup(name)
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err)
	}

	return group
}

func mustCreateFeed(t *testing.T, store *Store, groupID int64, name, link, siteURL, proxy string) *model.Feed {
	t.Helper()

	feed, err := store.CreateFeed(groupID, name, link, siteURL, proxy)
	if err != nil {
		t.Fatalf("CreateFeed() failed: %v", err)
	}

	return feed
}

func mustCreateItem(t *testing.T, store *Store, feedID int64, guid, title, link, content string, pubDate int64) *model.Item {
	t.Helper()

	item, err := store.CreateItem(feedID, guid, title, link, content, pubDate)
	if err != nil {
		t.Fatalf("CreateItem() failed: %v", err)
	}

	return item
}

func mustCreateBookmark(t *testing.T, store *Store, itemID *int64, link, title, content string, pubDate int64, feedName string) *model.Bookmark {
	t.Helper()

	bookmark, err := store.CreateBookmark(itemID, link, title, content, pubDate, feedName)
	if err != nil {
		t.Fatalf("CreateBookmark() failed: %v", err)
	}

	return bookmark
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
