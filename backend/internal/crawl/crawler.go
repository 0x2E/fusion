package crawl

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	readability "codeberg.org/readeck/go-readability/v2"
	"github.com/0x2E/fusion/internal/model"
	"github.com/0x2E/fusion/internal/pkg/httpc"
	"github.com/0x2E/fusion/internal/store"
	"golang.org/x/sync/semaphore"
)

const (
	// maxItemsPerRun caps how many uncrawled items are processed per feed per pull cycle.
	// Prevents bursts when crawl is first enabled on a feed with many historical items.
	maxItemsPerRun = 20

	// crawlConcurrency is the number of items crawled in parallel within a single feed.
	crawlConcurrency = 3
)

// Crawler fetches article content for feed items whose content is thin or missing.
type Crawler struct {
	store        *store.Store
	timeout      time.Duration
	allowPrivate bool
	logger       *slog.Logger
}

func New(st *store.Store, timeout time.Duration, allowPrivate bool) *Crawler {
	return &Crawler{
		store:        st,
		timeout:      timeout,
		allowPrivate: allowPrivate,
		logger:       slog.Default(),
	}
}

// CrawlFeedItems fetches and extracts content for uncrawled items belonging to feed.
// It is a no-op if feed.Crawl is false.
func (c *Crawler) CrawlFeedItems(ctx context.Context, feed *model.Feed) error {
	if !feed.Crawl {
		return nil
	}

	items, err := c.store.ListUncrawledItems(feed.ID, maxItemsPerRun)
	if err != nil {
		return fmt.Errorf("list uncrawled items: %w", err)
	}
	if len(items) == 0 {
		return nil
	}

	sem := semaphore.NewWeighted(crawlConcurrency)
	var wg sync.WaitGroup

	for _, item := range items {
		if err := sem.Acquire(ctx, 1); err != nil {
			break
		}
		wg.Add(1)
		go func(i *model.Item) {
			defer wg.Done()
			defer sem.Release(1)
			c.crawlItem(ctx, i, feed.Proxy)
		}(item)
	}

	wg.Wait()
	return nil
}

func (c *Crawler) crawlItem(ctx context.Context, item *model.Item, proxy string) {
	now := time.Now().Unix()

	content, err := c.fetchContent(ctx, item.Link, proxy)
	if err != nil {
		c.logger.Warn("failed to crawl item", "item_id", item.ID, "link", item.Link, "error", err)
		if err := c.store.UpdateItemCrawled(item.ID, item.Content, now); err != nil {
			c.logger.Error("failed to record crawl attempt", "item_id", item.ID, "error", err)
		}
		return
	}

	if err := c.store.UpdateItemCrawled(item.ID, content, now); err != nil {
		c.logger.Error("failed to save crawled content", "item_id", item.ID, "error", err)
	}
}

func (c *Crawler) fetchContent(ctx context.Context, rawURL, proxy string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}

	client, err := httpc.NewClient(c.timeout, proxy, c.allowPrivate)
	if err != nil {
		return "", fmt.Errorf("create http client: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpc.SetDefaultHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("http status %d", resp.StatusCode)
	}

	parser := readability.NewParser()
	article, err := parser.Parse(resp.Body, parsedURL)
	if err != nil {
		return "", fmt.Errorf("readability: %w", err)
	}

	if article.Node == nil {
		return "", fmt.Errorf("readability: no content extracted")
	}

	var buf bytes.Buffer
	if err := article.RenderHTML(&buf); err != nil {
		return "", fmt.Errorf("readability render: %w", err)
	}

	return buf.String(), nil
}
