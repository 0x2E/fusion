package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/0x2E/fusion/internal/config"
	"github.com/0x2E/fusion/internal/store"
)

type noopPuller struct{}

func (noopPuller) RefreshFeed(context.Context, int64) error { return nil }

func (noopPuller) RefreshAll(context.Context) (int, error) { return 0, nil }

func newFeverTestHandler(t *testing.T) (*Handler, *store.Store) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	st, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}

	cfg := &config.Config{
		Password:       "secret",
		FeverUsername:  "fusion",
		PullTimeout:    30,
		LoginRateLimit: 10,
		LoginWindow:    60,
		LoginBlock:     300,
	}

	h, err := New(st, cfg, noopPuller{})
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

func feverRequestBody(apiKey string, extra url.Values) string {
	values := url.Values{
		"api":     {"1"},
		"api_key": {apiKey},
	}
	for key, vals := range extra {
		for _, val := range vals {
			values.Add(key, val)
		}
	}

	return values.Encode()
}

func TestFeverAuthFailure(t *testing.T) {
	h, _ := newFeverTestHandler(t)

	r := newTestRouter()
	r.POST("/fever", h.fever)

	w := performRequest(
		r,
		http.MethodPost,
		"/fever",
		strings.NewReader(feverRequestBody("wrong", nil)),
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
	)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if payload["auth"] != float64(0) {
		t.Fatalf("expected auth=0, got %#v", payload["auth"])
	}
	if payload["api_version"] != float64(feverAPIVersion) {
		t.Fatalf("expected api_version=%d, got %#v", feverAPIVersion, payload["api_version"])
	}
}

func TestFeverReadAndMarkFlows(t *testing.T) {
	h, st := newFeverTestHandler(t)

	group, err := st.CreateGroup("Tech")
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	feed, err := st.CreateFeed(group.ID, "Fusion Feed", "https://example.com/rss.xml", "https://example.com", "")
	if err != nil {
		t.Fatalf("create feed: %v", err)
	}
	item, err := st.CreateItem(feed.ID, "guid-1", "Entry 1", "https://example.com/entry-1", "<p>Hello</p>", 1700000000)
	if err != nil {
		t.Fatalf("create item: %v", err)
	}

	r := newTestRouter()
	r.POST("/fever", h.fever)

	apiKey := deriveFeverAPIKey("fusion", "secret")

	markSavedBody := feverRequestBody(apiKey, url.Values{
		"mark": {"item"},
		"id":   {strconv.FormatInt(item.ID, 10)},
		"as":   {"saved"},
	})
	w := performRequest(
		r,
		http.MethodPost,
		"/fever",
		strings.NewReader(markSavedBody),
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for mark saved, got %d", w.Code)
	}

	exists, err := st.BookmarkExists(item.Link)
	if err != nil {
		t.Fatalf("check bookmark exists: %v", err)
	}
	if !exists {
		t.Fatal("expected bookmark to be created")
	}

	markReadBody := feverRequestBody(apiKey, url.Values{
		"mark": {"item"},
		"id":   {strconv.FormatInt(item.ID, 10)},
		"as":   {"read"},
	})
	w = performRequest(
		r,
		http.MethodPost,
		"/fever",
		strings.NewReader(markReadBody),
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for mark read, got %d", w.Code)
	}

	updatedItem, err := st.GetItem(item.ID)
	if err != nil {
		t.Fatalf("get item: %v", err)
	}
	if updatedItem.Unread {
		t.Fatal("expected item to be marked as read")
	}

	listBody := feverRequestBody(apiKey, url.Values{
		"groups":          {"1"},
		"feeds":           {"1"},
		"items":           {"1"},
		"saved_item_ids":  {"1"},
		"unread_item_ids": {"1"},
	})
	w = performRequest(
		r,
		http.MethodPost,
		"/fever",
		strings.NewReader(listBody),
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for list request, got %d", w.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if payload["auth"] != float64(1) {
		t.Fatalf("expected auth=1, got %#v", payload["auth"])
	}

	expectedSavedIDs := strconv.FormatInt(item.ID, 10)
	if payload["saved_item_ids"] != expectedSavedIDs {
		t.Fatalf("expected saved_item_ids to be %q, got %#v", expectedSavedIDs, payload["saved_item_ids"])
	}
	if payload["unread_item_ids"] != "" {
		t.Fatalf("expected unread_item_ids to be empty, got %#v", payload["unread_item_ids"])
	}

	items, ok := payload["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("expected exactly one item, got %#v", payload["items"])
	}

	itemObj, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("expected item object, got %#v", items[0])
	}

	if itemObj["is_saved"] != float64(1) {
		t.Fatalf("expected is_saved=1, got %#v", itemObj["is_saved"])
	}
	if itemObj["is_read"] != float64(1) {
		t.Fatalf("expected is_read=1, got %#v", itemObj["is_read"])
	}
}

func TestFeverMarkSavedLinksExistingBookmarkToItem(t *testing.T) {
	h, st := newFeverTestHandler(t)

	group, err := st.CreateGroup("Tech")
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	feed, err := st.CreateFeed(group.ID, "Fusion Feed", "https://example.com/rss.xml", "https://example.com", "")
	if err != nil {
		t.Fatalf("create feed: %v", err)
	}
	item, err := st.CreateItem(feed.ID, "guid-2", "Entry 2", "https://example.com/entry-2", "<p>Hello 2</p>", 1700000001)
	if err != nil {
		t.Fatalf("create item: %v", err)
	}

	if _, err := st.CreateBookmark(nil, item.Link, "Old Snapshot", "<p>old</p>", item.PubDate, feed.Name); err != nil {
		t.Fatalf("create preexisting bookmark: %v", err)
	}

	r := newTestRouter()
	r.POST("/fever", h.fever)

	apiKey := deriveFeverAPIKey("fusion", "secret")
	body := feverRequestBody(apiKey, url.Values{
		"mark": {"item"},
		"id":   {strconv.FormatInt(item.ID, 10)},
		"as":   {"saved"},
	})

	w := performRequest(
		r,
		http.MethodPost,
		"/fever",
		strings.NewReader(body),
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	savedIDs, err := st.ListSavedItemIDs()
	if err != nil {
		t.Fatalf("list saved item ids: %v", err)
	}
	if len(savedIDs) != 1 || savedIDs[0] != item.ID {
		t.Fatalf("expected saved item ids to contain %d, got %#v", item.ID, savedIDs)
	}
}

func TestHasFeverFlag(t *testing.T) {
	form := url.Values{}

	if hasFeverFlag(form, "items") {
		t.Fatal("expected missing flag to be false")
	}

	form.Set("items", "")
	if !hasFeverFlag(form, "items") {
		t.Fatal("expected empty but present flag to be true")
	}

	falseValues := []string{"0", "false", "FALSE", "no", "off", "  false  "}
	for _, val := range falseValues {
		form.Set("items", val)
		if hasFeverFlag(form, "items") {
			t.Fatalf("expected value %q to be false", val)
		}
	}

	trueValues := []string{"1", "true", "yes", "on", "anything"}
	for _, val := range trueValues {
		form.Set("items", val)
		if !hasFeverFlag(form, "items") {
			t.Fatalf("expected value %q to be true", val)
		}
	}
}

func TestFeverFeedsIncludesFeedsGroups(t *testing.T) {
	h, st := newFeverTestHandler(t)

	group, err := st.CreateGroup("Tech")
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	if _, err := st.CreateFeed(group.ID, "Fusion Feed", "https://example.com/rss.xml", "https://example.com", ""); err != nil {
		t.Fatalf("create feed: %v", err)
	}

	r := newTestRouter()
	r.POST("/fever", h.fever)

	apiKey := deriveFeverAPIKey("fusion", "secret")
	body := feverRequestBody(apiKey, url.Values{"feeds": {""}})

	w := performRequest(
		r,
		http.MethodPost,
		"/fever?api&feeds",
		strings.NewReader(body),
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	feedsGroups, ok := payload["feeds_groups"].([]any)
	if !ok || len(feedsGroups) == 0 {
		t.Fatalf("expected feeds_groups in feeds response, got %#v", payload["feeds_groups"])
	}
}

func TestFeverItemsWithMaxIDZeroReturnsRecentItems(t *testing.T) {
	h, st := newFeverTestHandler(t)

	group, err := st.CreateGroup("Tech")
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	feed, err := st.CreateFeed(group.ID, "Fusion Feed", "https://example.com/rss.xml", "https://example.com", "")
	if err != nil {
		t.Fatalf("create feed: %v", err)
	}
	if _, err := st.CreateItem(feed.ID, "guid-3", "Entry 3", "https://example.com/entry-3", "<p>Hello 3</p>", 1700000002); err != nil {
		t.Fatalf("create item: %v", err)
	}

	r := newTestRouter()
	r.POST("/fever", h.fever)

	apiKey := deriveFeverAPIKey("fusion", "secret")
	body := feverRequestBody(apiKey, url.Values{"items": {""}, "max_id": {"0"}})

	w := performRequest(
		r,
		http.MethodPost,
		"/fever?api&items&max_id=0",
		strings.NewReader(body),
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	items, ok := payload["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("expected non-empty items, got %#v", payload["items"])
	}

	if payload["total_items"] != float64(1) {
		t.Fatalf("expected total_items=1, got %#v", payload["total_items"])
	}
}

func TestParseListFeverItemsParamsLimitsWithIDs(t *testing.T) {
	ids := make([]string, 0, 60)
	for i := 1; i <= 60; i++ {
		ids = append(ids, strconv.Itoa(i))
	}

	params, err := parseListFeverItemsParams(url.Values{"with_ids": {strings.Join(ids, ",")}})
	if err != nil {
		t.Fatalf("parse params: %v", err)
	}

	if len(params.WithIDs) != feverItemsLimit {
		t.Fatalf("expected with_ids length %d, got %d", feverItemsLimit, len(params.WithIDs))
	}
}

func TestFeverFaviconsHaveDataURI(t *testing.T) {
	h, st := newFeverTestHandler(t)

	group, err := st.CreateGroup("Tech")
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	if _, err := st.CreateFeed(group.ID, "Fusion Feed", "https://example.com/rss.xml", "https://example.com", ""); err != nil {
		t.Fatalf("create feed: %v", err)
	}

	r := newTestRouter()
	r.POST("/fever", h.fever)

	apiKey := deriveFeverAPIKey("fusion", "secret")
	body := feverRequestBody(apiKey, url.Values{"favicons": {""}})

	w := performRequest(
		r,
		http.MethodPost,
		"/fever?api&favicons",
		strings.NewReader(body),
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	favicons, ok := payload["favicons"].([]any)
	if !ok || len(favicons) == 0 {
		t.Fatalf("expected favicons list, got %#v", payload["favicons"])
	}

	first, ok := favicons[0].(map[string]any)
	if !ok {
		t.Fatalf("expected favicon object, got %#v", favicons[0])
	}

	data, _ := first["data"].(string)
	if !strings.HasPrefix(data, "image/") || !strings.Contains(data, ";base64,") {
		t.Fatalf("expected favicon data URI, got %q", data)
	}
}

func TestFeverMarkFeedReadRespectsBefore(t *testing.T) {
	h, st := newFeverTestHandler(t)

	group, err := st.CreateGroup("Tech")
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	feed, err := st.CreateFeed(group.ID, "Fusion Feed", "https://example.com/rss.xml", "https://example.com", "")
	if err != nil {
		t.Fatalf("create feed: %v", err)
	}
	oldItem, err := st.CreateItem(feed.ID, "guid-old", "Old", "https://example.com/old", "<p>old</p>", 100)
	if err != nil {
		t.Fatalf("create old item: %v", err)
	}
	newItem, err := st.CreateItem(feed.ID, "guid-new", "New", "https://example.com/new", "<p>new</p>", 200)
	if err != nil {
		t.Fatalf("create new item: %v", err)
	}

	r := newTestRouter()
	r.POST("/fever", h.fever)

	apiKey := deriveFeverAPIKey("fusion", "secret")
	body := feverRequestBody(apiKey, url.Values{
		"mark":   {"feed"},
		"as":     {"read"},
		"id":     {strconv.FormatInt(feed.ID, 10)},
		"before": {"150"},
	})

	w := performRequest(
		r,
		http.MethodPost,
		"/fever?api",
		strings.NewReader(body),
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	updatedOld, err := st.GetItem(oldItem.ID)
	if err != nil {
		t.Fatalf("get old item: %v", err)
	}
	updatedNew, err := st.GetItem(newItem.ID)
	if err != nil {
		t.Fatalf("get new item: %v", err)
	}

	if updatedOld.Unread {
		t.Fatal("expected old item to be marked read")
	}
	if !updatedNew.Unread {
		t.Fatal("expected new item to remain unread")
	}
}

func TestFeverMarkReadSupportsCSVItemIDs(t *testing.T) {
	h, st := newFeverTestHandler(t)

	group, err := st.CreateGroup("Tech")
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	feed, err := st.CreateFeed(group.ID, "Fusion Feed", "https://example.com/rss.xml", "https://example.com", "")
	if err != nil {
		t.Fatalf("create feed: %v", err)
	}
	item1, err := st.CreateItem(feed.ID, "guid-4", "Entry 4", "https://example.com/entry-4", "<p>Hello 4</p>", 1700000003)
	if err != nil {
		t.Fatalf("create item 1: %v", err)
	}
	item2, err := st.CreateItem(feed.ID, "guid-5", "Entry 5", "https://example.com/entry-5", "<p>Hello 5</p>", 1700000004)
	if err != nil {
		t.Fatalf("create item 2: %v", err)
	}

	r := newTestRouter()
	r.POST("/fever", h.fever)

	apiKey := deriveFeverAPIKey("fusion", "secret")
	extra := url.Values{
		"mark": {"item"},
		"as":   {"read"},
	}
	extra.Add("id", strconv.FormatInt(item1.ID, 10))
	extra.Add("id", strconv.FormatInt(item2.ID, 10))
	body := feverRequestBody(apiKey, url.Values{
		"mark": extra["mark"],
		"as":   extra["as"],
		"id":   extra["id"],
	})

	w := performRequest(
		r,
		http.MethodPost,
		"/fever?api",
		strings.NewReader(body),
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload["unread_item_ids"] != "" {
		t.Fatalf("expected unread_item_ids to be empty after mark read, got %#v", payload["unread_item_ids"])
	}

	updated1, err := st.GetItem(item1.ID)
	if err != nil {
		t.Fatalf("get item1: %v", err)
	}
	updated2, err := st.GetItem(item2.ID)
	if err != nil {
		t.Fatalf("get item2: %v", err)
	}

	if updated1.Unread || updated2.Unread {
		t.Fatalf("expected both items to be read, got unread states: %v %v", updated1.Unread, updated2.Unread)
	}
}
