package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/0x2E/fusion/internal/config"
	"github.com/0x2E/fusion/internal/store"
	"github.com/gin-gonic/gin"
)

type recordingPuller struct {
	noopPuller
	wakes atomic.Int64
}

func (p *recordingPuller) Wake() { p.wakes.Add(1) }

func newAnonTestHandler(t *testing.T, puller *recordingPuller) (*Handler, *store.Store) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	st, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}

	// Empty password enables anonymous API access, so tests skip session setup.
	cfg := &config.Config{
		FeverUsername:  "fusion",
		PullTimeout:    30,
		LoginRateLimit: 10,
		LoginWindow:    60,
		LoginBlock:     300,
	}

	h, err := New(st, cfg, puller)
	if err != nil {
		_ = st.Close()
		t.Fatalf("new handler: %v", err)
	}

	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("close store: %v", err)
		}
	})

	return h, st
}

func decodeJSON(t *testing.T, w *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func TestFeedRefreshIntervalLifecycle(t *testing.T) {
	puller := &recordingPuller{}
	h, st := newAnonTestHandler(t, puller)
	router := h.SetupRouter()

	feedPath := ""

	group, err := st.CreateGroup("G")
	if err != nil {
		t.Fatalf("create group: %v", err)
	}

	interval := int64(900)
	createBody := mustJSONBody(t, gin.H{
		"group_id":                 group.ID,
		"name":                     "Feed",
		"link":                     "https://example.com/feed.xml",
		"refresh_interval_seconds": interval,
	})
	w := performRequest(router, http.MethodPost, "/api/feeds", createBody, map[string]string{"Content-Type": "application/json"})
	if w.Code != http.StatusOK {
		t.Fatalf("create feed: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if puller.wakes.Load() == 0 {
		t.Error("expected create feed to wake the puller")
	}

	feeds, err := st.ListFeeds()
	if err != nil || len(feeds) != 1 {
		t.Fatalf("list feeds: %v (len=%d)", err, len(feeds))
	}
	if got := feeds[0].RefreshIntervalSeconds; got == nil || *got != interval {
		t.Fatalf("expected stored interval %d, got %v", interval, got)
	}
	feedPath = "/api/feeds/" + strconv.FormatInt(feeds[0].ID, 10)

	// 0 on update clears the override back to NULL.
	clearBody := mustJSONBody(t, gin.H{"refresh_interval_seconds": 0})
	w = performRequest(router, http.MethodPatch, feedPath, clearBody, map[string]string{"Content-Type": "application/json"})
	if w.Code != http.StatusOK {
		t.Fatalf("clear interval: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	feeds, err = st.ListFeeds()
	if err != nil || len(feeds) != 1 {
		t.Fatalf("list feeds: %v (len=%d)", err, len(feeds))
	}
	if got := feeds[0].RefreshIntervalSeconds; got != nil {
		t.Fatalf("expected cleared interval, got %v", *got)
	}

	// Non-whitelisted values are rejected on both create and update.
	invalidBody := mustJSONBody(t, gin.H{"refresh_interval_seconds": 1234})
	w = performRequest(router, http.MethodPatch, feedPath, invalidBody, map[string]string{"Content-Type": "application/json"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid interval update: expected 400, got %d", w.Code)
	}
	invalidCreate := mustJSONBody(t, gin.H{
		"group_id":                 group.ID,
		"name":                     "Feed 2",
		"link":                     "https://example.com/feed2.xml",
		"refresh_interval_seconds": 1234,
	})
	w = performRequest(router, http.MethodPost, "/api/feeds", invalidCreate, map[string]string{"Content-Type": "application/json"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid interval create: expected 400, got %d", w.Code)
	}

	// 0 on create means "use the global interval" and stores NULL.
	zeroCreate := mustJSONBody(t, gin.H{
		"group_id":                 group.ID,
		"name":                     "Feed 3",
		"link":                     "https://example.com/feed3.xml",
		"refresh_interval_seconds": 0,
	})
	w = performRequest(router, http.MethodPost, "/api/feeds", zeroCreate, map[string]string{"Content-Type": "application/json"})
	if w.Code != http.StatusOK {
		t.Fatalf("zero interval create: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	feeds, err = st.ListFeeds()
	if err != nil {
		t.Fatalf("list feeds: %v", err)
	}
	for _, f := range feeds {
		if f.RefreshIntervalSeconds != nil && *f.RefreshIntervalSeconds == 0 {
			t.Error("stored refresh_interval_seconds must never be 0; NULL means default")
		}
	}
}

func TestGetAppInfo(t *testing.T) {
	puller := &recordingPuller{}
	h, _ := newAnonTestHandler(t, puller)
	router := h.SetupRouter()

	w := performRequest(router, http.MethodGet, "/api/app", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var res struct {
		Data appInfoResponse `json:"data"`
	}
	decodeJSON(t, w, &res)

	if len(res.Data.AllowedRefreshIntervals) == 0 {
		t.Fatal("expected non-empty allowed_refresh_intervals")
	}
	for _, v := range res.Data.AllowedRefreshIntervals {
		if !isAllowedRefreshInterval(v) {
			t.Fatalf("returned value %d is not whitelisted", v)
		}
	}
}

func TestNormalizeCreateRefreshInterval(t *testing.T) {
	zero := int64(0)
	if v, err := normalizeCreateRefreshInterval(nil); err != nil || v != nil {
		t.Errorf("nil: expected nil,nil got %v,%v", v, err)
	}
	if v, err := normalizeCreateRefreshInterval(&zero); err != nil || v != nil {
		t.Errorf("zero: expected nil,nil got %v,%v", v, err)
	}
	valid := int64(900)
	if v, err := normalizeCreateRefreshInterval(&valid); err != nil || v == nil || *v != 900 {
		t.Errorf("900: expected 900,nil got %v,%v", v, err)
	}
	invalid := int64(901)
	if _, err := normalizeCreateRefreshInterval(&invalid); err == nil {
		t.Error("901: expected error")
	}
}
