package sniff

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testParseHTMLContentItem struct {
	content []byte
	want    []FeedLink
}

func TestParseHTMLContentMatchLink(t *testing.T) {
	table := []testParseHTMLContentItem{
		{content: []byte(`
		<html>
		<head>
			<title>html title</title>
			<link type="application/rss+xml" title="feed title" href="https://example.com/x/rss.xml">
			<link type="application/atom+xml" href="https://example.com/x/atom.xml">
		</head>
		<body>
			<link type="application/feed+json" title="link in body" href="https://example.com/x/feed.json">
		</body>
		</html>
		`), want: []FeedLink{
			{Title: "feed title", Link: "https://example.com/x/rss.xml"},
			{Title: "html title", Link: "https://example.com/x/atom.xml"},
		}},
	}

	for _, tt := range table {
		sniffer := Sniffer{}
		feed, err := sniffer.parseHTMLContent(context.Background(), tt.content)
		assert.Nil(t, err)
		assert.ElementsMatch(t, tt.want, feed)
	}
}

func TestParseHTMLContentMatchLinkElement(t *testing.T) {
	table := []testParseHTMLContentItem{
		// match <a>
		{content: []byte(`
		<html>
		<head><title>html title</title></head>
		<body>
			<p>xxx</p>
			<main>
				<p>xxx</p>
                <a href="https://github.com/golang/go/releases.atom">RSS: Release notes from go</a>
			</main>
			<footer>
				<a href="https://github.com/golang/go">wrong rss</a>
			</footer>
		</body>
		</html>
		`), want: []FeedLink{
			{Title: "Release notes from go", Link: "https://github.com/golang/go/releases.atom"},
		}},
	}

	for _, tt := range table {
		sniffer := Sniffer{httpClient: newClient()}
		feed, err := sniffer.parseHTMLContent(context.Background(), tt.content)
		assert.Nil(t, err)
		assert.ElementsMatch(t, tt.want, feed)
	}
}
