package pull

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/patrickjmcd/reedme/internal/config"
	"github.com/patrickjmcd/reedme/internal/store"
)

func TestRefreshFeedPreservesValidatorsWhen304OmitHeaders(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	st, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer st.Close()

	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		if count == 1 {
			w.Header().Set("Content-Type", "application/rss+xml")
			w.Header().Set("ETag", `"etag-v1"`)
			w.Header().Set("Last-Modified", "Mon, 02 Jan 2006 15:04:05 GMT")
			w.Header().Set("Cache-Control", "max-age=86400")
			w.Header().Set("Expires", "Tue, 03 Jan 2006 15:04:05 GMT")
			_, _ = fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel><title>Demo</title><link>https://example.com</link>
<item><guid>g1</guid><title>Item</title><link>https://example.com/1</link></item>
</channel></rss>`)
			return
		}

		w.WriteHeader(http.StatusNotModified)
	}))
	defer server.Close()

	feed, err := st.CreateFeed(1, "Feed A", server.URL, "", "")
	if err != nil {
		t.Fatalf("create feed: %v", err)
	}

	p := New(st, &config.Config{
		PullInterval:      1800,
		PullTimeout:       5,
		PullConcurrency:   1,
		PullMaxBackoff:    604800,
		AllowPrivateFeeds: true,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := p.RefreshFeed(ctx, feed.ID); err != nil {
		t.Fatalf("first refresh: %v", err)
	}

	if err := p.RefreshFeed(ctx, feed.ID); err != nil {
		t.Fatalf("second refresh: %v", err)
	}

	updatedFeed, err := st.GetFeed(feed.ID)
	if err != nil {
		t.Fatalf("get feed: %v", err)
	}

	if updatedFeed.FetchState.ETag != `"etag-v1"` {
		t.Fatalf("etag = %q, want %q", updatedFeed.FetchState.ETag, `"etag-v1"`)
	}
	if updatedFeed.FetchState.LastModified != "Mon, 02 Jan 2006 15:04:05 GMT" {
		t.Fatalf("last_modified = %q, want %q", updatedFeed.FetchState.LastModified, "Mon, 02 Jan 2006 15:04:05 GMT")
	}
	if updatedFeed.FetchState.CacheControl != "max-age=86400" {
		t.Fatalf("cache_control = %q, want %q", updatedFeed.FetchState.CacheControl, "max-age=86400")
	}

	expires, err := http.ParseTime("Tue, 03 Jan 2006 15:04:05 GMT")
	if err != nil {
		t.Fatalf("parse expires: %v", err)
	}
	if updatedFeed.FetchState.ExpiresAt != expires.Unix() {
		t.Fatalf("expires_at = %d, want %d", updatedFeed.FetchState.ExpiresAt, expires.Unix())
	}
}

func TestRefreshAllWaitsForRunningJobs(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	st, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer st.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel><title>Demo</title><link>https://example.com</link>
<item><guid>%s</guid><title>Item</title><link>https://example.com%s</link></item>
</channel></rss>`, r.URL.Path, r.URL.Path)
	}))
	defer server.Close()

	if _, err := st.CreateFeed(1, "Feed A", server.URL+"/a", "", ""); err != nil {
		t.Fatalf("create feed A: %v", err)
	}
	if _, err := st.CreateFeed(1, "Feed B", server.URL+"/b", "", ""); err != nil {
		t.Fatalf("create feed B: %v", err)
	}

	p := New(st, &config.Config{
		PullInterval:      1800,
		PullTimeout:       5,
		PullConcurrency:   1,
		PullMaxBackoff:    604800,
		AllowPrivateFeeds: true,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := p.RefreshAll(ctx)
	if err != nil {
		t.Fatalf("refresh all: %v", err)
	}
	if count != 2 {
		t.Fatalf("refresh count = %d, want 2", count)
	}

	feeds, err := st.ListFeeds()
	if err != nil {
		t.Fatalf("list feeds: %v", err)
	}
	for _, feed := range feeds {
		if feed.FetchState.LastSuccessAt <= 0 {
			t.Fatalf("feed %d last_success_at = %d, want > 0", feed.ID, feed.FetchState.LastSuccessAt)
		}
	}
}
