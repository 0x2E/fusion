package pull

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0x2E/fusion/internal/config"
	"github.com/0x2E/fusion/internal/model"
	"github.com/0x2E/fusion/internal/store"
	"golang.org/x/sync/semaphore"
)

type Puller struct {
	store       *store.Store
	config      *config.Config
	interval    time.Duration
	timeout     time.Duration
	maxBackoff  time.Duration
	concurrency *semaphore.Weighted
}

func New(st *store.Store, cfg *config.Config) *Puller {
	return &Puller{
		store:       st,
		config:      cfg,
		interval:    time.Duration(cfg.PullInterval) * time.Second,
		timeout:     time.Duration(cfg.PullTimeout) * time.Second,
		maxBackoff:  time.Duration(cfg.PullMaxBackoff) * time.Second,
		concurrency: semaphore.NewWeighted(int64(cfg.PullConcurrency)),
	}
}

// Start begins periodic feed pulling. Blocks until context is cancelled.
func (p *Puller) Start(ctx context.Context) error {
	log.Printf("pull service started (interval: %v, timeout: %v, concurrency: %d)",
		p.interval, p.timeout, p.config.PullConcurrency)

	// Run immediately on startup
	p.pullAll(ctx)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("pull service stopping...")
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
		log.Printf("failed to list feeds: %v", err)
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
	log.Printf("pulling feed: %s (id=%d)", feed.Name, feed.ID)

	items, err := FetchAndParse(ctx, feed, p.timeout)
	if err != nil {
		log.Printf("failed to fetch feed %s: %v", feed.Name, err)
		if err := p.store.UpdateFeedFailure(feed.ID, err.Error()); err != nil {
			log.Printf("failed to record failure: %v", err)
		}
		return
	}

	newCount := 0
	for _, item := range items {
		exists, err := p.store.ItemExists(feed.ID, item.GUID)
		if err != nil {
			log.Printf("failed to check item existence: %v", err)
			continue
		}
		if exists {
			continue
		}

		_, err = p.store.CreateItem(feed.ID, item.GUID, item.Title, item.Link, item.Content, item.PubDate)
		if err != nil {
			log.Printf("failed to create item: %v", err)
			continue
		}
		newCount++
	}

	if err := p.store.UpdateFeedLastBuild(feed.ID, time.Now().Unix()); err != nil {
		log.Printf("failed to update last_build: %v", err)
		return
	}

	log.Printf("feed %s pulled successfully (%d new items)", feed.Name, newCount)
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
