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
		// #97: one more strict decode of what gofeed produced
		{name: "entity decoded once", raw: "Belief =&gt; Actions =&gt; Results", want: "Belief => Actions => Results"},
		{name: "named and numeric entities", raw: "A &amp; B &#8212; C &quot;D&quot;", want: `A & B — C "D"`},
		{name: "double-escaped entity peels one layer", raw: "&amp;gt;", want: "&gt;"},
		{name: "entity before text", raw: "&mdash; summary", want: "— summary"},

		// plain text with ampersands and angle brackets must stay literal
		{name: "angle bracket text untouched", raw: "std::vector<int> primer", want: "std::vector<int> primer"},
		{name: "generic type text untouched", raw: "Vec<T> and Option<T>", want: "Vec<T> and Option<T>"},
		{name: "tight brackets untouched", raw: "why a<b>c is unstable", want: "why a<b>c is unstable"},
		{name: "unclosed tag-like text untouched", raw: "read the <article", want: "read the <article"},
		{name: "comment-like text untouched", raw: "why <!-- matters", want: "why <!-- matters"},
		{name: "math comparisons untouched", raw: "score < 3, I <3 code, a<=b", want: "score < 3, I <3 code, a<=b"},
		{name: "bare ampersand words untouched", raw: "AT&T, Barnes & Noble, R&D", want: "AT&T, Barnes & Noble, R&D"},

		// HTML5 legacy no-semicolon references must never fire
		{name: "legacy micro prefix stays literal", raw: "kernels &microkernels", want: "kernels &microkernels"},
		{name: "legacy copy prefix stays literal", raw: "GPL &copyleft", want: "GPL &copyleft"},
		{name: "legacy para prefix stays literal", raw: "functions &parameters", want: "functions &parameters"},
		{name: "legacy sect prefix stays literal", raw: "see &section 3", want: "see &section 3"},
		{name: "legacy not prefix stays literal", raw: "all or &nothing", want: "all or &nothing"},
		{name: "no semicolon stays literal", raw: "&copy 2026, &#61", want: "&copy 2026, &#61"},
		{name: "ampersand without candidate stays literal", raw: "a & b; c", want: "a & b; c"},

		// tag-like input is not stripped: single-layer markup keeps showing as
		// characters, exactly like it did before this normalization existed
		{name: "single-layer markup untouched", raw: "Hello<br>World", want: "Hello<br>World"},
		{name: "script-like text untouched", raw: "Hello <script>alert(1)</script>", want: "Hello <script>alert(1)</script>"},

		{name: "whitespace trimmed", raw: "  trimmed  ", want: "trimmed"},
		{name: "empty", raw: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeTitle(tt.raw); got != tt.want {
				t.Fatalf("NormalizeTitle(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestNormalizeTitleIdempotent(t *testing.T) {
	// Idempotence holds for ordinary input but NOT for recombination cases:
	// "&amp;gt;" -> "&gt;" -> ">" (decoding &amp; re-synthesizes a new
	// candidate). That mirrors every entity decoder including browsers. The
	// contract is exactly one application at ingest; any future backfill must
	// also be single-pass, never loop-until-stable.
	inputs := []string{
		"Belief =&gt; Actions =&gt; Results",
		"A &amp; B",
		"Use &lt;b&gt; for bold",
		"std::vector<int> primer",
		"kernels &microkernels",
	}
	for _, raw := range inputs {
		once := NormalizeTitle(raw)
		if twice := NormalizeTitle(once); twice != once {
			t.Fatalf("NormalizeTitle not idempotent: %q -> %q -> %q", raw, once, twice)
		}
	}
}

func TestMapItemFallbackGUIDHashesRawTitle(t *testing.T) {
	now := time.Now()
	item := &gofeed.Item{
		Title:           "Belief =&gt; Actions =&gt; Results",
		PublishedParsed: &now,
	}

	first := mapItem(item, nil)
	second := mapItem(item, nil)

	if first.GUID != second.GUID {
		t.Fatalf("fallback GUID must be stable across pulls, got %q and %q", first.GUID, second.GUID)
	}
	if first.Title != "Belief => Actions => Results" {
		t.Fatalf("stored title = %q, want normalized form", first.Title)
	}
	if first.GUID != fallbackGUID(item.Title, item.Content, now.Unix(), true) {
		t.Fatal("fallback GUID must hash the raw title, not the normalized one")
	}
}

// TestFetchAndParseNormalizesHTMLTitles uses the exact byte shapes that make
// gofeed hand titles over with one entity layer still encoded: Atom
// type="html" with double-escaped references (the #97 report) and RSS titles
// wrapped in CDATA. Single-escaped plain titles must keep passing through
// unchanged so an identity NormalizeTitle cannot satisfy this test.
func TestFetchAndParseNormalizesHTMLTitles(t *testing.T) {
	feedXML := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Example &amp;amp; Blog</title>
  <entry>
    <title type="html">Belief =&amp;gt; Actions =&amp;gt; Results</title>
    <id>urn:entry-1</id>
    <updated>2026-01-01T00:00:00Z</updated>
  </entry>
  <entry>
    <title>Plain std::vector&lt;int&gt; primer</title>
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
		"Plain std::vector<int> primer",
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

func TestFetchAndParseNormalizesCDATATitles(t *testing.T) {
	feedXML := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>CDATA Demo</title>
    <item>
      <title><![CDATA[Inside CDATA: 5 &lt; 10 &amp;&amp; ok => y]]></title>
      <guid>cdata-1</guid>
      <pubDate>Thu, 01 Jan 2026 00:00:00 GMT</pubDate>
    </item>
  </channel>
</rss>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(feedXML))
	}))
	defer server.Close()

	result, err := FetchAndParse(context.Background(), &model.Feed{Link: server.URL}, 5*time.Second, true)
	if err != nil {
		t.Fatalf("FetchAndParse() failed: %v", err)
	}

	want := "Inside CDATA: 5 < 10 && ok => y"
	if len(result.Items) != 1 || result.Items[0].Title != want {
		t.Fatalf("title = %q, want %q", result.Items[0].Title, want)
	}
}
