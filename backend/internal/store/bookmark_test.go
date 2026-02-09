package store

import (
	"database/sql"
	"errors"
	"testing"
)

func TestListBookmarks(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	// Test empty list
	bookmarks, err := store.ListBookmarks(10, 0)
	if err != nil {
		t.Fatalf("ListBookmarks() failed: %v", err)
	}
	if len(bookmarks) != 0 {
		t.Errorf("expected 0 bookmarks, got %d", len(bookmarks))
	}

	pubDate := int64(123)
	b1 := mustCreateBookmark(t, store, nil, "https://example.com/1", "Bookmark 1", "Content 1", pubDate, "Feed 1")
	b2 := mustCreateBookmark(t, store, nil, "https://example.com/2", "Bookmark 2", "Content 2", pubDate, "Feed 2")
	b3 := mustCreateBookmark(t, store, nil, "https://example.com/3", "Bookmark 3", "Content 3", pubDate, "Feed 3")

	// Make created_at deterministic (avoid time.Sleep + unixepoch() 1s resolution)
	if _, err := store.db.Exec(
		`UPDATE bookmarks SET created_at = :created_at WHERE id = :id`,
		sql.Named("created_at", int64(100)),
		sql.Named("id", b1.ID),
	); err != nil {
		t.Fatalf("failed to set created_at: %v", err)
	}
	if _, err := store.db.Exec(
		`UPDATE bookmarks SET created_at = :created_at WHERE id = :id`,
		sql.Named("created_at", int64(200)),
		sql.Named("id", b2.ID),
	); err != nil {
		t.Fatalf("failed to set created_at: %v", err)
	}
	if _, err := store.db.Exec(
		`UPDATE bookmarks SET created_at = :created_at WHERE id = :id`,
		sql.Named("created_at", int64(300)),
		sql.Named("id", b3.ID),
	); err != nil {
		t.Fatalf("failed to set created_at: %v", err)
	}

	t.Run("list all bookmarks ordered by created_at DESC", func(t *testing.T) {
		bookmarks, err := store.ListBookmarks(10, 0)
		if err != nil {
			t.Fatalf("ListBookmarks() failed: %v", err)
		}

		if len(bookmarks) != 3 {
			t.Fatalf("expected 3 bookmarks, got %d", len(bookmarks))
		}

		// Should be in descending order (newest first)
		if bookmarks[0].ID != b3.ID || bookmarks[1].ID != b2.ID || bookmarks[2].ID != b1.ID {
			t.Error("bookmarks not ordered by created_at DESC")
		}
	})

	t.Run("pagination with limit", func(t *testing.T) {
		bookmarks, err := store.ListBookmarks(2, 0)
		if err != nil {
			t.Fatalf("ListBookmarks() failed: %v", err)
		}

		if len(bookmarks) != 2 {
			t.Errorf("expected 2 bookmarks with limit=2, got %d", len(bookmarks))
		}
	})

	t.Run("pagination with offset", func(t *testing.T) {
		bookmarks, err := store.ListBookmarks(10, 2)
		if err != nil {
			t.Fatalf("ListBookmarks() failed: %v", err)
		}

		if len(bookmarks) != 1 {
			t.Errorf("expected 1 bookmark with offset=2, got %d", len(bookmarks))
		}

		if bookmarks[0].ID != b1.ID {
			t.Error("incorrect bookmark returned with offset")
		}
	})

	t.Run("stable order when created_at ties", func(t *testing.T) {
		if _, err := store.db.Exec(`UPDATE bookmarks SET created_at = :created_at`, sql.Named("created_at", int64(100))); err != nil {
			t.Fatalf("failed to set created_at for tie test: %v", err)
		}

		bookmarks, err := store.ListBookmarks(10, 0)
		if err != nil {
			t.Fatalf("ListBookmarks() failed: %v", err)
		}
		if len(bookmarks) != 3 {
			t.Fatalf("expected 3 bookmarks, got %d", len(bookmarks))
		}
		if bookmarks[0].ID != b3.ID || bookmarks[1].ID != b2.ID || bookmarks[2].ID != b1.ID {
			t.Error("bookmarks not ordered by created_at DESC, id DESC")
		}
	})
}

func TestGetBookmark(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	created := mustCreateBookmark(t, store, nil, "https://example.com/test", "Test Bookmark", "Content", 123, "Test Feed")

	// Get existing bookmark
	bookmark, err := store.GetBookmark(created.ID)
	if err != nil {
		t.Fatalf("GetBookmark() failed: %v", err)
	}

	if bookmark.ID != created.ID || bookmark.Title != created.Title {
		t.Error("retrieved bookmark doesn't match created bookmark")
	}

	// Get non-existent bookmark
	_, err = store.GetBookmark(99999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for non-existent bookmark, got %v", err)
	}
}

func TestCreateBookmark(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	t.Run("create bookmark with item_id", func(t *testing.T) {
		group := mustCreateGroup(t, store, "Test Group")
		feed := mustCreateFeed(t, store, group.ID, "Test Feed", "https://example.com/feed", "https://example.com", "")
		item := mustCreateItem(t, store, feed.ID, "guid-1", "Test Item", "https://example.com/item", "Content", 123)

		bookmark := mustCreateBookmark(t, store, &item.ID, item.Link, item.Title, item.Content, item.PubDate, "Test Feed")

		if bookmark.ItemID == nil || *bookmark.ItemID != item.ID {
			t.Error("expected item_id to be set")
		}

		if bookmark.Link != item.Link || bookmark.Title != item.Title {
			t.Error("bookmark fields don't match input")
		}

		if bookmark.ID == 0 || bookmark.CreatedAt == 0 {
			t.Error("expected auto-populated fields to be set")
		}
	})

	t.Run("create bookmark with NULL item_id", func(t *testing.T) {
		bookmark := mustCreateBookmark(t, store, nil, "https://example.com/orphan", "Orphan Bookmark", "Content", 123, "Unknown Feed")

		if bookmark.ItemID != nil {
			t.Error("expected item_id to be NULL")
		}

		if bookmark.Link != "https://example.com/orphan" {
			t.Error("bookmark link doesn't match input")
		}
	})

	t.Run("unique constraint on link", func(t *testing.T) {
		link := "https://example.com/duplicate"
		mustCreateBookmark(t, store, nil, link, "Bookmark 1", "Content", 123, "Feed")

		// Try to create duplicate
		_, err := store.CreateBookmark(nil, link, "Bookmark 2", "Content", 123, "Feed")
		if err == nil {
			t.Error("expected error when creating duplicate bookmark link, got nil")
		}
	})
}

func TestDeleteBookmark(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	bookmark := mustCreateBookmark(t, store, nil, "https://example.com/test", "Test Bookmark", "Content", 123, "Test Feed")

	// Delete bookmark
	if err := store.DeleteBookmark(bookmark.ID); err != nil {
		t.Fatalf("DeleteBookmark() failed: %v", err)
	}

	// Verify deletion
	_, err := store.GetBookmark(bookmark.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after deletion, got %v", err)
	}
}

func TestBookmarkExists(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	link := "https://example.com/test"
	mustCreateBookmark(t, store, nil, link, "Test Bookmark", "Content", 123, "Test Feed")

	// Test existing bookmark
	exists, err := store.BookmarkExists(link)
	if err != nil {
		t.Fatalf("BookmarkExists() failed: %v", err)
	}

	if !exists {
		t.Error("expected bookmark to exist")
	}

	// Test non-existing bookmark
	exists, err = store.BookmarkExists("https://example.com/nonexistent")
	if err != nil {
		t.Fatalf("BookmarkExists() failed: %v", err)
	}

	if exists {
		t.Error("expected bookmark not to exist")
	}
}
