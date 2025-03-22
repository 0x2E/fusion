package pull

import (
	"context"
	"time"

	"github.com/0x2e/fusion/model"
	"github.com/0x2e/fusion/pkg/ptr"
	"github.com/0x2e/fusion/service/pull/client"
)

// ReadFeedItemsFn is responsible for reading a feed from an HTTP server and
// converting the result to fusion-native data types. The error return value
// is for request errors (e.g. HTTP errors).
type ReadFeedItemsFn func(ctx context.Context, feedURL string, options model.FeedRequestOptions) (client.FetchItemsResult, error)

// UpdateFeedInStoreFn is responsible for saving the result of a feed fetch to a data
// store. If the fetch failed, it records that in the data store. If the fetch
// succeeds, it stores the latest build time in the data store and adds any new
// feed items to the datastore.
type UpdateFeedInStoreFn func(feedID uint, items []*model.Item, lastBuild *time.Time, requestError error) error

// SingleFeedRepo represents a datastore for storing information about a feed.
type SingleFeedRepo interface {
	InsertItems(items []*model.Item) error
	RecordSuccess(lastBuild *time.Time) error
	RecordFailure(readErr error) error
}

type SingleFeedPuller struct {
	readFeed ReadFeedItemsFn
	repo     SingleFeedRepo
}

// NewSingleFeedPuller creates a new SingleFeedPuller with the given ReadFeedItemsFn and repository.
func NewSingleFeedPuller(readFeed ReadFeedItemsFn, repo SingleFeedRepo) SingleFeedPuller {
	return SingleFeedPuller{
		readFeed: readFeed,
		repo:     repo,
	}
}

// defaultSingleFeedRepo is the default implementation of SingleFeedRepo
type defaultSingleFeedRepo struct {
	feedID   uint
	feedRepo FeedRepo
	itemRepo ItemRepo
}

func (r *defaultSingleFeedRepo) InsertItems(items []*model.Item) error {
	// Set the correct feed ID for all items.
	for _, item := range items {
		item.FeedID = r.feedID
	}
	return r.itemRepo.Insert(items)
}

func (r *defaultSingleFeedRepo) RecordSuccess(lastBuild *time.Time) error {
	return r.feedRepo.Update(r.feedID, &model.Feed{
		LastBuild: lastBuild,
		Failure:   ptr.To(""),
	})
}

func (r *defaultSingleFeedRepo) RecordFailure(readErr error) error {
	return r.feedRepo.Update(r.feedID, &model.Feed{
		Failure: ptr.To(readErr.Error()),
	})
}

func (p SingleFeedPuller) Pull(ctx context.Context, feed *model.Feed) error {
	logger := pullLogger.With("feed_id", feed.ID, "feed_name", feed.Name)

	// We don't exit on error, as we want to record any error in the data store.
	fetchResult, readErr := p.readFeed(ctx, *feed.Link, feed.FeedRequestOptions)
	if readErr == nil {
		logger.Infof("fetched %d items", len(fetchResult.Items))
	} else {
		logger.Infof("fetch failed: %v", readErr)
	}

	return p.updateFeedInStore(feed.ID, fetchResult.Items, fetchResult.LastBuild, readErr)
}

// updateFeedInStore saves the result of a feed fetch to the data store.
// If the fetch failed, it records that in the data store.
// If the fetch succeeds, it stores the latest build time and adds any new feed items.
func (p SingleFeedPuller) updateFeedInStore(feedID uint, items []*model.Item, lastBuild *time.Time, requestError error) error {
	if requestError != nil {
		return p.repo.RecordFailure(requestError)
	}

	if err := p.repo.InsertItems(items); err != nil {
		return err
	}

	return p.repo.RecordSuccess(lastBuild)
}
