package pull

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/patrickjmcd/reedme/internal/model"
)

func TestFetchAndParseSendsConditionalHeadersAndHandles304(t *testing.T) {
	var gotIfNoneMatch string
	var gotIfModifiedSince string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotIfNoneMatch = r.Header.Get("If-None-Match")
		gotIfModifiedSince = r.Header.Get("If-Modified-Since")
		w.Header().Set("ETag", `"next-etag"`)
		w.Header().Set("Last-Modified", "Mon, 02 Jan 2006 15:04:05 GMT")
		w.WriteHeader(http.StatusNotModified)
	}))
	defer server.Close()

	feed := &model.Feed{
		Link: server.URL,
		FetchState: model.FeedFetchState{
			ETag:         `"prev-etag"`,
			LastModified: "Mon, 01 Jan 2006 15:04:05 GMT",
		},
	}

	result, err := FetchAndParse(context.Background(), feed, 5*time.Second, true)
	if err != nil {
		t.Fatalf("FetchAndParse() failed: %v", err)
	}

	if gotIfNoneMatch != `"prev-etag"` {
		t.Fatalf("expected If-None-Match header set, got %q", gotIfNoneMatch)
	}
	if gotIfModifiedSince != "Mon, 01 Jan 2006 15:04:05 GMT" {
		t.Fatalf("expected If-Modified-Since header set, got %q", gotIfModifiedSince)
	}
	if !result.NotModified {
		t.Fatalf("expected NotModified=true, got false")
	}
	if result.HTTPStatus != http.StatusNotModified {
		t.Fatalf("expected status 304, got %d", result.HTTPStatus)
	}
}

func TestFetchAndParseParsesCacheMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Header().Set("Cache-Control", "public, max-age=600")
		w.Header().Set("Expires", time.Unix(1700000600, 0).UTC().Format(http.TimeFormat))
		w.Header().Set("ETag", `"v1"`)
		w.Header().Set("Last-Modified", "Mon, 02 Jan 2006 15:04:05 GMT")
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel><title>Demo</title><link>https://example.com</link>
<item><guid>g1</guid><title>Item</title><link>https://example.com/1</link></item>
</channel></rss>`))
	}))
	defer server.Close()

	feed := &model.Feed{Link: server.URL}
	result, err := FetchAndParse(context.Background(), feed, 5*time.Second, true)
	if err != nil {
		t.Fatalf("FetchAndParse() failed: %v", err)
	}

	if result.NotModified {
		t.Fatal("expected NotModified=false for 200 response")
	}
	if result.HTTPStatus != http.StatusOK {
		t.Fatalf("expected status 200, got %d", result.HTTPStatus)
	}
	if result.CacheControl != "public, max-age=600" {
		t.Fatalf("expected cache-control metadata, got %q", result.CacheControl)
	}
	if result.ExpiresAt != 1700000600 {
		t.Fatalf("expected expires_at=1700000600, got %d", result.ExpiresAt)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}
}

func TestMapItemFallbackGUIDWhenMissingGUIDAndLink(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	item := &gofeed.Item{
		Title:           "Example",
		Description:     "Body",
		PublishedParsed: &now,
	}

	parsed := mapItem(item, nil)

	if !strings.HasPrefix(parsed.GUID, "generated:") {
		t.Fatalf("expected generated GUID, got %q", parsed.GUID)
	}
	if parsed.Link != "" {
		t.Fatalf("expected empty link, got %q", parsed.Link)
	}
}

func TestMapItemUsesNormalizedLinkAsGUIDFallback(t *testing.T) {
	baseURL, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("parse base URL: %v", err)
	}

	item := &gofeed.Item{Link: "/news/1"}
	parsed := mapItem(item, baseURL)

	if parsed.Link != "https://example.com/news/1" {
		t.Fatalf("expected absolute link, got %q", parsed.Link)
	}
	if parsed.GUID != parsed.Link {
		t.Fatalf("expected GUID fallback to normalized link, got guid=%q link=%q", parsed.GUID, parsed.Link)
	}
}

func TestMapItemDoesNotUseBaseURLWhenLinkIsMissing(t *testing.T) {
	baseURL, err := url.Parse("https://example.com/news")
	if err != nil {
		t.Fatalf("parse base URL: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	item := &gofeed.Item{
		Title:           "No link post",
		Description:     "content",
		PublishedParsed: &now,
	}

	parsed := mapItem(item, baseURL)

	if parsed.Link != "" {
		t.Fatalf("expected empty link when source link is missing, got %q", parsed.Link)
	}
	if !strings.HasPrefix(parsed.GUID, "generated:") {
		t.Fatalf("expected generated GUID, got %q", parsed.GUID)
	}
}

func TestFallbackGUIDIgnoresSyntheticPubDate(t *testing.T) {
	g1 := fallbackGUID("same title", "same content", 1700000000, false)
	g2 := fallbackGUID("same title", "same content", 1800000000, false)

	if g1 != g2 {
		t.Fatalf("expected stable GUID without source pub date, got %q and %q", g1, g2)
	}
}

func TestFallbackGUIDUsesSourcePubDateWhenProvided(t *testing.T) {
	g1 := fallbackGUID("same title", "same content", 1700000000, true)
	g2 := fallbackGUID("same title", "same content", 1800000000, true)

	if g1 == g2 {
		t.Fatalf("expected different GUID when source pub date differs, got %q", g1)
	}
}
