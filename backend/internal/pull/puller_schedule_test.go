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

	"github.com/0x2E/fusion/internal/config"
	"github.com/0x2E/fusion/internal/store"
)

func TestPullDueOnlyFetchesDueFeeds(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	st, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer st.Close()

	var hits int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel><title>Demo</title><link>https://example.com</link>
<item><guid>%s</guid><title>Item</title><link>https://example.com%s</link></item>
</channel></rss>`, r.URL.Path, r.URL.Path)
	}))
	defer server.Close()

	if _, err := st.CreateFeed(1, "Due", server.URL+"/due", "", "", nil); err != nil {
		t.Fatalf("create due feed: %v", err)
	}
	future, err := st.CreateFeed(1, "Future", server.URL+"/future", "", "", nil)
	if err != nil {
		t.Fatalf("create future feed: %v", err)
	}

	// Both feeds start due-now from CreateFeed; push the future one out.
	if err := st.UpdateFeedFetchSuccess(future.ID, store.UpdateFeedFetchSuccessParams{
		CheckedAt:   time.Now().Unix(),
		HTTPStatus:  200,
		NextCheckAt: time.Now().Unix() + 3600,
	}); err != nil {
		t.Fatalf("set future next_check_at: %v", err)
	}

	p := New(st, &config.Config{
		PullInterval:      1800,
		PullTimeout:       5,
		PullConcurrency:   4,
		PullMaxBackoff:    604800,
		AllowPrivateFeeds: true,
	})
	p.pullDue(context.Background())

	// Only the due feed is fetched; the future one sleeps through this pass.
	if got := atomic.LoadInt32(&hits); got != 1 {
		t.Fatalf("expected 1 fetch (due feed only), got %d", got)
	}
}

func TestScheduleDelayWakesOnMutation(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	st, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer st.Close()

	p := New(st, &config.Config{
		PullInterval:    1800,
		PullTimeout:     5,
		PullConcurrency: 4,
		PullMaxBackoff:  604800,
	})

	// No feeds: fall back to the global interval.
	if d := p.scheduleDelay(); d != 1800*time.Second {
		t.Fatalf("expected global interval with no feeds, got %v", d)
	}

	interval := int64(900)
	feed, err := st.CreateFeed(1, "Feed", "https://example.com/feed", "", "", &interval)
	if err != nil {
		t.Fatalf("create feed: %v", err)
	}

	// CreateFeed marks the feed due now; the delay must clamp to the floor
	// instead of sleeping into the past (or spinning).
	if d := p.scheduleDelay(); d != minScheduleDelay {
		t.Fatalf("expected clamped delay %v for due-now feed, got %v", minScheduleDelay, d)
	}

	// After the feed is scheduled ahead, the delay tracks its next_check_at.
	ahead := time.Now().Unix() + 600
	if err := st.UpdateFeedFetchSuccess(feed.ID, store.UpdateFeedFetchSuccessParams{
		CheckedAt:   time.Now().Unix(),
		HTTPStatus:  200,
		NextCheckAt: ahead,
	}); err != nil {
		t.Fatalf("set next_check_at: %v", err)
	}
	d := p.scheduleDelay()
	if d < 500*time.Second || d > 610*time.Second {
		t.Fatalf("expected delay near 600s, got %v", d)
	}

	// Wake is non-blocking: repeated calls with no reader must not deadlock.
	p.Wake()
	p.Wake()
}
