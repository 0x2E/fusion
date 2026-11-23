package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type batchCreateFeedsBody struct {
	Data struct {
		Created int      `json:"created"`
		Failed  int      `json:"failed"`
		Errors  []string `json:"errors"`
	} `json:"data"`
}

func postBatchCreateFeeds(t *testing.T, h *Handler, feeds []gin.H) (int, batchCreateFeedsBody) {
	t.Helper()

	r := newTestRouter()
	r.POST("/api/feeds/batch", h.batchCreateFeeds)
	w := performRequest(
		r,
		http.MethodPost,
		"/api/feeds/batch",
		mustJSONBody(t, gin.H{"feeds": feeds}),
		map[string]string{"Content-Type": "application/json"},
	)

	var body batchCreateFeedsBody
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil && w.Code == http.StatusOK {
		t.Fatalf("decode response: %v", err)
	}
	return w.Code, body
}

// IP literals keep URL validation offline: TEST-NET-1 addresses parse as
// public without a DNS lookup, while example.invalid always fails to resolve.
func TestBatchCreateFeedsImportsValidLinksAndReportsFailures(t *testing.T) {
	h, st := newAnonTestHandler(t, &recordingPuller{})

	group, err := st.CreateGroup("Batch")
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	existing, err := st.CreateFeed(group.ID, "Existing", "http://192.0.2.11/existing.xml", "", "", nil)
	if err != nil {
		t.Fatalf("create existing feed: %v", err)
	}

	validOne := "http://192.0.2.10/one.xml"
	validTwo := "http://192.0.2.12/two.xml"
	deadLink := "http://feeds.example.invalid/rss"
	duplicateLink := existing.Link

	code, body := postBatchCreateFeeds(t, h, []gin.H{
		{"group_id": group.ID, "name": "One", "link": validOne},
		{"group_id": group.ID, "name": "Dead", "link": deadLink},
		{"group_id": group.ID, "name": "Duplicate", "link": duplicateLink},
		{"group_id": group.ID, "name": "Two", "link": validTwo},
	})
	if code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", code)
	}
	if body.Data.Created != 2 {
		t.Fatalf("expected 2 created feeds, got %d", body.Data.Created)
	}
	if body.Data.Failed != 2 {
		t.Fatalf("expected 2 failed feeds, got %d (%v)", body.Data.Failed, body.Data.Errors)
	}
	if len(body.Data.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d (%v)", len(body.Data.Errors), body.Data.Errors)
	}

	joined := strings.Join(body.Data.Errors, "\n")
	for _, want := range []string{"invalid URL " + deadLink, "duplicate feed: " + duplicateLink} {
		if !strings.Contains(joined, want) {
			t.Errorf("expected errors to contain %q, got %v", want, body.Data.Errors)
		}
	}

	feeds, err := st.ListFeeds()
	if err != nil {
		t.Fatalf("list feeds: %v", err)
	}
	links := make(map[string]bool, len(feeds))
	for _, feed := range feeds {
		links[feed.Link] = true
	}
	for _, want := range []string{validOne, validTwo, existing.Link} {
		if !links[want] {
			t.Errorf("expected feed %q to be stored, got %v", want, links)
		}
	}
	if links[deadLink] {
		t.Errorf("invalid feed %q must not be stored", deadLink)
	}
}

func TestBatchCreateFeedsReportsValidationErrorsInInputOrder(t *testing.T) {
	h, st := newAnonTestHandler(t, &recordingPuller{})

	group, err := st.CreateGroup("Batch")
	if err != nil {
		t.Fatalf("create group: %v", err)
	}

	validOne := "http://192.0.2.10/one.xml"
	validTwo := "http://192.0.2.20/two.xml"
	validThree := "http://192.0.2.30/three.xml"
	deadDomain := "http://feeds.example.invalid/rss"
	badScheme := "ftp://example.com/feed.xml"
	privateHost := "http://127.0.0.1/feed.xml"

	code, body := postBatchCreateFeeds(t, h, []gin.H{
		{"group_id": group.ID, "name": "One", "link": validOne},
		{"group_id": group.ID, "name": "Dead domain", "link": deadDomain},
		{"group_id": group.ID, "name": "Bad scheme", "link": badScheme},
		{"group_id": group.ID, "name": "Two", "link": validTwo},
		{"group_id": group.ID, "name": "Private host", "link": privateHost},
		{"group_id": group.ID, "name": "Three", "link": validThree},
	})
	if code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", code)
	}
	if body.Data.Created != 3 {
		t.Fatalf("expected 3 created feeds, got %d", body.Data.Created)
	}
	if body.Data.Failed != 3 {
		t.Fatalf("expected 3 failed feeds, got %d (%v)", body.Data.Failed, body.Data.Errors)
	}

	wantErrorPrefixes := []string{
		"invalid URL " + deadDomain + ": resolve host: ",
		"invalid URL " + badScheme + ": unsupported url scheme: ftp",
		"invalid URL " + privateHost + ": private host is not allowed",
	}
	if len(body.Data.Errors) != len(wantErrorPrefixes) {
		t.Fatalf("expected %d errors, got %d (%v)", len(wantErrorPrefixes), len(body.Data.Errors), body.Data.Errors)
	}
	for i, want := range wantErrorPrefixes {
		if !strings.HasPrefix(body.Data.Errors[i], want) {
			t.Errorf("error %d: expected prefix %q, got %q", i, want, body.Data.Errors[i])
		}
	}

	feeds, err := st.ListFeeds()
	if err != nil {
		t.Fatalf("list feeds: %v", err)
	}
	links := make(map[string]bool, len(feeds))
	for _, feed := range feeds {
		links[feed.Link] = true
	}
	for _, want := range []string{validOne, validTwo, validThree} {
		if !links[want] {
			t.Errorf("expected feed %q to be stored, got %v", want, links)
		}
	}
	for _, reject := range []string{deadDomain, badScheme, privateHost} {
		if links[reject] {
			t.Errorf("invalid feed %q must not be stored", reject)
		}
	}
}

func TestBatchCreateFeedsAllInvalidReturnsOK(t *testing.T) {
	h, st := newAnonTestHandler(t, &recordingPuller{})

	group, err := st.CreateGroup("Batch")
	if err != nil {
		t.Fatalf("create group: %v", err)
	}

	code, body := postBatchCreateFeeds(t, h, []gin.H{
		{"group_id": group.ID, "name": "Dead", "link": "http://feeds.example.invalid/rss"},
		{"group_id": group.ID, "name": "Wrong scheme", "link": "ftp://example.com/feed.xml"},
	})
	if code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", code)
	}
	if body.Data.Created != 0 || body.Data.Failed != 2 || len(body.Data.Errors) != 2 {
		t.Fatalf("expected 0 created and 2 failed, got created=%d failed=%d errors=%v", body.Data.Created, body.Data.Failed, body.Data.Errors)
	}

	feeds, err := st.ListFeeds()
	if err != nil {
		t.Fatalf("list feeds: %v", err)
	}
	if len(feeds) != 0 {
		t.Fatalf("expected no feeds, got %d", len(feeds))
	}
}

func TestBatchCreateFeedsMalformedBodyReturnsBadRequest(t *testing.T) {
	h, _ := newAnonTestHandler(t, &recordingPuller{})

	r := newTestRouter()
	r.POST("/api/feeds/batch", h.batchCreateFeeds)

	tests := []struct {
		name string
		body any
	}{
		{"feeds is not a list", gin.H{"feeds": "not-a-list"}},
		{"feed misses required fields", gin.H{"feeds": []gin.H{{"name": "No link"}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(
				r,
				http.MethodPost,
				"/api/feeds/batch",
				mustJSONBody(t, tt.body),
				map[string]string{"Content-Type": "application/json"},
			)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected status 400, got %d", w.Code)
			}
		})
	}
}

func TestFeedHandlersReportInvalidURLReason(t *testing.T) {
	h, _ := newAnonTestHandler(t, &recordingPuller{})
	invalidLink := "ftp://example.com/feed.xml"

	r := newTestRouter()
	r.POST("/api/feeds", h.createFeed)
	r.PATCH("/api/feeds/:id", h.updateFeed)
	r.POST("/api/feeds/validate", h.validateFeed)

	tests := []struct {
		name   string
		method string
		path   string
		body   gin.H
	}{
		{"create", http.MethodPost, "/api/feeds", gin.H{"group_id": 1, "name": "Feed", "link": invalidLink}},
		{"update", http.MethodPatch, "/api/feeds/1", gin.H{"link": invalidLink}},
		{"validate", http.MethodPost, "/api/feeds/validate", gin.H{"url": invalidLink}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(
				r,
				tt.method,
				tt.path,
				mustJSONBody(t, tt.body),
				map[string]string{"Content-Type": "application/json"},
			)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected status 400, got %d", w.Code)
			}

			var res struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if !strings.HasPrefix(res.Error, "invalid URL: ") {
				t.Errorf("expected message to start with %q, got %q", "invalid URL: ", res.Error)
			}
			if !strings.Contains(res.Error, "unsupported url scheme: ftp") {
				t.Errorf("expected message to include the validation reason, got %q", res.Error)
			}
		})
	}
}
