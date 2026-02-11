package pull

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

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
