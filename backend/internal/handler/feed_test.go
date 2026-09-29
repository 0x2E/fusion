package handler

import (
	"testing"

	"github.com/0x2E/feedfinder"
)

func TestNormalizeDiscoveredFeedsNormalizesTitles(t *testing.T) {
	found := []feedfinder.Feed{
		{Title: "  Blog =&gt; Notes  ", Link: "http://example.com/a.xml"},
		{Title: "AT&T Feed", Link: "http://example.com/b.xml"},
		{Title: "std::vector&lt;int&gt;", Link: "http://example.com/c.xml"},
	}

	got := normalizeDiscoveredFeeds(found)

	want := []string{
		"Blog => Notes",
		"AT&T Feed",
		"std::vector<int>",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d feeds, want %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i].Title != w {
			t.Fatalf("feed %d title = %q, want %q", i, got[i].Title, w)
		}
	}
}
