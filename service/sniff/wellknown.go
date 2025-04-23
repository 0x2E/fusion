package sniff

import (
	"bytes"
	"context"
	"io"
	"net/url"

	"github.com/mmcdole/gofeed"
)

func (s *Sniffer) tryWellKnown(ctx context.Context, baseURL string) ([]FeedLink, error) {
	wellKnown := []string{
		"atom.xml",
		"feed.xml",
		"rss.xml",
		"index.xml",
		"atom.json",
		"feed.json",
		"rss.json",
		"index.json",
		"feed/",
		"rss/",
	}
	feeds := make([]FeedLink, 0)

	for _, suffix := range wellKnown {
		newTarget, err := url.JoinPath(baseURL, suffix)
		if err != nil {
			continue
		}
		feed, err := s.parseRSSUrl(ctx, newTarget)
		if err != nil {
			continue
		}
		if !isEmptyFeedLink(feed) {
			feed.Link = newTarget // this may be more accurate than the link parsed from the rss content
			feeds = append(feeds, feed)
		}
	}

	return feeds, nil
}

func (s *Sniffer) parseRSSUrl(ctx context.Context, target string) (FeedLink, error) {
	resp, err := s.httpClient.Get(target)
	if err != nil {
		return FeedLink{}, err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return FeedLink{}, err
	}
	return parseRSSContent(content)
}

func parseRSSContent(content []byte) (FeedLink, error) {
	parsed, err := gofeed.NewParser().Parse(bytes.NewReader(content))
	if err != nil || parsed == nil {
		return FeedLink{}, err
	}
	return FeedLink{
		// https://github.com/mmcdole/gofeed#default-mappings
		Title: parsed.Title,

		// set as default value, but the value parsed from rss are not always accurate.
		// it is better to use the url that gets the rss content.
		Link: parsed.FeedLink,
	}, nil
}
