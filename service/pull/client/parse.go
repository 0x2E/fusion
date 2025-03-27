package client

import (
	"net/url"
	"strings"

	"github.com/0x2e/fusion/model"
	"github.com/0x2e/fusion/pkg/ptr"

	"github.com/mmcdole/gofeed"
)

func ParseGoFeedItems(feedURL string, gfItems []*gofeed.Item) []*model.Item {
	items := make([]*model.Item, 0, len(gfItems))
	for _, item := range gfItems {
		if item == nil {
			continue
		}

		unread := true
		content := item.Content
		if content == "" {
			content = item.Description
		}
		guid := item.GUID
		if guid == "" {
			guid = item.Link
		}
		items = append(items, &model.Item{
			Title:   &item.Title,
			GUID:    &guid,
			Link:    ptr.To(parseLink(feedURL, item.Link)),
			Content: &content,
			PubDate: item.PublishedParsed,
			Unread:  &unread,
		})
	}

	return items
}

func parseLink(feedURL string, linkURL string) string {
	// If the link URL is not a relative path, treat it as a full URL.
	if !strings.HasPrefix(linkURL, "/") {
		return linkURL
	}

	baseURL, err := url.Parse(feedURL)
	// If we can't parse the feed URL, we can't repair a relative path, so just
	// return whatever is in the link URL.
	if err != nil {
		return linkURL
	}

	// Reduce the feed URL to just the scheme and hostname.
	base := url.URL{
		Scheme: baseURL.Scheme,
		Host:   baseURL.Host,
	}

	// Combine the feed base URL with the relative path to create a full URL.
	return base.String() + linkURL
}
