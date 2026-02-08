package pull

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
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

	for _, feed := range feeds {
		if ShouldSkip(feed, p.interval, p.maxBackoff) {
			continue
		}

		// Acquire semaphore slot (blocks if at capacity)
		if err := p.concurrency.Acquire(ctx, 1); err != nil {
			return // Context cancelled
		}

		go func(f *model.Feed) {
			defer p.concurrency.Release(1)
			p.pullFeed(ctx, f)
		}(feed)
	}
}

// pullFeed fetches single feed and saves new items.
func (p *Puller) pullFeed(ctx context.Context, feed *model.Feed) {
	p.logger.Debug("pulling feed", "feed_id", feed.ID, "feed_name", feed.Name)

	items, siteURL, err := FetchAndParse(ctx, feed, p.timeout)
	if err != nil {
		p.logger.Warn("failed to fetch feed", "feed_id", feed.ID, "feed_name", feed.Name, "error", err)
		if err := p.store.UpdateFeedFailure(feed.ID, err.Error()); err != nil {
			p.logger.Error("failed to record failure", "feed_id", feed.ID, "error", err)
		}
		return
	}

	newCount := 0
	for _, item := range items {
		exists, err := p.store.ItemExists(feed.ID, item.GUID)
		if err != nil {
			p.logger.Error("failed to check item existence", "feed_id", feed.ID, "error", err)
			continue
		}
		if exists {
			continue
		}

		_, err = p.store.CreateItem(feed.ID, item.GUID, item.Title, item.Link, item.Content, item.PubDate)
		if err != nil {
			p.logger.Error("failed to create item", "feed_id", feed.ID, "error", err)
			continue
		}
		newCount++
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
