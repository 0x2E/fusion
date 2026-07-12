package store

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/0x2E/fusion/internal/model"
)

func TestListBookmarks(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	// Test empty list
	bookmarks, err := store.ListBookmarks(ListBookmarksParams{Limit: 10})
	if err != nil {
		t.Fatalf("ListBookmarks() failed: %v", err)
	}
	if len(bookmarks) != 0 {
		t.Errorf("expected 0 bookmarks, got %d", len(bookmarks))
	}

	pubDate := int64(123)
	b1 := mustCreateBookmark(t, store, nil, nil, "https://example.com/1", "Bookmark 1", "Content 1", pubDate, "Feed 1")
	b2 := mustCreateBookmark(t, store, nil, nil, "https://example.com/2", "Bookmark 2", "Content 2", pubDate, "Feed 2")
	b3 := mustCreateBookmark(t, store, nil, nil, "https://example.com/3", "Bookmark 3", "Content 3", pubDate, "Feed 3")

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
		bookmarks, err := store.ListBookmarks(ListBookmarksParams{Limit: 10})
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
		bookmarks, err := store.ListBookmarks(ListBookmarksParams{Limit: 2})
		if err != nil {
			t.Fatalf("ListBookmarks() failed: %v", err)
		}

		if len(bookmarks) != 2 {
			t.Errorf("expected 2 bookmarks with limit=2, got %d", len(bookmarks))
		}
	})

	t.Run("pagination with cursor", func(t *testing.T) {
		// First page: no cursor, returns the 2 newest (b3 @300, b2 @200).
		page1, err := store.ListBookmarks(ListBookmarksParams{Limit: 2})
		if err != nil {
			t.Fatalf("ListBookmarks() failed: %v", err)
		}
		if len(page1) != 2 {
			t.Fatalf("expected 2 bookmarks on first page, got %d", len(page1))
		}
		if page1[0].ID != b3.ID || page1[1].ID != b2.ID {
			t.Errorf("expected first page [b3, b2], got ids [%d, %d]", page1[0].ID, page1[1].ID)
		}

		// Second page: cursor from the last item of page1 (b2).
		last := page1[len(page1)-1]
		page2, err := store.ListBookmarks(ListBookmarksParams{
			Limit:           2,
			BeforeCreatedAt: &last.CreatedAt,
			BeforeID:        &last.ID,
		})
		if err != nil {
			t.Fatalf("ListBookmarks() failed: %v", err)
		}
		if len(page2) != 1 {
			t.Fatalf("expected 1 bookmark on second page, got %d", len(page2))
		}
		if page2[0].ID != b1.ID {
			t.Errorf("expected second page [b1], got id %d", page2[0].ID)
		}

		// Beyond-last page: cursor from the final item returns nothing.
		last = page2[len(page2)-1]
		page3, err := store.ListBookmarks(ListBookmarksParams{
			Limit:           2,
			BeforeCreatedAt: &last.CreatedAt,
			BeforeID:        &last.ID,
		})
		if err != nil {
			t.Fatalf("ListBookmarks() failed: %v", err)
		}
		if len(page3) != 0 {
			t.Errorf("expected empty page beyond last item, got %d", len(page3))
		}
	})

	t.Run("stable order when created_at ties", func(t *testing.T) {
		if _, err := store.db.Exec(`UPDATE bookmarks SET created_at = :created_at`, sql.Named("created_at", int64(100))); err != nil {
			t.Fatalf("failed to set created_at for tie test: %v", err)
		}

		bookmarks, err := store.ListBookmarks(ListBookmarksParams{Limit: 10})
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

func TestListBookmarksFilters(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group1 := mustCreateGroup(t, store, "Group 1")
	group2 := mustCreateGroup(t, store, "Group 2")
	feedA := mustCreateFeed(t, store, group1.ID, "Feed A", "https://example.com/a", "https://example.com", "")
	feedB := mustCreateFeed(t, store, group1.ID, "Feed B", "https://example.com/b", "https://example.com", "")
	feedC := mustCreateFeed(t, store, group2.ID, "Feed C", "https://example.com/c", "https://example.com", "")

	// Linked items carry real unread state; orphan bookmark has feed_id=0.
	itemA := mustCreateItem(t, store, feedA.ID, "guid-a", "Item A", "https://example.com/ia", "Content", 100) // unread (default)
	itemB := mustCreateItem(t, store, feedB.ID, "guid-b", "Item B", "https://example.com/ib", "Content", 100)
	if err := store.UpdateItemUnread(itemB.ID, false); err != nil {
		t.Fatalf("mark itemB read: %v", err)
	}

	mustCreateBookmark(t, store, &itemA.ID, &itemA.FeedID, itemA.Link, itemA.Title, itemA.Content, itemA.PubDate, feedA.Name)
	mustCreateBookmark(t, store, &itemB.ID, &itemB.FeedID, itemB.Link, itemB.Title, itemB.Content, itemB.PubDate, feedB.Name)
	// Bookmark linked to feedC but without an item (orphan with known feed).
	mustCreateBookmark(t, store, nil, &feedC.ID, "https://example.com/ic", "Orphan C", "Content", 100, feedC.Name)
	// Bookmark whose source feed is gone entirely (feed_id = NULL).
	mustCreateBookmark(t, store, nil, nil, "https://example.com/orphan", "Orphan", "Content", 100, "Gone Feed")

	t.Run("filter by feed_id", func(t *testing.T) {
		got, err := store.ListBookmarks(ListBookmarksParams{FeedID: &feedA.ID, Limit: 10})
		if err != nil {
			t.Fatalf("ListBookmarks() failed: %v", err)
		}
		if len(got) != 1 || got[0].FeedID == nil || *got[0].FeedID != feedA.ID {
			t.Errorf("expected only feedA bookmark, got %d", len(got))
		}

		count, err := store.CountBookmarks(ListBookmarksParams{FeedID: &feedA.ID})
		if err != nil {
			t.Fatalf("CountBookmarks() failed: %v", err)
		}
		if count != 1 {
			t.Errorf("expected count 1 for feedA, got %d", count)
		}
	})

	t.Run("filter by group_id joins feeds", func(t *testing.T) {
		got, err := store.ListBookmarks(ListBookmarksParams{GroupID: &group1.ID, Limit: 10})
		if err != nil {
			t.Fatalf("ListBookmarks() failed: %v", err)
		}
		// feedA + feedB belong to group1; feedC and orphan excluded.
		if len(got) != 2 {
			t.Errorf("expected 2 bookmarks in group1, got %d", len(got))
		}

		count, err := store.CountBookmarks(ListBookmarksParams{GroupID: &group1.ID})
		if err != nil {
			t.Fatalf("CountBookmarks() failed: %v", err)
		}
		if count != 2 {
			t.Errorf("expected count 2 for group1, got %d", count)
		}
	})

	t.Run("unread mirrors linked item", func(t *testing.T) {
		got, err := store.ListBookmarks(ListBookmarksParams{FeedID: &feedA.ID, Limit: 10})
		if err != nil {
			t.Fatalf("ListBookmarks() failed: %v", err)
		}
		if !got[0].Unread {
			t.Error("expected bookmark of unread itemA to be unread")
		}

		gotB, err := store.ListBookmarks(ListBookmarksParams{FeedID: &feedB.ID, Limit: 10})
		if err != nil {
			t.Fatalf("ListBookmarks() failed: %v", err)
		}
		if gotB[0].Unread {
			t.Error("expected bookmark of read itemB to be read")
		}
	})

	t.Run("orphan bookmark has nil feed_id and reads as read", func(t *testing.T) {
		got, err := store.ListBookmarks(ListBookmarksParams{Limit: 10})
		if err != nil {
			t.Fatalf("ListBookmarks() failed: %v", err)
		}
		var orphan *model.Bookmark
		for _, b := range got {
			if b.FeedID == nil {
				orphan = b
			}
		}
		if orphan == nil {
			t.Fatal("expected an orphan bookmark with nil feed_id")
		}
		if orphan.Unread {
			t.Error("expected orphan bookmark to read as read")
		}
	})

	t.Run("pagination applies after feed_id filter", func(t *testing.T) {
		// Fresh feed keeps the count deterministic and isolated from earlier subtests.
		pagFeed := mustCreateFeed(t, store, group1.ID, "Pag Feed", "https://example.com/pag", "https://example.com", "")
		const n = 4
		ids := make([]int64, n)
		for i := 0; i < n; i++ {
			b := mustCreateBookmark(t, store, nil, &pagFeed.ID, fmt.Sprintf("https://example.com/pag/%d", i+1), "Pag", "Content", 100, pagFeed.Name)
			ids[i] = b.ID
			// Deterministic created_at (DESC order: highest value sorts first).
			if _, err := store.db.Exec(
				`UPDATE bookmarks SET created_at = :created_at WHERE id = :id`,
				sql.Named("created_at", int64(100*(i+1))),
				sql.Named("id", b.ID),
			); err != nil {
				t.Fatalf("failed to set created_at: %v", err)
			}
		}

		total, err := store.CountBookmarks(ListBookmarksParams{FeedID: &pagFeed.ID})
		if err != nil {
			t.Fatalf("CountBookmarks() failed: %v", err)
		}
		if total != n {
			t.Fatalf("expected filtered count %d, got %d", n, total)
		}

		// Page 1: newest two -> ids[3], ids[2].
		page1, err := store.ListBookmarks(ListBookmarksParams{FeedID: &pagFeed.ID, Limit: 2})
		if err != nil {
			t.Fatalf("ListBookmarks() page1 failed: %v", err)
		}
		if len(page1) != 2 {
			t.Fatalf("expected page1 len 2, got %d", len(page1))
		}
		if page1[0].ID != ids[3] || page1[1].ID != ids[2] {
			t.Errorf("page1 order = [%d, %d], want [%d, %d]", page1[0].ID, page1[1].ID, ids[3], ids[2])
		}

		// Page 2: cursor past page1's last item -> ids[1], ids[0].
		last := page1[len(page1)-1]
		page2, err := store.ListBookmarks(ListBookmarksParams{
			FeedID:          &pagFeed.ID,
			Limit:           2,
			BeforeCreatedAt: &last.CreatedAt,
			BeforeID:        &last.ID,
		})
		if err != nil {
			t.Fatalf("ListBookmarks() page2 failed: %v", err)
		}
		if len(page2) != 2 {
			t.Fatalf("expected page2 len 2, got %d", len(page2))
		}
		if page2[0].ID != ids[1] || page2[1].ID != ids[0] {
			t.Errorf("page2 order = [%d, %d], want [%d, %d]", page2[0].ID, page2[1].ID, ids[1], ids[0])
		}

		// Past the last page returns nothing while count stays full.
		last = page2[len(page2)-1]
		page3, err := store.ListBookmarks(ListBookmarksParams{
			FeedID:          &pagFeed.ID,
			Limit:           2,
			BeforeCreatedAt: &last.CreatedAt,
			BeforeID:        &last.ID,
		})
		if err != nil {
			t.Fatalf("ListBookmarks() page3 failed: %v", err)
		}
		if len(page3) != 0 {
			t.Errorf("expected empty page3, got %d", len(page3))
		}
	})
}

func TestGetBookmark(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	created := mustCreateBookmark(t, store, nil, nil, "https://example.com/test", "Test Bookmark", "Content", 123, "Test Feed")

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

		bookmark := mustCreateBookmark(t, store, &item.ID, &item.FeedID, item.Link, item.Title, item.Content, item.PubDate, "Test Feed")

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
		bookmark := mustCreateBookmark(t, store, nil, nil, "https://example.com/orphan", "Orphan Bookmark", "Content", 123, "Unknown Feed")

		if bookmark.ItemID != nil {
			t.Error("expected item_id to be NULL")
		}

		if bookmark.Link != "https://example.com/orphan" {
			t.Error("bookmark link doesn't match input")
		}
	})

	t.Run("unique constraint on link", func(t *testing.T) {
		link := "https://example.com/duplicate"
		mustCreateBookmark(t, store, nil, nil, link, "Bookmark 1", "Content", 123, "Feed")

		// Try to create duplicate
		_, err := store.CreateBookmark(nil, nil, link, "Bookmark 2", "Content", 123, "Feed")
		if err == nil {
			t.Error("expected error when creating duplicate bookmark link, got nil")
		}
	})
}

func TestDeleteBookmark(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	bookmark := mustCreateBookmark(t, store, nil, nil, "https://example.com/test", "Test Bookmark", "Content", 123, "Test Feed")

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
	mustCreateBookmark(t, store, nil, nil, link, "Test Bookmark", "Content", 123, "Test Feed")

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
