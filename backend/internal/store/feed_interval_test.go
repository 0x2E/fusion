package store

import (
	"database/sql"
	"testing"
	"time"

	"github.com/0x2E/fusion/internal/model"
)

func setFetchState(t *testing.T, s *Store, feedID, nextCheckAt, retryAfterUntil int64) {
	t.Helper()
	_, err := s.db.Exec(`
		UPDATE feed_fetch_state
		SET next_check_at = :next_check_at, retry_after_until = :retry_after_until, updated_at = unixepoch()
		WHERE feed_id = :feed_id
	`, sql.Named("next_check_at", nextCheckAt),
		sql.Named("retry_after_until", retryAfterUntil),
		sql.Named("feed_id", feedID))
	if err != nil {
		t.Fatalf("set fetch state: %v", err)
	}
}

func TestUpdateFeedRefreshInterval(t *testing.T) {
	s, _ := setupTestDB(t)
	defer closeStore(t, s)

	group, err := s.CreateGroup("G")
	if err != nil {
		t.Fatalf("create group: %v", err)
	}

	interval := int64(900)
	feed, err := s.CreateFeed(group.ID, "Feed", "https://example.com/feed", "", "", &interval)
	if err != nil {
		t.Fatalf("create feed: %v", err)
	}
	if got := feed.RefreshIntervalSeconds; got == nil || *got != interval {
		t.Fatalf("expected interval %d after create, got %v", interval, got)
	}

	// Absent param leaves the override untouched.
	if err := s.UpdateFeed(feed.ID, UpdateFeedParams{}); err != nil {
		t.Fatalf("update feed: %v", err)
	}
	got, err := s.GetFeed(feed.ID)
	if err != nil {
		t.Fatalf("get feed: %v", err)
	}
	if got.RefreshIntervalSeconds == nil || *got.RefreshIntervalSeconds != interval {
		t.Fatalf("expected interval to persist when not updated, got %v", got.RefreshIntervalSeconds)
	}

	// Changing the interval makes the feed due immediately and drops any
	// Retry-After hold.
	setFetchState(t, s, feed.ID, time.Now().Unix()+3600, time.Now().Unix()+3600)

	newInterval := int64(1800)
	if err := s.UpdateFeed(feed.ID, UpdateFeedParams{RefreshIntervalSeconds: &newInterval}); err != nil {
		t.Fatalf("update interval: %v", err)
	}

	got, err = s.GetFeed(feed.ID)
	if err != nil {
		t.Fatalf("get feed: %v", err)
	}
	if got.RefreshIntervalSeconds == nil || *got.RefreshIntervalSeconds != newInterval {
		t.Fatalf("expected interval %d, got %v", newInterval, got.RefreshIntervalSeconds)
	}
	if got.FetchState.NextCheckAt > time.Now().Unix()+5 {
		t.Fatalf("expected next_check_at reset to now, got %d", got.FetchState.NextCheckAt)
	}
	if got.FetchState.RetryAfterUntil != 0 {
		t.Fatalf("expected retry_after_until cleared, got %d", got.FetchState.RetryAfterUntil)
	}

	// ClearRefreshInterval resets to NULL.
	if err := s.UpdateFeed(feed.ID, UpdateFeedParams{ClearRefreshInterval: true}); err != nil {
		t.Fatalf("clear interval: %v", err)
	}
	got, err = s.GetFeed(feed.ID)
	if err != nil {
		t.Fatalf("get feed: %v", err)
	}
	if got.RefreshIntervalSeconds != nil {
		t.Fatalf("expected NULL interval after clear, got %d", *got.RefreshIntervalSeconds)
	}
}

func TestListDueFeedsAndNextWakeTime(t *testing.T) {
	s, _ := setupTestDB(t)
	defer closeStore(t, s)

	group, err := s.CreateGroup("G")
	if err != nil {
		t.Fatalf("create group: %v", err)
	}

	now := time.Now().Unix()
	due, err := s.CreateFeed(group.ID, "Due", "https://example.com/due", "", "", nil)
	if err != nil {
		t.Fatalf("create due feed: %v", err)
	}
	future, err := s.CreateFeed(group.ID, "Future", "https://example.com/future", "", "", nil)
	if err != nil {
		t.Fatalf("create future feed: %v", err)
	}
	suspended, err := s.CreateFeed(group.ID, "Suspended", "https://example.com/suspended", "", "", nil)
	if err != nil {
		t.Fatalf("create suspended feed: %v", err)
	}
	retryHold, err := s.CreateFeed(group.ID, "Retry", "https://example.com/retry", "", "", nil)
	if err != nil {
		t.Fatalf("create retry feed: %v", err)
	}

	setFetchState(t, s, due.ID, now-10, 0)
	setFetchState(t, s, future.ID, now+3600, 0)
	setFetchState(t, s, suspended.ID, now-10, 0)
	// Due by next_check_at but held by Retry-After: not listed, and it pushes
	// the next wake time out to the retry deadline.
	setFetchState(t, s, retryHold.ID, now-10, now+600)

	if err := s.UpdateFeed(suspended.ID, UpdateFeedParams{Suspended: ptrBool(true)}); err != nil {
		t.Fatalf("suspend feed: %v", err)
	}

	feeds, err := s.ListDueFeeds(now)
	if err != nil {
		t.Fatalf("list due feeds: %v", err)
	}
	// The retry-held feed is returned by the SQL prefilter; ShouldSkip in the
	// puller is what skips it. Assert only that due and suspended behave.
	byID := map[int64]*model.Feed{}
	for _, f := range feeds {
		byID[f.ID] = f
	}
	if _, ok := byID[due.ID]; !ok {
		t.Error("expected due feed to be listed")
	}
	if _, ok := byID[future.ID]; ok {
		t.Error("expected future feed to be excluded")
	}
	if _, ok := byID[suspended.ID]; ok {
		t.Error("expected suspended feed to be excluded")
	}

	next, err := s.NextWakeTime()
	if err != nil {
		t.Fatalf("next wake time: %v", err)
	}
	// The past-due feed is the earliest wake time (the scheduler clamps the
	// sleep to its minimum instead of spinning).
	if next > now {
		t.Fatalf("expected past wake time for due feed, got %d (now=%d)", next, now)
	}

	// Once nothing is past due, the Retry-After hold pushes the wake time out.
	setFetchState(t, s, due.ID, now+3600, 0)
	next, err = s.NextWakeTime()
	if err != nil {
		t.Fatalf("next wake time: %v", err)
	}
	if next < now+500 || next > now+700 {
		t.Fatalf("expected next wake near %d from retry hold, got %d", now+600, next)
	}
}

func ptrBool(v bool) *bool { return &v }
