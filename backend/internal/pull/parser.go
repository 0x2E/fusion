package pull

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/0x2E/fusion/internal/model"
	"github.com/0x2E/fusion/internal/pkg/httpc"
	"github.com/mmcdole/gofeed"
)

// ParsedItem represents a feed item after parsing and field mapping.
type ParsedItem struct {
	GUID    string
	Title   string
	Link    string
	Content string
	PubDate int64
}

// FetchAndParse fetches RSS/Atom feed and parses into items.
// Returns parsed items or error if fetch/parse fails.
func FetchAndParse(ctx context.Context, feed *model.Feed, timeout time.Duration) ([]*ParsedItem, error) {
	client, err := httpc.NewClient(timeout, feed.Proxy)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", feed.Link, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpc.SetDefaultHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	fp := gofeed.NewParser()
	parsedFeed, err := fp.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse feed: %w", err)
	}

	baseURL, _ := url.Parse(feed.SiteURL)
	if baseURL == nil || baseURL.String() == "" {
		baseURL, _ = url.Parse(feed.Link)
	}

	items := make([]*ParsedItem, 0, len(parsedFeed.Items))
	for _, item := range parsedFeed.Items {
		items = append(items, mapItem(item, baseURL))
	}

	return items, nil
}

// mapItem converts gofeed.Item to ParsedItem following mapping rules:
// - guid: prefer GUID, fallback to Link
// - content: prefer Content, fallback to Description
// - pub_date: prefer PublishedParsed, fallback to UpdatedParsed
// - link: convert to absolute URL
func mapItem(item *gofeed.Item, baseURL *url.URL) *ParsedItem {
	guid := item.GUID
	if guid == "" {
		guid = item.Link
	}

	content := item.Content
	if content == "" {
		content = item.Description
	}

	var pubDate int64
	if item.PublishedParsed != nil {
		pubDate = item.PublishedParsed.Unix()
	} else if item.UpdatedParsed != nil {
		pubDate = item.UpdatedParsed.Unix()
	} else {
		pubDate = time.Now().Unix()
	}

	link := item.Link
	if baseURL != nil {
		if absURL, err := baseURL.Parse(link); err == nil {
			link = absURL.String()
		}
	}

	return &ParsedItem{
		GUID:    guid,
		Title:   item.Title,
		Link:    link,
		Content: content,
		PubDate: pubDate,
	}
}
