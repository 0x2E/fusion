package pull

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/0x2E/fusion/internal/config"
	"github.com/0x2E/fusion/internal/model"
	"github.com/0x2E/fusion/internal/store"
	"golang.org/x/sync/semaphore"
)

type Puller struct {
	store       *store.Store
	config      *config.Config
	logger      *slog.Logger
	interval    time.Duration
	timeout     time.Duration
	maxBackoff  time.Duration
	concurrency *semaphore.Weighted
}

func New(st *store.Store, cfg *config.Config) *Puller {
	return &Puller{
		store:       st,
		config:      cfg,
		logger:      slog.Default(),
		interval:    time.Duration(cfg.PullInterval) * time.Second,
		timeout:     time.Duration(cfg.PullTimeout) * time.Second,
		maxBackoff:  time.Duration(cfg.PullMaxBackoff) * time.Second,
		concurrency: semaphore.NewWeighted(int64(cfg.PullConcurrency)),
	}
}

// Start begins periodic feed pulling. Blocks until context is cancelled.
func (p *Puller) Start(ctx context.Context) error {
	p.logger.Info("pull service started", "interval", p.interval, "timeout", p.timeout, "concurrency", p.config.PullConcurrency)

	// Run immediately on startup
	p.pullAll(ctx)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("pull service stopping")
			return ctx.Err()
		case <-ticker.C:
			p.pullAll(ctx)
		}
	}
}

// pullAll fetches all feeds concurrently with semaphore limiting.
func (p *Puller) pullAll(ctx context.Context) {
	feeds, err := p.store.ListFeeds()
	if err != nil {
		p.logger.Error("failed to list feeds", "error", err)
		return
	}

	_, _ = p.dispatchFeeds(ctx, feeds, func(feed *model.Feed) bool {
		return !ShouldSkip(feed, p.interval, p.maxBackoff)
	})
}

// pullFeed fetches single feed and saves new items.
func (p *Puller) pullFeed(ctx context.Context, feed *model.Feed) {
	p.logger.Debug("pulling feed", "feed_id", feed.ID, "feed_name", feed.Name)

	items, siteURL, err := FetchAndParse(ctx, feed, p.timeout, p.config.AllowPrivateFeeds)
	if err != nil {
		p.logger.Warn("failed to fetch feed", "feed_id", feed.ID, "feed_name", feed.Name, "error", err)
		if err := p.store.UpdateFeedFailure(feed.ID, err.Error()); err != nil {
			p.logger.Error("failed to record failure", "feed_id", feed.ID, "error", err)
		}
		return
	}

	inputs := make([]store.BatchCreateItemInput, 0, len(items))
	for _, item := range items {
		inputs = append(inputs, store.BatchCreateItemInput{
			GUID:    item.GUID,
			Title:   item.Title,
			Link:    item.Link,
			Content: item.Content,
			PubDate: item.PubDate,
		})
	}

	newCount, err := p.store.BatchCreateItemsIgnore(feed.ID, inputs)
	if err != nil {
		p.logger.Error("failed to batch create items", "feed_id", feed.ID, "error", err)
		return
	}

	if err := p.store.UpdateFeedLastBuild(feed.ID, time.Now().Unix()); err != nil {
		p.logger.Error("failed to update last_build", "feed_id", feed.ID, "error", err)
		return
	}

	if strings.TrimSpace(feed.SiteURL) == "" && siteURL != "" {
		if err := p.store.UpdateFeedSiteURLIfEmpty(feed.ID, siteURL); err != nil {
			p.logger.Warn("failed to auto-fill site_url", "feed_id", feed.ID, "site_url", siteURL, "error", err)
		}
	}

	p.logger.Info("feed pulled successfully", "feed_id", feed.ID, "feed_name", feed.Name, "new_items", newCount)
}

// RefreshAll triggers refresh for all non-suspended feeds and waits until all
// started refresh jobs have completed. It bypasses backoff/interval skip logic.
// Concurrency is controlled by the same semaphore as periodic pulls.
func (p *Puller) RefreshAll(ctx context.Context) (int, error) {
	feeds, err := p.store.ListFeeds()
	if err != nil {
		return 0, fmt.Errorf("list feeds: %w", err)
	}

	count, err := p.dispatchFeeds(ctx, feeds, func(feed *model.Feed) bool {
		return !feed.Suspended
	})
	if err != nil {
		return count, err
	}

	return count, nil
}

func (p *Puller) dispatchFeeds(ctx context.Context, feeds []*model.Feed, shouldPull func(*model.Feed) bool) (int, error) {
	count := 0
	var wg sync.WaitGroup
	var acquireErr error

	for _, feed := range feeds {
		if !shouldPull(feed) {
			continue
		}

		if err := p.concurrency.Acquire(ctx, 1); err != nil {
			acquireErr = err
			break
		}

		count++
		wg.Add(1)
		go func(f *model.Feed) {
			defer wg.Done()
			defer p.concurrency.Release(1)
			p.pullFeed(ctx, f)
		}(feed)
	}

	wg.Wait()
	return count, acquireErr
}

// RefreshFeed manually triggers refresh for specific feed (bypasses skip logic).
// Used by HTTP handler for manual refresh requests.
func (p *Puller) RefreshFeed(ctx context.Context, feedID int64) error {
	feed, err := p.store.GetFeed(feedID)
	if err != nil {
		return fmt.Errorf("get feed: %w", err)
	}

	if err := p.concurrency.Acquire(ctx, 1); err != nil {
		return err
	}
	defer p.concurrency.Release(1)

	p.pullFeed(ctx, feed)
	return nil
}
