package pull

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
// Returns parsed items and optional site URL discovered from feed metadata.
func FetchAndParse(ctx context.Context, feed *model.Feed, timeout time.Duration, allowPrivateFeeds bool) ([]*ParsedItem, string, error) {
	if err := httpc.ValidateRequestURL(ctx, feed.Link, allowPrivateFeeds); err != nil {
		return nil, "", fmt.Errorf("validate feed url: %w", err)
	}

	client, err := httpc.NewClient(timeout, feed.Proxy, allowPrivateFeeds)
	if err != nil {
		return nil, "", fmt.Errorf("create client: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", feed.Link, nil)
	if err != nil {
		return nil, "", fmt.Errorf("create request: %w", err)
	}
	httpc.SetDefaultHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("fetch feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	fp := gofeed.NewParser()
	parsedFeed, err := fp.Parse(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("parse feed: %w", err)
	}

	siteURL := normalizeSiteURL(parsedFeed.Link)

	baseURL, _ := url.Parse(feed.SiteURL)
	if baseURL == nil || baseURL.String() == "" {
		baseURL, _ = url.Parse(siteURL)
	}
	if baseURL == nil || baseURL.String() == "" {
		baseURL, _ = url.Parse(feed.Link)
	}

	items := make([]*ParsedItem, 0, len(parsedFeed.Items))
	for _, item := range parsedFeed.Items {
		items = append(items, mapItem(item, baseURL))
	}

	return items, siteURL, nil
}

func normalizeSiteURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed == nil {
		return ""
	}
	if parsed.Host == "" {
		return ""
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}

	parsed.Fragment = ""
	return parsed.String()
}

// mapItem converts gofeed.Item to ParsedItem following mapping rules:
// - guid: prefer GUID, fallback to Link
// - content: prefer Content, fallback to Description
// - pub_date: prefer PublishedParsed, fallback to UpdatedParsed
// - link: convert to absolute URL
func mapItem(item *gofeed.Item, baseURL *url.URL) *ParsedItem {
	content := item.Content
	if content == "" {
		content = item.Description
	}

	var sourcePubDate int64
	hasSourcePubDate := false
	var pubDate int64
	if item.PublishedParsed != nil {
		sourcePubDate = item.PublishedParsed.Unix()
		hasSourcePubDate = true
		pubDate = sourcePubDate
	} else if item.UpdatedParsed != nil {
		sourcePubDate = item.UpdatedParsed.Unix()
		hasSourcePubDate = true
		pubDate = sourcePubDate
	} else {
		pubDate = time.Now().Unix()
	}

	rawLink := strings.TrimSpace(item.Link)
	link := rawLink
	if rawLink != "" && baseURL != nil {
		if absURL, err := baseURL.Parse(rawLink); err == nil {
			link = absURL.String()
		}
	}

	guid := strings.TrimSpace(item.GUID)
	if guid == "" {
		guid = strings.TrimSpace(link)
	}
	if guid == "" {
		guid = fallbackGUID(item.Title, content, sourcePubDate, hasSourcePubDate)
	}

	return &ParsedItem{
		GUID:    guid,
		Title:   item.Title,
		Link:    link,
		Content: content,
		PubDate: pubDate,
	}
}

func fallbackGUID(title, content string, sourcePubDate int64, hasSourcePubDate bool) string {
	pubDatePart := ""
	if hasSourcePubDate {
		pubDatePart = strconv.FormatInt(sourcePubDate, 10)
	}

	h := sha256.Sum256([]byte(strings.TrimSpace(title) + "\n" + strings.TrimSpace(content) + "\n" + pubDatePart))
	return "generated:" + hex.EncodeToString(h[:])
}
