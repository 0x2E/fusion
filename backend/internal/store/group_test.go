package store

import (
	"testing"
	"time"
)

func TestListGroups(t *testing.T) {
	store, dbPath := setupTestDB(t)
	defer teardownTestDB(t, store, dbPath)

	// Test database with default group (id=1 created by migration)
	groups, err := store.ListGroups()
	if err != nil {
		t.Fatalf("ListGroups() failed: %v", err)
	}
	if len(groups) != 1 {
		t.Errorf("expected 1 default group, got %d", len(groups))
	}

	// Create test groups
	g1, err := store.CreateGroup("Group 1")
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err)
	}

	g2, err := store.CreateGroup("Group 2")
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err)
	}

	// List again
	groups, err = store.ListGroups()
	if err != nil {
		t.Fatalf("ListGroups() failed: %v", err)
	}

	if len(groups) != 3 {
		t.Fatalf("expected 3 groups (1 default + 2 created), got %d", len(groups))
	}

	// Verify created IDs are in the list (skip index 0 which is default group)
	if groups[1].ID != g1.ID || groups[2].ID != g2.ID {
		t.Error("group IDs don't match")
	}
}

func TestGetGroup(t *testing.T) {
	store, dbPath := setupTestDB(t)
	defer teardownTestDB(t, store, dbPath)

	// Create a group
	created, err := store.CreateGroup("Test Group")
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err)
	}

	// Get existing group
	group, err := store.GetGroup(created.ID)
	if err != nil {
		t.Fatalf("GetGroup() failed: %v", err)
	}

	if group.ID != created.ID || group.Name != created.Name {
		t.Error("retrieved group doesn't match created group")
	}

	// Get non-existent group
	_, err = store.GetGroup(99999)
	if err == nil {
		t.Error("expected error for non-existent group, got nil")
	}
}

func TestCreateGroup(t *testing.T) {
	store, dbPath := setupTestDB(t)
	defer teardownTestDB(t, store, dbPath)

	name := "New Group"
	group, err := store.CreateGroup(name)
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err)
	}

	if group.Name != name {
		t.Errorf("expected name %q, got %q", name, group.Name)
	}

	if group.ID == 0 {
		t.Error("expected non-zero ID")
	}

	if group.CreatedAt == 0 {
		t.Error("expected non-zero CreatedAt")
	}

	if group.UpdatedAt == 0 {
		t.Error("expected non-zero UpdatedAt")
	}

	// Test UNIQUE constraint
	_, err = store.CreateGroup(name)
	if err == nil {
		t.Error("expected error when creating duplicate group name, got nil")
	}
}

func TestUpdateGroup(t *testing.T) {
	store, dbPath := setupTestDB(t)
	defer teardownTestDB(t, store, dbPath)

	// Create a group
	group, err := store.CreateGroup("Original Name")
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err)
	}

	originalUpdatedAt := group.UpdatedAt

	// Wait a bit to ensure updated_at will be different (unixepoch() has 1-second resolution)
	time.Sleep(1100 * time.Millisecond)

	// Update the group
	newName := "Updated Name"
	if err := store.UpdateGroup(group.ID, newName); err != nil {
		t.Fatalf("UpdateGroup() failed: %v", err)
	}

	// Verify update
	updated, err := store.GetGroup(group.ID)
	if err != nil {
		t.Fatalf("GetGroup() failed: %v", err)
	}

	if updated.Name != newName {
		t.Errorf("expected name %q, got %q", newName, updated.Name)
	}

	if updated.UpdatedAt <= originalUpdatedAt {
		t.Errorf("expected UpdatedAt to be updated: original=%d, updated=%d", originalUpdatedAt, updated.UpdatedAt)
	}
}

func TestDeleteGroup(t *testing.T) {
	store, dbPath := setupTestDB(t)
	defer teardownTestDB(t, store, dbPath)

	t.Run("delete normal group", func(t *testing.T) {
		group, err := store.CreateGroup("Test Group")
		if err != nil {
			t.Fatalf("CreateGroup() failed: %v", err)
		}

		if err := store.DeleteGroup(group.ID); err != nil {
			t.Fatalf("DeleteGroup() failed: %v", err)
		}

		// Verify deletion
		_, err = store.GetGroup(group.ID)
		if err == nil {
			t.Error("expected error after deletion, got nil")
		}
	})

	t.Run("cannot delete default group", func(t *testing.T) {
		err := store.DeleteGroup(1)
		if err == nil {
			t.Error("expected error when deleting default group (id=1), got nil")
		}
	})

	t.Run("cascade feeds to default group", func(t *testing.T) {
		// Create a group
		group, err := store.CreateGroup("Group with Feeds")
		if err != nil {
			t.Fatalf("CreateGroup() failed: %v", err)
		}

		// Create a feed in this group
		feed, err := store.CreateFeed(group.ID, "Test Feed", "https://example.com/feed", "https://example.com", "")
		if err != nil {
			t.Fatalf("CreateFeed() failed: %v", err)
		}

		// Delete the group
		if err := store.DeleteGroup(group.ID); err != nil {
			t.Fatalf("DeleteGroup() failed: %v", err)
		}

		// Verify feed was moved to default group (id=1)
		updatedFeed, err := store.GetFeed(feed.ID)
		if err != nil {
			t.Fatalf("GetFeed() failed: %v", err)
		}

		if updatedFeed.GroupID != 1 {
			t.Errorf("expected feed to be in default group (id=1), got group_id=%d", updatedFeed.GroupID)
		}
	})
}
