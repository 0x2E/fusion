package store

import (
	"database/sql"
	"errors"
	"testing"
	"time"
)

func TestListFeeds(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	// Create a group
	group, err := store.CreateGroup("Test Group")
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err)
	}

	// Test empty list
	feeds, err := store.ListFeeds()
	if err != nil {
		t.Fatalf("ListFeeds() failed: %v", err)
	}
	if len(feeds) != 0 {
		t.Errorf("expected 0 feeds, got %d", len(feeds))
	}

	// Create feeds
	f1, err := store.CreateFeed(group.ID, "Feed 1", "https://example.com/feed1", "https://example.com", "")
	if err != nil {
		t.Fatalf("CreateFeed() failed: %v", err)
	}

	f2, err := store.CreateFeed(group.ID, "Feed 2", "https://example.com/feed2", "https://example.com", "")
	if err != nil {
		t.Fatalf("CreateFeed() failed: %v", err)
	}

	// List feeds
	feeds, err = store.ListFeeds()
	if err != nil {
		t.Fatalf("ListFeeds() failed: %v", err)
	}

	if len(feeds) != 2 {
		t.Fatalf("expected 2 feeds, got %d", len(feeds))
	}

	if feeds[0].ID != f1.ID || feeds[1].ID != f2.ID {
		t.Error("feed IDs don't match")
	}

	if feeds[0].Suspended != false || feeds[1].Suspended != false {
		t.Error("expected suspended to be false by default")
	}
}

func TestGetFeed(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group, err := store.CreateGroup("Test Group")
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err)
	}

	created, err := store.CreateFeed(group.ID, "Test Feed", "https://example.com/feed", "https://example.com", "")
	if err != nil {
		t.Fatalf("CreateFeed() failed: %v", err)
	}

	// Get existing feed
	feed, err := store.GetFeed(created.ID)
	if err != nil {
		t.Fatalf("GetFeed() failed: %v", err)
	}

	if feed.ID != created.ID || feed.Name != created.Name {
		t.Error("retrieved feed doesn't match created feed")
	}

	// Get non-existent feed
	_, err = store.GetFeed(99999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for non-existent feed, got %v", err)
	}
}

func TestCreateFeed(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group, err := store.CreateGroup("Test Group")
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err)
	}

	name := "New Feed"
	link := "https://example.com/feed"
	siteURL := "https://example.com"
	proxy := "http://proxy.example.com"

	feed, err := store.CreateFeed(group.ID, name, link, siteURL, proxy)
	if err != nil {
		t.Fatalf("CreateFeed() failed: %v", err)
	}

	if feed.Name != name || feed.Link != link || feed.SiteURL != siteURL || feed.Proxy != proxy {
		t.Error("feed fields don't match input")
	}

	if feed.Suspended != false {
		t.Error("expected suspended to default to false")
	}

	if feed.Failures != 0 {
		t.Error("expected failures to default to 0")
	}

	if feed.ID == 0 || feed.CreatedAt == 0 || feed.UpdatedAt == 0 {
		t.Error("expected auto-populated fields to be set")
	}

	// Test UNIQUE constraint on link
	_, err = store.CreateFeed(group.ID, "Duplicate", link, siteURL, "")
	if err == nil {
		t.Error("expected error when creating duplicate feed link, got nil")
	}
}

func TestUpdateFeed(t *testing.T) {
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

	feed, err := store.CreateFeed(group1.ID, "Original Feed", "https://example.com/feed", "https://example.com", "")
	if err != nil {
		t.Fatalf("CreateFeed() failed: %v", err)
	}

	// Update with partial params
	newName := "Updated Feed"
	newSiteURL := "https://updated.example.com"
	suspended := true

	params := UpdateFeedParams{
		Name:      &newName,
		SiteURL:   &newSiteURL,
		Suspended: &suspended,
		GroupID:   &group2.ID,
	}

	if err := store.UpdateFeed(feed.ID, params); err != nil {
		t.Fatalf("UpdateFeed() failed: %v", err)
	}

	updated, err := store.GetFeed(feed.ID)
	if err != nil {
		t.Fatalf("GetFeed() failed: %v", err)
	}

	if updated.Name != newName {
		t.Errorf("expected name %q, got %q", newName, updated.Name)
	}

	if updated.SiteURL != newSiteURL {
		t.Errorf("expected site_url %q, got %q", newSiteURL, updated.SiteURL)
	}

	if updated.Suspended != suspended {
		t.Errorf("expected suspended %v, got %v", suspended, updated.Suspended)
	}

	if updated.GroupID != group2.ID {
		t.Errorf("expected group_id %d, got %d", group2.ID, updated.GroupID)
	}

	// Test updating only one field (others should remain unchanged)
	anotherName := "Another Name"
	params2 := UpdateFeedParams{Name: &anotherName}

	if err := store.UpdateFeed(feed.ID, params2); err != nil {
		t.Fatalf("UpdateFeed() failed: %v", err)
	}

	updated2, err := store.GetFeed(feed.ID)
	if err != nil {
		t.Fatalf("GetFeed() failed: %v", err)
	}

	if updated2.Name != anotherName {
		t.Error("name was not updated")
	}

	if updated2.SiteURL != newSiteURL {
		t.Error("site_url should remain unchanged")
	}
}

func TestDeleteFeed(t *testing.T) {
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

	item, err := store.CreateItem(feed.ID, "guid-1", "Item 1", "https://example.com/item1", "Content 1", time.Now().Unix())
	if err != nil {
		t.Fatalf("CreateItem() failed: %v", err)
	}

	bookmark, err := store.CreateBookmark(&item.ID, "https://example.com/item1", "Item 1", "Content 1", item.PubDate, "Test Feed")
	if err != nil {
		t.Fatalf("CreateBookmark() failed: %v", err)
	}

	if err := store.DeleteFeed(feed.ID); err != nil {
		t.Fatalf("DeleteFeed() failed: %v", err)
	}

	_, err = store.GetFeed(feed.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after deletion, got %v", err)
	}

	_, err = store.GetItem(item.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for item after feed deletion, got %v", err)
	}

	updatedBookmark, err := store.GetBookmark(bookmark.ID)
	if err != nil {
		t.Fatalf("GetBookmark() failed: %v", err)
	}

	if updatedBookmark.ItemID != nil {
		t.Error("expected bookmark's item_id to be NULL after feed deletion")
	}
}

func TestUpdateFeedLastBuild(t *testing.T) {
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

	if err := store.UpdateFeedFailure(feed.ID, "test error"); err != nil {
		t.Fatalf("UpdateFeedFailure() failed: %v", err)
	}

	lastBuild := time.Now().Unix()
	if err := store.UpdateFeedLastBuild(feed.ID, lastBuild); err != nil {
		t.Fatalf("UpdateFeedLastBuild() failed: %v", err)
	}

	updated, err := store.GetFeed(feed.ID)
	if err != nil {
		t.Fatalf("GetFeed() failed: %v", err)
	}

	if updated.LastBuild != lastBuild {
		t.Errorf("expected last_build %d, got %d", lastBuild, updated.LastBuild)
	}

	if updated.Failure != "" {
		t.Error("expected failure to be cleared")
	}

	if updated.Failures != 0 {
		t.Error("expected failures to be cleared")
	}
}

func TestUpdateFeedFailure(t *testing.T) {
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

	errorMsg1 := "first error"
	if err := store.UpdateFeedFailure(feed.ID, errorMsg1); err != nil {
		t.Fatalf("UpdateFeedFailure() failed: %v", err)
	}

	updated1, err := store.GetFeed(feed.ID)
	if err != nil {
		t.Fatalf("GetFeed() failed: %v", err)
	}

	if updated1.Failure != errorMsg1 {
		t.Errorf("expected failure %q, got %q", errorMsg1, updated1.Failure)
	}

	if updated1.Failures != 1 {
		t.Errorf("expected failures count to be 1, got %d", updated1.Failures)
	}

	errorMsg2 := "second error"
	if err := store.UpdateFeedFailure(feed.ID, errorMsg2); err != nil {
		t.Fatalf("UpdateFeedFailure() failed: %v", err)
	}

	updated2, err := store.GetFeed(feed.ID)
	if err != nil {
		t.Fatalf("GetFeed() failed: %v", err)
	}

	if updated2.Failure != errorMsg2 {
		t.Errorf("expected failure %q, got %q", errorMsg2, updated2.Failure)
	}

	if updated2.Failures != 2 {
		t.Errorf("expected failures count to be 2, got %d", updated2.Failures)
	}
}

func TestUpdateFeedNoParamsDoesNothing(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group, err := store.CreateGroup("Test Group")
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err)
	}

	feed, err := store.CreateFeed(group.ID, "Original Feed", "https://example.com/feed", "https://example.com", "")
	if err != nil {
		t.Fatalf("CreateFeed() failed: %v", err)
	}

	if _, err := store.db.Exec(
		`UPDATE feeds SET updated_at = :updated_at WHERE id = :id`,
		sql.Named("updated_at", int64(1)),
		sql.Named("id", feed.ID),
	); err != nil {
		t.Fatalf("failed to force updated_at for test: %v", err)
	}

	if err := store.UpdateFeed(feed.ID, UpdateFeedParams{}); err != nil {
		t.Fatalf("UpdateFeed() failed: %v", err)
	}

	updated, err := store.GetFeed(feed.ID)
	if err != nil {
		t.Fatalf("GetFeed() failed: %v", err)
	}

	if updated.UpdatedAt != 1 {
		t.Errorf("expected updated_at to be unchanged, got %d", updated.UpdatedAt)
	}
}

func TestUpdateFeedNotFound(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	name := "Updated"
	err := store.UpdateFeed(99999, UpdateFeedParams{Name: &name})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
