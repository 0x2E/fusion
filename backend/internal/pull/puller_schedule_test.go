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

func TestScheduleDelayBounds(t *testing.T) {
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
	if d := p.scheduleDelay(false); d != 1800*time.Second {
		t.Fatalf("expected global interval with no feeds, got %v", d)
	}

	interval := int64(900)
	feed, err := st.CreateFeed(1, "Feed", "https://example.com/feed", "", "", &interval)
	if err != nil {
		t.Fatalf("create feed: %v", err)
	}

	// CreateFeed marks the feed due now; the delay clamps to the floor
	// instead of sleeping into the past (or spinning).
	if d := p.scheduleDelay(false); d != minScheduleDelay {
		t.Fatalf("expected clamped delay %v for due-now feed, got %v", minScheduleDelay, d)
	}

	// A due-now feed right after a dispatched pass means the fetch-state
	// persist failed; back off to the global interval instead of 1s retries.
	if d := p.scheduleDelay(true); d != 1800*time.Second {
		t.Fatalf("expected persist-failure backoff to global interval, got %v", d)
	}

	// The delay tracks the next_check_at within one global interval...
	near := time.Now().Unix() + 600
	if err := st.UpdateFeedFetchSuccess(feed.ID, store.UpdateFeedFetchSuccessParams{
		CheckedAt:   time.Now().Unix(),
		HTTPStatus:  200,
		NextCheckAt: near,
	}); err != nil {
		t.Fatalf("set next_check_at: %v", err)
	}
	if d := p.scheduleDelay(false); d < 500*time.Second || d > 610*time.Second {
		t.Fatalf("expected delay near 600s, got %v", d)
	}

	// ...and never exceeds it, however far out the feed is scheduled.
	far := time.Now().Unix() + 86400
	if err := st.UpdateFeedFetchSuccess(feed.ID, store.UpdateFeedFetchSuccessParams{
		CheckedAt:   time.Now().Unix(),
		HTTPStatus:  200,
		NextCheckAt: far,
	}); err != nil {
		t.Fatalf("set far next_check_at: %v", err)
	}
	if d := p.scheduleDelay(false); d != 1800*time.Second {
		t.Fatalf("expected delay capped at global interval, got %v", d)
	}

	// Wake is non-blocking: repeated calls with no reader must not deadlock.
	p.Wake()
	p.Wake()
}

// TestStartSleepsUntilDueAndWakes exercises the real scheduler loop: it must
// wake at the feed's next_check_at (not at the global interval) and be
// interruptible by Wake while sleeping.
func TestStartSleepsUntilDueAndWakes(t *testing.T) {
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
		_, _ = fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel><title>Demo</title><link>https://example.com</link>
<item><guid>g1</guid><title>Item</title><link>https://example.com/1</link></item>
</channel></rss>`)
	}))
	defer server.Close()

	feed, err := st.CreateFeed(1, "Feed", server.URL+"/feed", "", "", nil)
	if err != nil {
		t.Fatalf("create feed: %v", err)
	}

	// Due in 2s — far below the 600s global interval, so a fixed-interval
	// scheduler would miss it.
	dueAt := time.Now().Unix() + 2
	if err := st.UpdateFeedFetchSuccess(feed.ID, store.UpdateFeedFetchSuccessParams{
		CheckedAt:   time.Now().Unix(),
		HTTPStatus:  200,
		NextCheckAt: dueAt,
	}); err != nil {
		t.Fatalf("set next_check_at: %v", err)
	}

	p := New(st, &config.Config{
		PullInterval:      600,
		PullTimeout:       5,
		PullConcurrency:   4,
		PullMaxBackoff:    604800,
		AllowPrivateFeeds: true,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go p.Start(ctx)

	// The loop must pull around the 2s mark (not at 600s).
	deadline := time.Now().Add(8 * time.Second)
	for atomic.LoadInt32(&hits) == 0 && time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
	}
	if got := atomic.LoadInt32(&hits); got == 0 {
		t.Fatal("scheduler did not pull the feed at its next_check_at")
	}

	// The successful pull scheduled the feed at +600s (the global interval);
	// a manual reset to due-now plus Wake must interrupt that sleep.
	reset := int64(900)
	if err := st.UpdateFeed(feed.ID, store.UpdateFeedParams{RefreshIntervalSeconds: &reset}); err != nil {
		t.Fatalf("reset interval: %v", err)
	}
	p.Wake()

	deadline = time.Now().Add(8 * time.Second)
	for atomic.LoadInt32(&hits) < 2 && time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
	}
	if got := atomic.LoadInt32(&hits); got < 2 {
		t.Fatal("Wake did not interrupt the scheduler sleep")
	}
}
