package pull_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/0x2e/fusion/model"
	"github.com/0x2e/fusion/pkg/ptr"
	"github.com/0x2e/fusion/service/pull"
)

func TestDecideFeedUpdateAction(t *testing.T) {
	// Helper function to parse ISO8601 string to time.Time.
	parseTime := func(iso8601 string) time.Time {
		t, err := time.Parse(time.RFC3339, iso8601)
		if err != nil {
			panic(err)
		}
		return t
	}

	for _, tt := range []struct {
		description        string
		currentTime        time.Time
		feed               model.Feed
		expectedAction     pull.FeedUpdateAction
		expectedSkipReason *pull.FeedSkipReason
	}{
		{
			description: "suspended feed should skip update",
			currentTime: parseTime("2025-01-01T12:00:00Z"),
			feed: model.Feed{
				Suspended: ptr.To(true),
				UpdatedAt: parseTime("2025-01-01T12:00:00Z"),
			},
			expectedAction:     pull.ActionSkipUpdate,
			expectedSkipReason: &pull.SkipReasonSuspended,
		},
		{
			description: "failed feed should skip update",
			currentTime: parseTime("2025-01-01T12:00:00Z"),
			feed: model.Feed{
				Failure:   ptr.To("dummy previous error"),
				Suspended: ptr.To(false),
				UpdatedAt: parseTime("2025-01-01T12:00:00Z"),
			},
			expectedAction:     pull.ActionSkipUpdate,
			expectedSkipReason: &pull.SkipReasonLastUpdateFailed,
		},
		{
			description: "recently updated feed should skip update",
			currentTime: parseTime("2025-01-01T12:00:00Z"),
			feed: model.Feed{
				Failure:   ptr.To(""),
				Suspended: ptr.To(false),
				UpdatedAt: parseTime("2025-01-01T11:45:00Z"), // 15 minutes before current time
			},
			expectedAction:     pull.ActionSkipUpdate,
			expectedSkipReason: &pull.SkipReasonTooSoon,
		},
		{
			description: "feed should be updated when conditions are met",
			currentTime: parseTime("2025-01-01T12:00:00Z"),
			feed: model.Feed{
				Failure:   ptr.To(""),
				Suspended: ptr.To(false),
				UpdatedAt: parseTime("2025-01-01T11:15:00Z"), // 45 minutes before current time
			},
			expectedAction:     pull.ActionFetchUpdate,
			expectedSkipReason: nil,
		},
		{
			description: "feed with nil failure should be updated",
			currentTime: parseTime("2025-01-01T12:00:00Z"),
			feed: model.Feed{
				Failure:   nil,
				Suspended: ptr.To(false),
				UpdatedAt: parseTime("2025-01-01T11:15:00Z"), // 45 minutes before current time
			},
			expectedAction:     pull.ActionFetchUpdate,
			expectedSkipReason: nil,
		},
		{
			description: "feed with nil suspended should be updated",
			currentTime: parseTime("2025-01-01T12:00:00Z"),
			feed: model.Feed{
				Failure:   ptr.To(""),
				Suspended: nil,
				UpdatedAt: parseTime("2025-01-01T11:15:00Z"), // 45 minutes before current time
			},
			expectedAction:     pull.ActionFetchUpdate,
			expectedSkipReason: nil,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			action, skipReason := pull.DecideFeedUpdateAction(&tt.feed, tt.currentTime)
			assert.Equal(t, tt.expectedAction, action)
			assert.Equal(t, tt.expectedSkipReason, skipReason)
		})
	}
}
