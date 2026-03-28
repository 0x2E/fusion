package crawl

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/0x2E/fusion/internal/model"
	"github.com/0x2E/fusion/internal/store"
)

func setupTestDB(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func mustCreateFeed(t *testing.T, st *store.Store, crawl bool) *model.Feed {
	t.Helper()
	group, err := st.CreateGroup("Group")
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}
	feed, err := st.CreateFeed(group.ID, "Feed", "https://example.com/feed.xml", "", "", crawl)
	if err != nil {
		t.Fatalf("CreateFeed: %v", err)
	}
	return feed
}

func mustCreateItem(t *testing.T, st *store.Store, feedID int64, guid, link string) *model.Item {
	t.Helper()
	item, err := st.CreateItem(feedID, guid, "Title", link, "original content", 0)
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}
	return item
}

// TestCrawlFeedItemsSkipsWhenCrawlDisabled verifies the crawler is a no-op when feed.Crawl=false.
func TestCrawlFeedItemsSkipsWhenCrawlDisabled(t *testing.T) {
	st := setupTestDB(t)

	var serverHit bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverHit = true
		fmt.Fprint(w, "<html><body><p>Article content</p></body></html>")
	}))
	defer server.Close()

	feed := mustCreateFeed(t, st, false) // crawl=false
	mustCreateItem(t, st, feed.ID, "guid-1", server.URL+"/article")

	c := New(st, 5*time.Second, true)
	if err := c.CrawlFeedItems(context.Background(), feed); err != nil {
		t.Fatalf("CrawlFeedItems() error: %v", err)
	}

	if serverHit {
		t.Error("expected no HTTP requests when crawl=false")
	}
}

// TestCrawlFeedItemsFetchesContent verifies that crawlable items get their content replaced.
func TestCrawlFeedItemsFetchesContent(t *testing.T) {
	st := setupTestDB(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<html><body><article><p>Full article text here.</p></article></body></html>`)
	}))
	defer server.Close()

	feed := mustCreateFeed(t, st, true)
	item := mustCreateItem(t, st, feed.ID, "guid-1", server.URL+"/article")

	c := New(st, 5*time.Second, true)
	if err := c.CrawlFeedItems(context.Background(), feed); err != nil {
		t.Fatalf("CrawlFeedItems() error: %v", err)
	}

	retrieved, err := st.GetItem(item.ID)
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	if retrieved.Content == "original content" {
		t.Error("expected content to be replaced with crawled content")
	}
	if retrieved.Content == "" {
		t.Error("expected non-empty crawled content")
	}

	// Item should now be marked as crawled (not returned by ListUncrawledItems).
	uncrawled, err := st.ListUncrawledItems(feed.ID, 10)
	if err != nil {
		t.Fatalf("ListUncrawledItems: %v", err)
	}
	if len(uncrawled) != 0 {
		t.Errorf("expected 0 uncrawled items after successful crawl, got %d", len(uncrawled))
	}
}

// TestCrawlFeedItemsMarksAttemptedOnFailure verifies that a failed crawl still sets crawled_at
// so the item is not retried indefinitely.
func TestCrawlFeedItemsMarksAttemptedOnFailure(t *testing.T) {
	st := setupTestDB(t)

	// Server always returns 500.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	feed := mustCreateFeed(t, st, true)
	item := mustCreateItem(t, st, feed.ID, "guid-1", server.URL+"/article")

	c := New(st, 5*time.Second, true)
	if err := c.CrawlFeedItems(context.Background(), feed); err != nil {
		t.Fatalf("CrawlFeedItems() error: %v", err)
	}

	// Original content preserved on failure.
	retrieved, err := st.GetItem(item.ID)
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	if retrieved.Content != "original content" {
		t.Errorf("expected original content preserved on failure, got %q", retrieved.Content)
	}

	// Item must not be returned as uncrawled (crawled_at was set).
	uncrawled, err := st.ListUncrawledItems(feed.ID, 10)
	if err != nil {
		t.Fatalf("ListUncrawledItems: %v", err)
	}
	if len(uncrawled) != 0 {
		t.Errorf("expected 0 uncrawled items after failed crawl attempt, got %d", len(uncrawled))
	}
}

// TestCrawlFeedItemsSkipsItemsWithoutLink verifies items with empty links are not crawled.
func TestCrawlFeedItemsSkipsItemsWithoutLink(t *testing.T) {
	st := setupTestDB(t)

	var serverHit bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverHit = true
	}))
	defer server.Close()

	feed := mustCreateFeed(t, st, true)
	mustCreateItem(t, st, feed.ID, "guid-no-link", "") // empty link

	c := New(st, 5*time.Second, true)
	if err := c.CrawlFeedItems(context.Background(), feed); err != nil {
		t.Fatalf("CrawlFeedItems() error: %v", err)
	}

	if serverHit {
		t.Error("expected no HTTP requests for items with empty links")
	}
}

// TestCrawlFeedItemsRespectsLimit verifies that at most maxItemsPerRun items are crawled per call.
func TestCrawlFeedItemsRespectsLimit(t *testing.T) {
	st := setupTestDB(t)

	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		fmt.Fprint(w, `<html><body><p>content</p></body></html>`)
	}))
	defer server.Close()

	feed := mustCreateFeed(t, st, true)
	for i := range maxItemsPerRun + 5 {
		mustCreateItem(t, st, feed.ID, fmt.Sprintf("guid-%d", i), fmt.Sprintf("%s/article/%d", server.URL, i))
	}

	c := New(st, 5*time.Second, true)
	if err := c.CrawlFeedItems(context.Background(), feed); err != nil {
		t.Fatalf("CrawlFeedItems() error: %v", err)
	}

	if requestCount > maxItemsPerRun {
		t.Errorf("expected at most %d requests, got %d", maxItemsPerRun, requestCount)
	}
}
