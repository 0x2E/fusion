package store

import (
	"testing"
	"time"
)

func TestListItems(t *testing.T) {
	store, dbPath := setupTestDB(t)
	defer teardownTestDB(t, store, dbPath)

	group, err := store.CreateGroup("Test Group")
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err)
	}

	feed1, err := store.CreateFeed(group.ID, "Feed 1", "https://example.com/feed1", "https://example.com", "")
	if err != nil {
		t.Fatalf("CreateFeed() failed: %v", err)
	}

	feed2, err := store.CreateFeed(group.ID, "Feed 2", "https://example.com/feed2", "https://example.com", "")
	if err != nil {
		t.Fatalf("CreateFeed() failed: %v", err)
	}

	now := time.Now().Unix()

	// Create items
	item1, _ := store.CreateItem(feed1.ID, "guid-1", "Item 1", "https://example.com/1", "Content 1", now-100)
	item2, _ := store.CreateItem(feed1.ID, "guid-2", "Item 2", "https://example.com/2", "Content 2", now-50)
	item3, _ := store.CreateItem(feed2.ID, "guid-3", "Item 3", "https://example.com/3", "Content 3", now)

	// Mark item2 as read
	store.UpdateItemUnread(item2.ID, false)

	t.Run("list all items", func(t *testing.T) {
		items, err := store.ListItems(ListItemsParams{})
		if err != nil {
			t.Fatalf("ListItems() failed: %v", err)
		}

		if len(items) != 3 {
			t.Errorf("expected 3 items, got %d", len(items))
		}
	})

	t.Run("filter by feed_id", func(t *testing.T) {
		items, err := store.ListItems(ListItemsParams{FeedID: &feed1.ID})
		if err != nil {
			t.Fatalf("ListItems() failed: %v", err)
		}

		if len(items) != 2 {
			t.Errorf("expected 2 items for feed1, got %d", len(items))
		}
	})

	t.Run("filter by unread=true", func(t *testing.T) {
		unread := true
		items, err := store.ListItems(ListItemsParams{Unread: &unread})
		if err != nil {
			t.Fatalf("ListItems() failed: %v", err)
		}

		if len(items) != 2 {
			t.Errorf("expected 2 unread items, got %d", len(items))
		}
	})

	t.Run("filter by unread=false", func(t *testing.T) {
		unread := false
		items, err := store.ListItems(ListItemsParams{Unread: &unread})
		if err != nil {
			t.Fatalf("ListItems() failed: %v", err)
		}

		if len(items) != 1 {
			t.Errorf("expected 1 read item, got %d", len(items))
		}
	})

	t.Run("pagination with limit and offset", func(t *testing.T) {
		items, err := store.ListItems(ListItemsParams{Limit: 2, Offset: 0})
		if err != nil {
			t.Fatalf("ListItems() failed: %v", err)
		}

		if len(items) != 2 {
			t.Errorf("expected 2 items with limit=2, got %d", len(items))
		}

		items2, err := store.ListItems(ListItemsParams{Limit: 2, Offset: 2})
		if err != nil {
			t.Fatalf("ListItems() failed: %v", err)
		}

		if len(items2) != 1 {
			t.Errorf("expected 1 item with offset=2, got %d", len(items2))
		}
	})

	t.Run("order by pub_date", func(t *testing.T) {
		items, err := store.ListItems(ListItemsParams{OrderBy: "pub_date"})
		if err != nil {
			t.Fatalf("ListItems() failed: %v", err)
		}

		// Should be in descending order (newest first)
		if items[0].ID != item3.ID || items[1].ID != item2.ID || items[2].ID != item1.ID {
			t.Error("items not ordered by pub_date DESC")
		}
	})

	t.Run("order by created_at", func(t *testing.T) {
		items, err := store.ListItems(ListItemsParams{OrderBy: "created_at"})
		if err != nil {
			t.Fatalf("ListItems() failed: %v", err)
		}

		// Should be in descending order
		// FIX check order
		if len(items) != 3 {
			t.Error("expected 3 items")
		}
	})
}

func TestGetItem(t *testing.T) {
	store, dbPath := setupTestDB(t)
	defer teardownTestDB(t, store, dbPath)

	group, _ := store.CreateGroup("Test Group")
	feed, _ := store.CreateFeed(group.ID, "Test Feed", "https://example.com/feed", "https://example.com", "")

	created, err := store.CreateItem(feed.ID, "guid-1", "Test Item", "https://example.com/1", "Content", time.Now().Unix())
	if err != nil {
		t.Fatalf("CreateItem() failed: %v", err)
	}

	// Get existing item
	item, err := store.GetItem(created.ID)
	if err != nil {
		t.Fatalf("GetItem() failed: %v", err)
	}

	if item.ID != created.ID || item.Title != created.Title {
		t.Error("retrieved item doesn't match created item")
	}

	// Get non-existent item
	_, err = store.GetItem(99999)
	if err == nil {
		t.Error("expected error for non-existent item, got nil")
	}
}

func TestCreateItem(t *testing.T) {
	store, dbPath := setupTestDB(t)
	defer teardownTestDB(t, store, dbPath)

	group, _ := store.CreateGroup("Test Group")
	feed, _ := store.CreateFeed(group.ID, "Test Feed", "https://example.com/feed", "https://example.com", "")

	guid := "unique-guid-1"
	title := "Test Item"
	link := "https://example.com/item"
	content := "Test content"
	pubDate := time.Now().Unix()

	item, err := store.CreateItem(feed.ID, guid, title, link, content, pubDate)
	if err != nil {
		t.Fatalf("CreateItem() failed: %v", err)
	}

	if item.GUID != guid || item.Title != title || item.Link != link || item.Content != content {
		t.Error("item fields don't match input")
	}

	if !item.Unread {
		t.Error("expected unread to default to true")
	}

	if item.ID == 0 || item.CreatedAt == 0 {
		t.Error("expected auto-populated fields to be set")
	}

	// Test unique constraint on feed_id + guid
	_, err = store.CreateItem(feed.ID, guid, "Different Title", link, content, pubDate)
	if err == nil {
		t.Error("expected error when creating duplicate feed_id+guid, got nil")
	}
}

func TestUpdateItemUnread(t *testing.T) {
	store, dbPath := setupTestDB(t)
	defer teardownTestDB(t, store, dbPath)

	group, _ := store.CreateGroup("Test Group")
	feed, _ := store.CreateFeed(group.ID, "Test Feed", "https://example.com/feed", "https://example.com", "")
	item, _ := store.CreateItem(feed.ID, "guid-1", "Test Item", "https://example.com/1", "Content", time.Now().Unix())

	// Mark as read
	if err := store.UpdateItemUnread(item.ID, false); err != nil {
		t.Fatalf("UpdateItemUnread() failed: %v", err)
	}

	updated, _ := store.GetItem(item.ID)
	if updated.Unread {
		t.Error("expected unread to be false")
	}

	// Mark as unread
	if err := store.UpdateItemUnread(item.ID, true); err != nil {
		t.Fatalf("UpdateItemUnread() failed: %v", err)
	}

	updated2, _ := store.GetItem(item.ID)
	if !updated2.Unread {
		t.Error("expected unread to be true")
	}
}

func TestBatchUpdateItemsUnread(t *testing.T) {
	store, dbPath := setupTestDB(t)
	defer teardownTestDB(t, store, dbPath)

	group, _ := store.CreateGroup("Test Group")
	feed, _ := store.CreateFeed(group.ID, "Test Feed", "https://example.com/feed", "https://example.com", "")

	item1, _ := store.CreateItem(feed.ID, "guid-1", "Item 1", "https://example.com/1", "Content 1", time.Now().Unix())
	item2, _ := store.CreateItem(feed.ID, "guid-2", "Item 2", "https://example.com/2", "Content 2", time.Now().Unix())
	item3, _ := store.CreateItem(feed.ID, "guid-3", "Item 3", "https://example.com/3", "Content 3", time.Now().Unix())

	// Batch mark as read
	ids := []int64{item1.ID, item2.ID}
	if err := store.BatchUpdateItemsUnread(ids, false); err != nil {
		t.Fatalf("BatchUpdateItemsUnread() failed: %v", err)
	}

	// Verify
	updated1, _ := store.GetItem(item1.ID)
	updated2, _ := store.GetItem(item2.ID)
	updated3, _ := store.GetItem(item3.ID)

	if updated1.Unread || updated2.Unread {
		t.Error("expected items 1 and 2 to be read")
	}

	if !updated3.Unread {
		t.Error("expected item 3 to remain unread")
	}

	// Test empty ids list
	if err := store.BatchUpdateItemsUnread([]int64{}, true); err != nil {
		t.Errorf("BatchUpdateItemsUnread() with empty ids should not error: %v", err)
	}
}

func TestMarkAllAsRead(t *testing.T) {
	store, dbPath := setupTestDB(t)
	defer teardownTestDB(t, store, dbPath)

	group, _ := store.CreateGroup("Test Group")
	feed1, _ := store.CreateFeed(group.ID, "Feed 1", "https://example.com/feed1", "https://example.com", "")
	feed2, _ := store.CreateFeed(group.ID, "Feed 2", "https://example.com/feed2", "https://example.com", "")

	item1, _ := store.CreateItem(feed1.ID, "guid-1", "Item 1", "https://example.com/1", "Content 1", time.Now().Unix())
	item2, _ := store.CreateItem(feed1.ID, "guid-2", "Item 2", "https://example.com/2", "Content 2", time.Now().Unix())
	item3, _ := store.CreateItem(feed2.ID, "guid-3", "Item 3", "https://example.com/3", "Content 3", time.Now().Unix())

	t.Run("mark all items in a specific feed", func(t *testing.T) {
		if err := store.MarkAllAsRead(&feed1.ID); err != nil {
			t.Fatalf("MarkAllAsRead() failed: %v", err)
		}

		updated1, _ := store.GetItem(item1.ID)
		updated2, _ := store.GetItem(item2.ID)
		updated3, _ := store.GetItem(item3.ID)

		if updated1.Unread || updated2.Unread {
			t.Error("expected feed1 items to be read")
		}

		if !updated3.Unread {
			t.Error("expected feed2 items to remain unread")
		}
	})

	// Reset item3 to unread for next test
	store.UpdateItemUnread(item3.ID, true)

	t.Run("mark all items across all feeds", func(t *testing.T) {
		if err := store.MarkAllAsRead(nil); err != nil {
			t.Fatalf("MarkAllAsRead() failed: %v", err)
		}

		updated3, _ := store.GetItem(item3.ID)
		if updated3.Unread {
			t.Error("expected all items to be read")
		}
	})
}

func TestDeleteItem(t *testing.T) {
	store, dbPath := setupTestDB(t)
	defer teardownTestDB(t, store, dbPath)

	group, _ := store.CreateGroup("Test Group")
	feed, _ := store.CreateFeed(group.ID, "Test Feed", "https://example.com/feed", "https://example.com", "")
	item, _ := store.CreateItem(feed.ID, "guid-1", "Test Item", "https://example.com/1", "Content", time.Now().Unix())

	// Create bookmark with item_id
	bookmark, err := store.CreateBookmark(&item.ID, "https://example.com/1", "Test Item", "Content", item.PubDate, "Test Feed")
	if err != nil {
		t.Fatalf("CreateBookmark() failed: %v", err)
	}

	// Delete item
	if err := store.DeleteItem(item.ID); err != nil {
		t.Fatalf("DeleteItem() failed: %v", err)
	}

	// Verify item is deleted
	_, err = store.GetItem(item.ID)
	if err == nil {
		t.Error("expected error after deletion, got nil")
	}

	// Verify bookmark's item_id is set to NULL
	updatedBookmark, err := store.GetBookmark(bookmark.ID)
	if err != nil {
		t.Fatalf("GetBookmark() failed: %v", err)
	}

	if updatedBookmark.ItemID != nil {
		t.Error("expected bookmark's item_id to be NULL after item deletion")
	}
}

func TestItemExists(t *testing.T) {
	store, dbPath := setupTestDB(t)
	defer teardownTestDB(t, store, dbPath)

	group, _ := store.CreateGroup("Test Group")
	feed, _ := store.CreateFeed(group.ID, "Test Feed", "https://example.com/feed", "https://example.com", "")

	guid := "test-guid"
	_, err := store.CreateItem(feed.ID, guid, "Test Item", "https://example.com/1", "Content", time.Now().Unix())
	if err != nil {
		t.Fatalf("CreateItem() failed: %v", err)
	}

	// Test existing item
	exists, err := store.ItemExists(feed.ID, guid)
	if err != nil {
		t.Fatalf("ItemExists() failed: %v", err)
	}

	if !exists {
		t.Error("expected item to exist")
	}

	// Test non-existing item
	exists, err = store.ItemExists(feed.ID, "nonexistent-guid")
	if err != nil {
		t.Fatalf("ItemExists() failed: %v", err)
	}

	if exists {
		t.Error("expected item not to exist")
	}
}
