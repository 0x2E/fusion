package store

import (
	"database/sql"
	"errors"
	"testing"
)

func TestListItems(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

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

	// Create items with deterministic pub_date ordering.
	item1, err := store.CreateItem(feed1.ID, "guid-1", "Item 1", "https://example.com/1", "Content 1", 100)
	if err != nil {
		t.Fatalf("CreateItem() failed: %v", err)
	}
	item2, err := store.CreateItem(feed1.ID, "guid-2", "Item 2", "https://example.com/2", "Content 2", 200)
	if err != nil {
		t.Fatalf("CreateItem() failed: %v", err)
	}
	item3, err := store.CreateItem(feed2.ID, "guid-3", "Item 3", "https://example.com/3", "Content 3", 300)
	if err != nil {
		t.Fatalf("CreateItem() failed: %v", err)
	}

	// Mark item2 as read
	if err := store.UpdateItemUnread(item2.ID, false); err != nil {
		t.Fatalf("UpdateItemUnread() failed: %v", err)
	}

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
		if len(items) != 3 {
			t.Fatalf("expected 3 items, got %d", len(items))
		}
		if items[0].ID != item3.ID || items[1].ID != item2.ID || items[2].ID != item1.ID {
			t.Error("items not ordered by pub_date DESC")
		}
	})

	t.Run("stable order when pub_date ties", func(t *testing.T) {
		if _, err := store.db.Exec(
			`UPDATE items SET pub_date = :pub_date`,
			sql.Named("pub_date", int64(100)),
		); err != nil {
			t.Fatalf("failed to set pub_date for tie test: %v", err)
		}

		items, err := store.ListItems(ListItemsParams{OrderBy: "pub_date"})
		if err != nil {
			t.Fatalf("ListItems() failed: %v", err)
		}
		if len(items) != 3 {
			t.Fatalf("expected 3 items, got %d", len(items))
		}
		if items[0].ID != item3.ID || items[1].ID != item2.ID || items[2].ID != item1.ID {
			t.Error("items not ordered by pub_date DESC, id DESC")
		}
	})

	t.Run("order by created_at", func(t *testing.T) {
		if _, err := store.db.Exec(
			`UPDATE items SET created_at = :created_at WHERE id = :id`,
			sql.Named("created_at", int64(100)),
			sql.Named("id", item1.ID),
		); err != nil {
			t.Fatalf("failed to set created_at: %v", err)
		}
		if _, err := store.db.Exec(
			`UPDATE items SET created_at = :created_at WHERE id = :id`,
			sql.Named("created_at", int64(200)),
			sql.Named("id", item2.ID),
		); err != nil {
			t.Fatalf("failed to set created_at: %v", err)
		}
		if _, err := store.db.Exec(
			`UPDATE items SET created_at = :created_at WHERE id = :id`,
			sql.Named("created_at", int64(300)),
			sql.Named("id", item3.ID),
		); err != nil {
			t.Fatalf("failed to set created_at: %v", err)
		}

		items, err := store.ListItems(ListItemsParams{OrderBy: "created_at"})
		if err != nil {
			t.Fatalf("ListItems() failed: %v", err)
		}
		if len(items) != 3 {
			t.Fatalf("expected 3 items, got %d", len(items))
		}
		if items[0].ID != item3.ID || items[1].ID != item2.ID || items[2].ID != item1.ID {
			t.Error("items not ordered by created_at DESC")
		}
	})

	t.Run("stable order when created_at ties", func(t *testing.T) {
		if _, err := store.db.Exec(`UPDATE items SET created_at = :created_at`, sql.Named("created_at", int64(100))); err != nil {
			t.Fatalf("failed to set created_at for tie test: %v", err)
		}

		items, err := store.ListItems(ListItemsParams{OrderBy: "created_at"})
		if err != nil {
			t.Fatalf("ListItems() failed: %v", err)
		}
		if len(items) != 3 {
			t.Fatalf("expected 3 items, got %d", len(items))
		}
		if items[0].ID != item3.ID || items[1].ID != item2.ID || items[2].ID != item1.ID {
			t.Error("items not ordered by created_at DESC, id DESC")
		}
	})
}

func TestListItemsFilterByGroupID(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group1, err := store.CreateGroup("Group 1")
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err)
	}
	group2, err := store.CreateGroup("Group 2")
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err)
	}

	feed1, err := store.CreateFeed(group1.ID, "Feed 1", "https://example.com/group1", "https://example.com", "")
	if err != nil {
		t.Fatalf("CreateFeed() failed: %v", err)
	}
	feed2, err := store.CreateFeed(group2.ID, "Feed 2", "https://example.com/group2", "https://example.com", "")
	if err != nil {
		t.Fatalf("CreateFeed() failed: %v", err)
	}

	_, err = store.CreateItem(feed1.ID, "guid-1", "Item 1", "https://example.com/1", "Content 1", 100)
	if err != nil {
		t.Fatalf("CreateItem() failed: %v", err)
	}
	_, err = store.CreateItem(feed2.ID, "guid-2", "Item 2", "https://example.com/2", "Content 2", 200)
	if err != nil {
		t.Fatalf("CreateItem() failed: %v", err)
	}

	items, err := store.ListItems(ListItemsParams{GroupID: &group1.ID})
	if err != nil {
		t.Fatalf("ListItems() failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item for group1, got %d", len(items))
	}
	if items[0].FeedID != feed1.ID {
		t.Errorf("expected item to be from feed1, got feed_id=%d", items[0].FeedID)
	}
}

func TestGetItem(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group, err := store.CreateGroup("Test Group")
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err)
	}
	feed, err := store.CreateFeed(group.ID, "Test Feed", "https://example.com/feed", "https://example.com", "")
	if err != nil {
		t.Fatalf("CreateFeed() failed: %v", err)
	}

	created, err := store.CreateItem(feed.ID, "guid-1", "Test Item", "https://example.com/1", "Content", 123)
	if err != nil {
		t.Fatalf("CreateItem() failed: %v", err)
	}

	item, err := store.GetItem(created.ID)
	if err != nil {
		t.Fatalf("GetItem() failed: %v", err)
	}

	if item.ID != created.ID || item.Title != created.Title {
		t.Error("retrieved item doesn't match created item")
	}

	_, err = store.GetItem(99999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for non-existent item, got %v", err)
	}
}

func TestCreateItem(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group, err := store.CreateGroup("Test Group")
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err)
	}
	feed, err := store.CreateFeed(group.ID, "Test Feed", "https://example.com/feed", "https://example.com", "")
	if err != nil {
		t.Fatalf("CreateFeed() failed: %v", err)
	}

	guid := "unique-guid-1"
	title := "Test Item"
	link := "https://example.com/item"
	content := "Test content"
	pubDate := int64(123)

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

	_, err = store.CreateItem(feed.ID, guid, "Different Title", link, content, pubDate)
	if err == nil {
		t.Error("expected error when creating duplicate feed_id+guid, got nil")
	}
}

func TestUpdateItemUnread(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group, err := store.CreateGroup("Test Group")
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err)
	}
	feed, err := store.CreateFeed(group.ID, "Test Feed", "https://example.com/feed", "https://example.com", "")
	if err != nil {
		t.Fatalf("CreateFeed() failed: %v", err)
	}
	item, err := store.CreateItem(feed.ID, "guid-1", "Test Item", "https://example.com/1", "Content", 123)
	if err != nil {
		t.Fatalf("CreateItem() failed: %v", err)
	}

	if err := store.UpdateItemUnread(item.ID, false); err != nil {
		t.Fatalf("UpdateItemUnread() failed: %v", err)
	}

	updated, err := store.GetItem(item.ID)
	if err != nil {
		t.Fatalf("GetItem() failed: %v", err)
	}
	if updated.Unread {
		t.Error("expected unread to be false")
	}

	if err := store.UpdateItemUnread(item.ID, true); err != nil {
		t.Fatalf("UpdateItemUnread() failed: %v", err)
	}

	updated2, err := store.GetItem(item.ID)
	if err != nil {
		t.Fatalf("GetItem() failed: %v", err)
	}
	if !updated2.Unread {
		t.Error("expected unread to be true")
	}
}

func TestUpdateItemUnreadNotFound(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	err := store.UpdateItemUnread(99999, false)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestBatchUpdateItemsUnread(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group, err := store.CreateGroup("Test Group")
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err)
	}
	feed, err := store.CreateFeed(group.ID, "Test Feed", "https://example.com/feed", "https://example.com", "")
	if err != nil {
		t.Fatalf("CreateFeed() failed: %v", err)
	}

	item1, err := store.CreateItem(feed.ID, "guid-1", "Item 1", "https://example.com/1", "Content 1", 100)
	if err != nil {
		t.Fatalf("CreateItem() failed: %v", err)
	}
	item2, err := store.CreateItem(feed.ID, "guid-2", "Item 2", "https://example.com/2", "Content 2", 200)
	if err != nil {
		t.Fatalf("CreateItem() failed: %v", err)
	}
	item3, err := store.CreateItem(feed.ID, "guid-3", "Item 3", "https://example.com/3", "Content 3", 300)
	if err != nil {
		t.Fatalf("CreateItem() failed: %v", err)
	}

	ids := []int64{item1.ID, item2.ID}
	if err := store.BatchUpdateItemsUnread(ids, false); err != nil {
		t.Fatalf("BatchUpdateItemsUnread() failed: %v", err)
	}

	updated1, err := store.GetItem(item1.ID)
	if err != nil {
		t.Fatalf("GetItem() failed: %v", err)
	}
	updated2, err := store.GetItem(item2.ID)
	if err != nil {
		t.Fatalf("GetItem() failed: %v", err)
	}
	updated3, err := store.GetItem(item3.ID)
	if err != nil {
		t.Fatalf("GetItem() failed: %v", err)
	}

	if updated1.Unread || updated2.Unread {
		t.Error("expected items 1 and 2 to be read")
	}
	if !updated3.Unread {
		t.Error("expected item 3 to remain unread")
	}

	if err := store.BatchUpdateItemsUnread([]int64{}, true); err != nil {
		t.Errorf("BatchUpdateItemsUnread() with empty ids should not error: %v", err)
	}
}

func TestMarkAllAsRead(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

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

	item1, err := store.CreateItem(feed1.ID, "guid-1", "Item 1", "https://example.com/1", "Content 1", 100)
	if err != nil {
		t.Fatalf("CreateItem() failed: %v", err)
	}
	item2, err := store.CreateItem(feed1.ID, "guid-2", "Item 2", "https://example.com/2", "Content 2", 200)
	if err != nil {
		t.Fatalf("CreateItem() failed: %v", err)
	}
	item3, err := store.CreateItem(feed2.ID, "guid-3", "Item 3", "https://example.com/3", "Content 3", 300)
	if err != nil {
		t.Fatalf("CreateItem() failed: %v", err)
	}

	t.Run("mark all items in a specific feed", func(t *testing.T) {
		if err := store.MarkAllAsRead(&feed1.ID); err != nil {
			t.Fatalf("MarkAllAsRead() failed: %v", err)
		}

		updated1, err := store.GetItem(item1.ID)
		if err != nil {
			t.Fatalf("GetItem() failed: %v", err)
		}
		updated2, err := store.GetItem(item2.ID)
		if err != nil {
			t.Fatalf("GetItem() failed: %v", err)
		}
		updated3, err := store.GetItem(item3.ID)
		if err != nil {
			t.Fatalf("GetItem() failed: %v", err)
		}

		if updated1.Unread || updated2.Unread {
			t.Error("expected feed1 items to be read")
		}
		if !updated3.Unread {
			t.Error("expected feed2 items to remain unread")
		}
	})

	if err := store.UpdateItemUnread(item3.ID, true); err != nil {
		t.Fatalf("UpdateItemUnread() failed: %v", err)
	}

	t.Run("mark all items across all feeds", func(t *testing.T) {
		if err := store.MarkAllAsRead(nil); err != nil {
			t.Fatalf("MarkAllAsRead() failed: %v", err)
		}

		updated3, err := store.GetItem(item3.ID)
		if err != nil {
			t.Fatalf("GetItem() failed: %v", err)
		}
		if updated3.Unread {
			t.Error("expected all items to be read")
		}
	})
}

func TestItemExists(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group, err := store.CreateGroup("Test Group")
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err)
	}
	feed, err := store.CreateFeed(group.ID, "Test Feed", "https://example.com/feed", "https://example.com", "")
	if err != nil {
		t.Fatalf("CreateFeed() failed: %v", err)
	}

	guid := "test-guid"
	_, err = store.CreateItem(feed.ID, guid, "Test Item", "https://example.com/1", "Content", 123)
	if err != nil {
		t.Fatalf("CreateItem() failed: %v", err)
	}

	exists, err := store.ItemExists(feed.ID, guid)
	if err != nil {
		t.Fatalf("ItemExists() failed: %v", err)
	}
	if !exists {
		t.Error("expected item to exist")
	}

	exists, err = store.ItemExists(feed.ID, "nonexistent-guid")
	if err != nil {
		t.Fatalf("ItemExists() failed: %v", err)
	}
	if exists {
		t.Error("expected item not to exist")
	}
}
