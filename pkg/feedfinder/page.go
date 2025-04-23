package feedfinder

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func (f *Finder) tryPageSource(ctx context.Context) ([]FeedLink, error) {
	resp, err := f.httpClient.Get(f.target.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status %d", resp.StatusCode)
	}

	feeds, err := f.parseHTMLContent(ctx, content)
	if err != nil {
		slog.Error(err.Error(), "content_type", "HTML")
	}
	if len(feeds) != 0 {
		for i := range feeds {
			feed := &feeds[i]
			feed.Link = formatLinkToAbs(f.target.String(), feed.Link)
		}
		return feeds, nil
	}

	feed, err := parseRSSContent(content)
	if err != nil {
		slog.Error(err.Error(), "content_type", "RSS")
	}
	if !isEmptyFeedLink(feed) {
		if feed.Link == "" {
			feed.Link = f.target.String()
		}
		return []FeedLink{feed}, nil
	}

	return nil, nil
}

func (f *Finder) parseHTMLContent(ctx context.Context, content []byte) ([]FeedLink, error) {
	feeds := make([]FeedLink, 0)

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}

	pageTitle := doc.FindMatcher(goquery.Single("title")).Text()

	// find <link> type rss in <header>
	linkExprs := []string{
		"link[type='application/rss+xml']",
		"link[type='application/atom+xml']",
		"link[type='application/json']",
		"link[type='application/feed+json']",
	}
	for _, expr := range linkExprs {
		doc.Find("head").Find(expr).Each(func(_ int, s *goquery.Selection) {
			feed := FeedLink{}
			feed.Title, _ = s.Attr("title")
			feed.Link, _ = s.Attr("href")

			if feed.Title == "" {
				feed.Title = pageTitle
			}
			feeds = append(feeds, feed)
		})
	}

	// find <a> type rss in <body>
	aExpr := "a:contains('rss')"
	suspected := make(map[string]struct{})
	doc.Find("body").Find(aExpr).Each(func(_ int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if !exists {
			return
		}
		suspected[link] = struct{}{}
	})
	for link := range suspected {
		feed, err := f.parseRSSUrl(ctx, link)
		if err != nil {
			continue
		}
		if !isEmptyFeedLink(feed) {
			feed.Link = link // this may be more accurate than the link parsed from the rss content
			feeds = append(feeds, feed)
		}
	}

	return feeds, nil
}
