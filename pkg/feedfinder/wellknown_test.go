package feedfinder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRSSContent(t *testing.T) {
	type testItem struct {
		content []byte
		want    FeedLink
	}

	// TODO: match all types, e.g. https://github.com/mmcdole/gofeed/tree/master/testdata
	table := []testItem{
		{content: []byte(`
		<?xml version="1.0" encoding="utf-8"?>
		<rss xmlns:atom="http://www.w3.org/2005/Atom" xmlns:content="http://purl.org/rss/1.0/modules/content/" version="2.0">  
		  <channel> 
			<title>test</title>  
			<link>https://example.com/</link>  
			<language>en</language>
			<lastBuildDate>Fri, 24 Feb 2023 00:43:57 +0800</lastBuildDate>  
			<atom:link href="https://example.com/feed.xml" rel="self" type="application/rss+xml"/>  
			<item> 
			  <title>post1</title>  
			  <link>https://example.com/post1/</link>  
			  <pubDate>Fri, 24 Feb 2023 00:43:57 +0800</pubDate>  
			  <guid>https://example.com/post1/</guid>  
			</item>  
		  </channel> 
		</rss>
		`), want: FeedLink{Title: "test", Link: "https://example.com/feed.xml"}},
	}

	for _, tt := range table {
		feed, err := parseRSSContent(tt.content)
		assert.Nil(t, err)
		assert.Equal(t, tt.want, feed)
	}
}
