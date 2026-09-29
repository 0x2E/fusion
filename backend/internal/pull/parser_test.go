package pull

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/0x2E/fusion/internal/model"
	"github.com/mmcdole/gofeed"
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

func TestNormalizeTitle(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "plain title unchanged", raw: "Belief => Actions => Results", want: "Belief => Actions => Results"},
		{name: "entity decoded once", raw: "Belief =&gt; Actions =&gt; Results", want: "Belief => Actions => Results"},
		{name: "named and numeric entities decoded", raw: "A &amp; B &#8212; C &quot;D&quot;", want: `A & B — C "D"`},
		{name: "tags stripped", raw: "Announcing <b>Go</b> 1.24 &mdash; <i>notes</i>", want: "Announcing Go 1.24 — notes"},
		{name: "entity-encoded tags stay visible", raw: "Use &lt;b&gt; for bold", want: "Use <b> for bold"},
		{name: "double-escaped entity decoded once", raw: "AT&amp;amp;T &amp;lt;tag&amp;gt;", want: "AT&amp;T &lt;tag&gt;"},
		{name: "whitespace collapsed edges trimmed", raw: "  <b>  spaced  </b>  ", want: "spaced"},
		{name: "empty", raw: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeTitle(tt.raw); got != tt.want {
				t.Fatalf("normalizeTitle(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestFetchAndParseNormalizesHTMLTitles(t *testing.T) {
	// Mirrors the feed shape from #97: Atom type="html" titles arrive from
	// gofeed with escaped entities and markup still embedded.
	feedXML := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Example &amp; Blog</title>
  <entry>
    <title type="html">Belief =&gt; Actions =&gt; Results</title>
    <id>urn:entry-1</id>
    <updated>2026-01-01T00:00:00Z</updated>
  </entry>
  <entry>
    <title type="html">A &lt;b&gt;bold&lt;/b&gt; title &amp;amp; more</title>
    <id>urn:entry-2</id>
    <updated>2026-01-02T00:00:00Z</updated>
  </entry>
</feed>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/atom+xml")
		_, _ = w.Write([]byte(feedXML))
	}))
	defer server.Close()

	result, err := FetchAndParse(context.Background(), &model.Feed{Link: server.URL}, 5*time.Second, true)
	if err != nil {
		t.Fatalf("FetchAndParse() failed: %v", err)
	}

	if result.FeedTitle != "Example & Blog" {
		t.Fatalf("feed title = %q, want %q", result.FeedTitle, "Example & Blog")
	}

	wantTitles := []string{
		"Belief => Actions => Results",
		// gofeed decodes type="html" entities into real tags; normalizeTitle
		// then strips them, yielding the text content a browser would show.
		"A bold title & more",
	}
	if len(result.Items) != len(wantTitles) {
		t.Fatalf("got %d items, want %d", len(result.Items), len(wantTitles))
	}
	for i, want := range wantTitles {
		if result.Items[i].Title != want {
			t.Fatalf("item %d title = %q, want %q", i, result.Items[i].Title, want)
		}
	}
}
