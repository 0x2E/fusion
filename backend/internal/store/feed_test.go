package store

import (
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestListFeeds(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	// Create a group
	group := mustCreateGroup(t, store, "Test Group")

	// Test empty list
	feeds, err := store.ListFeeds()
	if err != nil {
		t.Fatalf("ListFeeds() failed: %v", err)
	}
	if len(feeds) != 0 {
		t.Errorf("expected 0 feeds, got %d", len(feeds))
	}

	// Create feeds
	f1 := mustCreateFeed(t, store, group.ID, "Feed 1", "https://example.com/feed1", "https://example.com", "")
	f2 := mustCreateFeed(t, store, group.ID, "Feed 2", "https://example.com/feed2", "https://example.com", "")

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

	group := mustCreateGroup(t, store, "Test Group")
	created := mustCreateFeed(t, store, group.ID, "Test Feed", "https://example.com/feed", "https://example.com", "")

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

	group := mustCreateGroup(t, store, "Test Group")

	name := "New Feed"
	link := "https://example.com/feed"
	siteURL := "https://example.com"
	proxy := "http://proxy.example.com"

	feed := mustCreateFeed(t, store, group.ID, name, link, siteURL, proxy)

	if feed.Name != name || feed.Link != link || feed.SiteURL != siteURL || feed.Proxy != proxy {
		t.Error("feed fields don't match input")
	}

	if feed.Suspended != false {
		t.Error("expected suspended to default to false")
	}

	if feed.FetchState.ConsecutiveFailures != 0 {
		t.Error("expected consecutive_failures to default to 0")
	}

	if feed.ID == 0 || feed.CreatedAt == 0 || feed.UpdatedAt == 0 {
		t.Error("expected auto-populated fields to be set")
	}

	// Test UNIQUE constraint on link
	_, err := store.CreateFeed(group.ID, "Duplicate", link, siteURL, "")
	if err == nil {
		t.Error("expected error when creating duplicate feed link, got nil")
	}
}

func TestUpdateFeed(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group1 := mustCreateGroup(t, store, "Group 1")
	group2 := mustCreateGroup(t, store, "Group 2")
	feed := mustCreateFeed(t, store, group1.ID, "Original Feed", "https://example.com/feed", "https://example.com", "")

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

func TestUpdateFeedLinkResetsFetchState(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group := mustCreateGroup(t, store, "Group")
	feed := mustCreateFeed(t, store, group.ID, "Feed", "https://example.com/feed.xml", "https://example.com", "")

	checkedAt := time.Now().Unix()
	if err := store.UpdateFeedFetchSuccess(feed.ID, UpdateFeedFetchSuccessParams{
		CheckedAt:    checkedAt,
		HTTPStatus:   200,
		ETag:         "etag-v1",
		LastModified: "Mon, 02 Jan 2006 15:04:05 GMT",
		NextCheckAt:  checkedAt + 3600,
	}); err != nil {
		t.Fatalf("UpdateFeedFetchSuccess() failed: %v", err)
	}

	newLink := "https://example.com/new-feed.xml"
	if err := store.UpdateFeed(feed.ID, UpdateFeedParams{Link: &newLink}); err != nil {
		t.Fatalf("UpdateFeed() failed: %v", err)
	}

	updated, err := store.GetFeed(feed.ID)
	if err != nil {
		t.Fatalf("GetFeed() failed: %v", err)
	}

	if updated.Link != newLink {
		t.Fatalf("expected link %q, got %q", newLink, updated.Link)
	}
	if updated.FetchState.ETag != "" {
		t.Fatalf("expected etag to be reset, got %q", updated.FetchState.ETag)
	}
	if updated.FetchState.LastModified != "" {
		t.Fatalf("expected last_modified to be reset, got %q", updated.FetchState.LastModified)
	}
	if updated.FetchState.ConsecutiveFailures != 0 {
		t.Fatalf("expected consecutive_failures reset to 0, got %d", updated.FetchState.ConsecutiveFailures)
	}
}

func TestDeleteFeed(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group := mustCreateGroup(t, store, "Test Group")
	feed := mustCreateFeed(t, store, group.ID, "Test Feed", "https://example.com/feed", "https://example.com", "")

	item := mustCreateItem(t, store, feed.ID, "guid-1", "Item 1", "https://example.com/item1", "Content 1", time.Now().Unix())
	bookmark := mustCreateBookmark(t, store, &item.ID, "https://example.com/item1", "Item 1", "Content 1", item.PubDate, "Test Feed")

	if err := store.DeleteFeed(feed.ID); err != nil {
		t.Fatalf("DeleteFeed() failed: %v", err)
	}

	_, err := store.GetFeed(feed.ID)
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

func TestUpdateFeedFetchSuccess(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group := mustCreateGroup(t, store, "Test Group")
	feed := mustCreateFeed(t, store, group.ID, "Test Feed", "https://example.com/feed", "https://example.com", "")

	seedCheckedAt := time.Now().Unix()
	if err := store.UpdateFeedFetchSuccess(feed.ID, UpdateFeedFetchSuccessParams{
		CheckedAt:    seedCheckedAt,
		HTTPStatus:   200,
		CacheControl: "max-age=1800",
		ExpiresAt:    seedCheckedAt + 1800,
		NextCheckAt:  seedCheckedAt + 1800,
	}); err != nil {
		t.Fatalf("UpdateFeedFetchSuccess() seed failed: %v", err)
	}

	if err := store.UpdateFeedFetchFailure(feed.ID, UpdateFeedFetchFailureParams{
		CheckedAt:       time.Now().Unix(),
		HTTPStatus:      500,
		LastError:       "test error",
		IntervalSeconds: 1800,
		MaxBackoff:      604800,
	}); err != nil {
		t.Fatalf("UpdateFeedFetchFailure() failed: %v", err)
	}

	checkedAt := time.Now().Unix()
	nextCheckAt := time.Now().Add(30 * time.Minute).Unix()
	if err := store.UpdateFeedFetchSuccess(feed.ID, UpdateFeedFetchSuccessParams{
		CheckedAt:    checkedAt,
		HTTPStatus:   200,
		ETag:         "abc",
		LastModified: "Mon, 02 Jan 2006 15:04:05 GMT",
		CacheControl: "max-age=1800",
		ExpiresAt:    checkedAt + 1800,
		NextCheckAt:  nextCheckAt,
	}); err != nil {
		t.Fatalf("UpdateFeedFetchSuccess() failed: %v", err)
	}

	updated, err := store.GetFeed(feed.ID)
	if err != nil {
		t.Fatalf("GetFeed() failed: %v", err)
	}

	if updated.FetchState.LastSuccessAt != checkedAt {
		t.Errorf("expected last_success_at %d, got %d", checkedAt, updated.FetchState.LastSuccessAt)
	}

	if updated.FetchState.LastError != "" {
		t.Error("expected last_error to be cleared")
	}

	if updated.FetchState.ConsecutiveFailures != 0 {
		t.Error("expected consecutive_failures to be cleared")
	}
	if updated.FetchState.LastErrorAt != 0 {
		t.Error("expected last_error_at to be cleared")
	}
	if updated.FetchState.ETag != "abc" {
		t.Errorf("expected etag abc, got %q", updated.FetchState.ETag)
	}
	if updated.FetchState.NextCheckAt != nextCheckAt {
		t.Errorf("expected next_check_at %d, got %d", nextCheckAt, updated.FetchState.NextCheckAt)
	}
}

func TestUpdateFeedFetchFailure(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group := mustCreateGroup(t, store, "Test Group")
	feed := mustCreateFeed(t, store, group.ID, "Test Feed", "https://example.com/feed", "https://example.com", "")

	seedCheckedAt := time.Now().Unix()
	if err := store.UpdateFeedFetchSuccess(feed.ID, UpdateFeedFetchSuccessParams{
		CheckedAt:    seedCheckedAt,
		HTTPStatus:   200,
		CacheControl: "max-age=1800",
		ExpiresAt:    seedCheckedAt + 1800,
		NextCheckAt:  seedCheckedAt + 1800,
	}); err != nil {
		t.Fatalf("UpdateFeedFetchSuccess() seed failed: %v", err)
	}

	errorMsg1 := "first error"
	firstCheckedAt := time.Now().Unix()
	if err := store.UpdateFeedFetchFailure(feed.ID, UpdateFeedFetchFailureParams{
		CheckedAt:       firstCheckedAt,
		HTTPStatus:      500,
		LastError:       errorMsg1,
		IntervalSeconds: 1800,
		MaxBackoff:      604800,
	}); err != nil {
		t.Fatalf("UpdateFeedFetchFailure() failed: %v", err)
	}

	updated1, err := store.GetFeed(feed.ID)
	if err != nil {
		t.Fatalf("GetFeed() failed: %v", err)
	}

	if updated1.FetchState.LastError != errorMsg1 {
		t.Errorf("expected last_error %q, got %q", errorMsg1, updated1.FetchState.LastError)
	}

	if updated1.FetchState.ConsecutiveFailures != 1 {
		t.Errorf("expected consecutive_failures to be 1, got %d", updated1.FetchState.ConsecutiveFailures)
	}
	if updated1.FetchState.LastErrorAt == 0 {
		t.Error("expected last_error_at to be set after failure")
	}
	if updated1.FetchState.CacheControl != "max-age=1800" {
		t.Errorf("expected cache_control to remain from last success, got %q", updated1.FetchState.CacheControl)
	}
	if updated1.FetchState.ExpiresAt != seedCheckedAt+1800 {
		t.Errorf("expected expires_at to remain from last success, got %d", updated1.FetchState.ExpiresAt)
	}
	firstFailureAt := updated1.FetchState.LastErrorAt
	firstNextCheckAt := updated1.FetchState.NextCheckAt

	errorMsg2 := "second error"
	secondCheckedAt := time.Now().Unix()
	if err := store.UpdateFeedFetchFailure(feed.ID, UpdateFeedFetchFailureParams{
		CheckedAt:       secondCheckedAt,
		HTTPStatus:      502,
		LastError:       errorMsg2,
		IntervalSeconds: 1800,
		MaxBackoff:      604800,
	}); err != nil {
		t.Fatalf("UpdateFeedFetchFailure() failed: %v", err)
	}

	updated2, err := store.GetFeed(feed.ID)
	if err != nil {
		t.Fatalf("GetFeed() failed: %v", err)
	}

	if updated2.FetchState.LastError != errorMsg2 {
		t.Errorf("expected last_error %q, got %q", errorMsg2, updated2.FetchState.LastError)
	}

	if updated2.FetchState.ConsecutiveFailures != 2 {
		t.Errorf("expected consecutive_failures to be 2, got %d", updated2.FetchState.ConsecutiveFailures)
	}
	if updated2.FetchState.LastErrorAt < firstFailureAt {
		t.Error("expected last_error_at to be monotonic")
	}
	if updated2.FetchState.CacheControl != "max-age=1800" {
		t.Errorf("expected cache_control to remain from last success, got %q", updated2.FetchState.CacheControl)
	}
	if updated2.FetchState.ExpiresAt != seedCheckedAt+1800 {
		t.Errorf("expected expires_at to remain from last success, got %d", updated2.FetchState.ExpiresAt)
	}
	if updated2.FetchState.NextCheckAt <= firstNextCheckAt {
		t.Error("expected next_check_at to increase with more failures")
	}
}

func TestUpdateFeedFetchFailureIgnoresFreshnessHeadersForScheduling(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group := mustCreateGroup(t, store, "Test Group")
	feed := mustCreateFeed(t, store, group.ID, "Test Feed", "https://example.com/feed", "https://example.com", "")

	checkedAt := int64(1700000000)
	if err := store.UpdateFeedFetchFailure(feed.ID, UpdateFeedFetchFailureParams{
		CheckedAt:       checkedAt,
		HTTPStatus:      500,
		LastError:       "upstream failure",
		RetryAfterUntil: 0,
		IntervalSeconds: 60,
		MaxBackoff:      600,
	}); err != nil {
		t.Fatalf("UpdateFeedFetchFailure() failed: %v", err)
	}

	updated, err := store.GetFeed(feed.ID)
	if err != nil {
		t.Fatalf("GetFeed() failed: %v", err)
	}

	if updated.FetchState.NextCheckAt >= checkedAt+3600 {
		t.Fatalf("expected failure scheduling to ignore freshness headers, got next_check_at=%d", updated.FetchState.NextCheckAt)
	}
}

func TestUpdateFeedFetchFailureConcurrentIncrements(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group := mustCreateGroup(t, store, "Test Group")
	feed := mustCreateFeed(t, store, group.ID, "Test Feed", "https://example.com/feed", "https://example.com", "")

	checkedAt := time.Now().Unix()
	params := UpdateFeedFetchFailureParams{
		CheckedAt:       checkedAt,
		HTTPStatus:      500,
		LastError:       "concurrent error",
		IntervalSeconds: 1800,
		MaxBackoff:      604800,
	}

	var wg sync.WaitGroup
	var err1 error
	var err2 error
	wg.Add(2)
	go func() {
		defer wg.Done()
		err1 = store.UpdateFeedFetchFailure(feed.ID, params)
	}()
	go func() {
		defer wg.Done()
		err2 = store.UpdateFeedFetchFailure(feed.ID, params)
	}()
	wg.Wait()

	if err1 != nil {
		t.Fatalf("first UpdateFeedFetchFailure() failed: %v", err1)
	}
	if err2 != nil {
		t.Fatalf("second UpdateFeedFetchFailure() failed: %v", err2)
	}

	updated, err := store.GetFeed(feed.ID)
	if err != nil {
		t.Fatalf("GetFeed() failed: %v", err)
	}

	if updated.FetchState.ConsecutiveFailures != 2 {
		t.Fatalf("expected consecutive_failures=2 after concurrent failures, got %d", updated.FetchState.ConsecutiveFailures)
	}
}

func TestUpdateFeedSiteURLIfEmpty(t *testing.T) {
	tests := []struct {
		name           string
		initialSiteURL string
	}{
		{name: "empty", initialSiteURL: ""},
		{name: "whitespace", initialSiteURL: "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, _ := setupTestDB(t)
			defer closeStore(t, store)

			group := mustCreateGroup(t, store, "Test Group")
			feed := mustCreateFeed(t, store, group.ID, "Test Feed", "https://example.com/feed", tt.initialSiteURL, "")

			firstSiteURL := "https://example.com"
			if err := store.UpdateFeedSiteURLIfEmpty(feed.ID, firstSiteURL); err != nil {
				t.Fatalf("UpdateFeedSiteURLIfEmpty() failed: %v", err)
			}

			updated, err := store.GetFeed(feed.ID)
			if err != nil {
				t.Fatalf("GetFeed() failed: %v", err)
			}
			if updated.SiteURL != firstSiteURL {
				t.Fatalf("expected site_url %q, got %q", firstSiteURL, updated.SiteURL)
			}

			secondSiteURL := "https://should-not-overwrite.example.com"
			if err := store.UpdateFeedSiteURLIfEmpty(feed.ID, secondSiteURL); err != nil {
				t.Fatalf("UpdateFeedSiteURLIfEmpty() second call failed: %v", err)
			}

			updated, err = store.GetFeed(feed.ID)
			if err != nil {
				t.Fatalf("GetFeed() failed: %v", err)
			}
			if updated.SiteURL != firstSiteURL {
				t.Fatalf("site_url should remain %q, got %q", firstSiteURL, updated.SiteURL)
			}
		})
	}
}

func TestUpdateFeedNoParamsDoesNothing(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group := mustCreateGroup(t, store, "Test Group")
	feed := mustCreateFeed(t, store, group.ID, "Original Feed", "https://example.com/feed", "https://example.com", "")

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

func TestBatchCreateFeedsHandlesExistingAndInBatchDuplicates(t *testing.T) {
	store, _ := setupTestDB(t)
	defer closeStore(t, store)

	group := mustCreateGroup(t, store, "Batch Group")
	mustCreateFeed(t, store, group.ID, "Existing", "https://example.com/existing.xml", "", "")

	result, err := store.BatchCreateFeeds([]BatchCreateFeedsInput{
		{GroupID: group.ID, Name: "Dup Existing", Link: "https://example.com/existing.xml"},
		{GroupID: group.ID, Name: "Fresh 1", Link: "https://example.com/fresh1.xml"},
		{GroupID: group.ID, Name: "Dup In Batch", Link: "https://example.com/fresh1.xml"},
		{GroupID: group.ID, Name: "Fresh 2", Link: "https://example.com/fresh2.xml"},
	})
	if err != nil {
		t.Fatalf("BatchCreateFeeds() failed: %v", err)
	}

	if result.Created != 2 {
		t.Fatalf("expected 2 created feeds, got %d", result.Created)
	}
	if len(result.CreatedIDs) != 2 {
		t.Fatalf("expected 2 created IDs, got %d", len(result.CreatedIDs))
	}
	if len(result.Errors) != 2 {
		t.Fatalf("expected 2 duplicate errors, got %d (%v)", len(result.Errors), result.Errors)
	}

	feeds, err := store.ListFeeds()
	if err != nil {
		t.Fatalf("ListFeeds() failed: %v", err)
	}
	if len(feeds) != 3 {
		t.Fatalf("expected total 3 feeds (1 existing + 2 new), got %d", len(feeds))
	}
}
