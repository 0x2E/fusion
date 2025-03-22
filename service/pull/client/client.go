package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/0x2e/fusion/model"
	"github.com/0x2e/fusion/pkg/httpx"
)

type HttpRequestFn func(ctx context.Context, link string, options model.FeedRequestOptions) (*http.Response, error)

// FeedClient retrieves a feed given a feed URL and parses the result.
type FeedClient struct {
	httpRequestFn HttpRequestFn
}

// NewFeedClient creates a feed client with the default options.
func NewFeedClient() FeedClient {
	return NewFeedClientWithRequestFn(httpx.FusionRequest)
}

// NewFeedClientWithRequestFn creates a feed client that uses a custom
// HttpRequestFn to retrieve remote feeds.
func NewFeedClientWithRequestFn(httpRequestFn HttpRequestFn) FeedClient {
	return FeedClient{
		httpRequestFn: httpRequestFn,
	}
}

func (c FeedClient) FetchTitle(ctx context.Context, feedURL string, options model.FeedRequestOptions) (string, error) {
	feed, err := c.fetchFeed(ctx, feedURL, options)
	if err != nil {
		return "", err
	}

	return feed.Title, nil
}

// FetchDeclaredLink retrieves the feed link declared within the feed content
func (c FeedClient) FetchDeclaredLink(ctx context.Context, feedURL string, options model.FeedRequestOptions) (string, error) {
	feed, err := c.fetchFeed(ctx, feedURL, options)
	if err != nil {
		return "", err
	}

	if feed.FeedLink != "" {
		return feed.FeedLink, nil
	}

	return feed.Link, nil
}

type FetchItemsResult struct {
	LastBuild *time.Time
	Items     []*model.Item
}

func (c FeedClient) FetchItems(ctx context.Context, feedURL string, options model.FeedRequestOptions) (FetchItemsResult, error) {
	feed, err := c.fetchFeed(ctx, feedURL, options)
	if err != nil {
		return FetchItemsResult{}, err
	}

	return FetchItemsResult{
		LastBuild: feed.UpdatedParsed,
		Items:     ParseGoFeedItems(feed.Items),
	}, nil
}

func (c FeedClient) fetchFeed(ctx context.Context, feedURL string, options model.FeedRequestOptions) (*gofeed.Feed, error) {
	resp, err := c.httpRequestFn(ctx, feedURL, options)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got status code %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return gofeed.NewParser().ParseString(string(data))
}
