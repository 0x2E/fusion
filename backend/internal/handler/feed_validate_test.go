package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/0x2E/fusion/internal/config"
)

type validateFeedBody struct {
	Data struct {
		Feeds []struct {
			Title string `json:"title"`
			Link  string `json:"link"`
		} `json:"feeds"`
	} `json:"data"`
}

func postValidateFeed(t *testing.T, h *Handler, url string) (int, validateFeedBody) {
	t.Helper()

	r := newTestRouter()
	r.POST("/api/feeds/validate", h.validateFeed)
	w := performRequest(
		r,
		http.MethodPost,
		"/api/feeds/validate",
		strings.NewReader(fmt.Sprintf(`{"url":%q}`, url)),
		map[string]string{"Content-Type": "application/json"},
	)

	var body validateFeedBody
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil && w.Code == http.StatusOK {
		t.Fatalf("decode response: %v", err)
	}
	return w.Code, body
}

// The add-feed dialog calls validate before creating a feed and blocks the
// creation when no feed is found, so a malformed document must yield an empty
// feeds list rather than an error or a fabricated entry.
func TestValidateFeedMalformedXMLReturnsNoFeeds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = fmt.Fprint(w, `<?xml version="1.0"?><rss><channel><title>Broken`)
	}))
	defer server.Close()

	h := &Handler{config: &config.Config{AllowPrivateFeeds: true}}

	code, body := postValidateFeed(t, h, server.URL)
	if code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", code)
	}
	if len(body.Data.Feeds) != 0 {
		t.Fatalf("expected 0 feeds for malformed XML, got %d: %+v", len(body.Data.Feeds), body.Data.Feeds)
	}
}

func TestValidateFeedValidRSSReturnsFeed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Serve the feed only at its real path so the finder's common-path
		// guesses (index.xml, atom.xml, ...) miss like they would in production.
		if r.URL.Path != "/feed.xml" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel><title>Tech Daily</title><link>https://example.com</link>
<item><guid>g1</guid><title>Item</title><link>https://example.com/1</link></item>
</channel></rss>`)
	}))
	defer server.Close()

	h := &Handler{config: &config.Config{AllowPrivateFeeds: true}}

	code, body := postValidateFeed(t, h, server.URL+"/feed.xml")
	if code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", code)
	}
	if len(body.Data.Feeds) != 1 {
		t.Fatalf("expected 1 feed, got %d: %+v", len(body.Data.Feeds), body.Data.Feeds)
	}
	if body.Data.Feeds[0].Title != "Tech Daily" {
		t.Fatalf("expected title %q, got %q", "Tech Daily", body.Data.Feeds[0].Title)
	}
	if body.Data.Feeds[0].Link != server.URL+"/feed.xml" {
		t.Fatalf("expected link %q, got %q", server.URL+"/feed.xml", body.Data.Feeds[0].Link)
	}
}

func TestValidateFeedUnreachableURLReturnsNoFeeds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	url := server.URL
	server.Close()

	h := &Handler{config: &config.Config{AllowPrivateFeeds: true}}

	code, body := postValidateFeed(t, h, url)
	if code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", code)
	}
	if len(body.Data.Feeds) != 0 {
		t.Fatalf("expected 0 feeds for unreachable URL, got %d", len(body.Data.Feeds))
	}
}
