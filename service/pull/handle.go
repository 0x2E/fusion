package pull

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/0x2e/fusion/model"
	"github.com/0x2e/fusion/pkg/ptr"
	"github.com/0x2e/fusion/service/pull/client"
)

func (p *Puller) do(ctx context.Context, f *model.Feed, force bool) error {
	logger := slog.With("feed_id", f.ID, "feed_link", ptr.From(f.Link))
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	updateAction, skipReason := DecideFeedUpdateAction(f, time.Now())
	if skipReason == &SkipReasonSuspended {
		logger.Info(fmt.Sprintf("skip: %s", skipReason))
		return nil
	}
	if !force {
		switch updateAction {
		case ActionSkipUpdate:
			logger.Info(fmt.Sprintf("skip: %s", skipReason))
			return nil
		case ActionFetchUpdate:
			// Proceed to perform the fetch.
		default:
			panic("unexpected FeedUpdateAction")
		}
	}

	repo := defaultSingleFeedRepo{
		feedID:   f.ID,
		feedRepo: p.feedRepo,
		itemRepo: p.itemRepo,
	}
	return NewSingleFeedPuller(client.NewFeedClient().FetchItems, &repo).Pull(ctx, f)
}

// FeedUpdateAction represents the action to take when considering checking a
// feed for updates.
type FeedUpdateAction uint8

const (
	ActionFetchUpdate FeedUpdateAction = iota
	ActionSkipUpdate
)

// FeedSkipReason represents a reason for skipping a feed update.
type FeedSkipReason struct {
	reason string
}

func (r FeedSkipReason) String() string {
	return r.reason
}

var (
	SkipReasonSuspended  = FeedSkipReason{"user suspended feed updates"}
	SkipReasonCoolingOff = FeedSkipReason{"slowing down requests due to past failures to update feed"}
	SkipReasonTooSoon    = FeedSkipReason{"feed was updated too recently"}
)

func DecideFeedUpdateAction(f *model.Feed, now time.Time) (FeedUpdateAction, *FeedSkipReason) {
	if f.IsSuspended() {
		return ActionSkipUpdate, &SkipReasonSuspended
	} else if f.ConsecutiveFailures > 0 {
		backoffTime := CalculateBackoffTime(f.ConsecutiveFailures)
		timeSinceUpdate := now.Sub(f.UpdatedAt)
		if timeSinceUpdate < backoffTime {
			slog.Info(fmt.Sprintf("%d consecutive feed update failures, so next attempt is after %v", f.ConsecutiveFailures, f.UpdatedAt.Add(backoffTime).Format(time.RFC3339)), "feed_id", f.ID, "feed_link", ptr.From(f.Link))
			return ActionSkipUpdate, &SkipReasonCoolingOff
		}
	} else if now.Sub(f.UpdatedAt) < interval {
		return ActionSkipUpdate, &SkipReasonTooSoon
	}
	return ActionFetchUpdate, nil
}
