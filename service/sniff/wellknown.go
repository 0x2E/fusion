package sniff

import (
	"bytes"
	"context"
	"net/url"

	"github.com/mmcdole/gofeed"

	"github.com/0x2e/fusion/model"
	"github.com/0x2e/fusion/service/pull/client"
)

func tryWellKnown(ctx context.Context, baseURL string) ([]FeedLink, error) {
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
		feed, err := parseRSSUrl(ctx, newTarget)
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

func parseRSSUrl(ctx context.Context, url string) (FeedLink, error) {
	feedClient := client.NewFeedClient()

	title, err := feedClient.FetchTitle(ctx, url, model.FeedRequestOptions{})
	if err != nil {
		return FeedLink{}, err
	}

	declaredLink, err := feedClient.FetchDeclaredLink(ctx, url, model.FeedRequestOptions{})
	if err != nil {
		return FeedLink{}, err
	}

	return FeedLink{
		Title: title,
		Link:  declaredLink,
	}, nil
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
