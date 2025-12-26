package store

import (
	"testing"
	"time"
)

func TestListBookmarks(t *testing.T) {
	store, dbPath := setupTestDB(t)
	defer teardownTestDB(t, store, dbPath)

	// Test empty list
	bookmarks, err := store.ListBookmarks(10, 0)
	if err != nil {
		t.Fatalf("ListBookmarks() failed: %v", err)
	}
	if len(bookmarks) != 0 {
		t.Errorf("expected 0 bookmarks, got %d", len(bookmarks))
	}

	// Create bookmarks with different created_at times
	now := time.Now().Unix()

	b1, _ := store.CreateBookmark(nil, "https://example.com/1", "Bookmark 1", "Content 1", now-100, "Feed 1")
	time.Sleep(1100 * time.Millisecond) // Ensure different created_at (unix seconds)
	b2, _ := store.CreateBookmark(nil, "https://example.com/2", "Bookmark 2", "Content 2", now-50, "Feed 2")
	time.Sleep(1100 * time.Millisecond)
	b3, _ := store.CreateBookmark(nil, "https://example.com/3", "Bookmark 3", "Content 3", now, "Feed 3")

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
}

func TestGetBookmark(t *testing.T) {
	store, dbPath := setupTestDB(t)
	defer teardownTestDB(t, store, dbPath)

	created, err := store.CreateBookmark(nil, "https://example.com/test", "Test Bookmark", "Content", time.Now().Unix(), "Test Feed")
	if err != nil {
		t.Fatalf("CreateBookmark() failed: %v", err)
	}

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
	if err == nil {
		t.Error("expected error for non-existent bookmark, got nil")
	}
}

func TestCreateBookmark(t *testing.T) {
	store, dbPath := setupTestDB(t)
	defer teardownTestDB(t, store, dbPath)

	t.Run("create bookmark with item_id", func(t *testing.T) {
		group, _ := store.CreateGroup("Test Group")
		feed, _ := store.CreateFeed(group.ID, "Test Feed", "https://example.com/feed", "https://example.com", "")
		item, _ := store.CreateItem(feed.ID, "guid-1", "Test Item", "https://example.com/item", "Content", time.Now().Unix())

		bookmark, err := store.CreateBookmark(&item.ID, item.Link, item.Title, item.Content, item.PubDate, "Test Feed")
		if err != nil {
			t.Fatalf("CreateBookmark() failed: %v", err)
		}

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
		bookmark, err := store.CreateBookmark(nil, "https://example.com/orphan", "Orphan Bookmark", "Content", time.Now().Unix(), "Unknown Feed")
		if err != nil {
			t.Fatalf("CreateBookmark() failed: %v", err)
		}

		if bookmark.ItemID != nil {
			t.Error("expected item_id to be NULL")
		}

		if bookmark.Link != "https://example.com/orphan" {
			t.Error("bookmark link doesn't match input")
		}
	})

	t.Run("unique constraint on link", func(t *testing.T) {
		link := "https://example.com/duplicate"
		_, err := store.CreateBookmark(nil, link, "Bookmark 1", "Content", time.Now().Unix(), "Feed")
		if err != nil {
			t.Fatalf("CreateBookmark() failed: %v", err)
		}

		// Try to create duplicate
		_, err = store.CreateBookmark(nil, link, "Bookmark 2", "Content", time.Now().Unix(), "Feed")
		if err == nil {
			t.Error("expected error when creating duplicate bookmark link, got nil")
		}
	})
}

func TestDeleteBookmark(t *testing.T) {
	store, dbPath := setupTestDB(t)
	defer teardownTestDB(t, store, dbPath)

	bookmark, err := store.CreateBookmark(nil, "https://example.com/test", "Test Bookmark", "Content", time.Now().Unix(), "Test Feed")
	if err != nil {
		t.Fatalf("CreateBookmark() failed: %v", err)
	}

	// Delete bookmark
	if err := store.DeleteBookmark(bookmark.ID); err != nil {
		t.Fatalf("DeleteBookmark() failed: %v", err)
	}

	// Verify deletion
	_, err = store.GetBookmark(bookmark.ID)
	if err == nil {
		t.Error("expected error after deletion, got nil")
	}
}

func TestBookmarkExists(t *testing.T) {
	store, dbPath := setupTestDB(t)
	defer teardownTestDB(t, store, dbPath)

	link := "https://example.com/test"
	_, err := store.CreateBookmark(nil, link, "Test Bookmark", "Content", time.Now().Unix(), "Test Feed")
	if err != nil {
		t.Fatalf("CreateBookmark() failed: %v", err)
	}

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
