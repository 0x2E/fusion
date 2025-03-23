package pull_test

import (
	"math"
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
		{
			description: "failed feed with 1 consecutive failure should skip update before 54 minutes",
			currentTime: parseTime("2025-01-01T12:00:00Z"),
			feed: model.Feed{
				Failure:             ptr.To("dummy previous error"),
				Suspended:           ptr.To(false),
				UpdatedAt:           parseTime("2025-01-01T11:15:00Z"), // 45 minutes before current time
				ConsecutiveFailures: 1,
			},
			expectedAction:     pull.ActionSkipUpdate,
			expectedSkipReason: &pull.SkipReasonCoolingOff,
		},
		{
			description: "failed feed with 1 consecutive failure should be updated after 54 minutes",
			currentTime: parseTime("2025-01-01T12:00:00Z"),
			feed: model.Feed{
				Failure:             ptr.To("dummy previous error"),
				Suspended:           ptr.To(false),
				UpdatedAt:           parseTime("2025-01-01T11:06:00Z"), // 54 minutes before current time
				ConsecutiveFailures: 1,
			},
			expectedAction:     pull.ActionFetchUpdate,
			expectedSkipReason: nil,
		},
		{
			description: "failed feed with 3 consecutive failures should skip update for 174 minutes",
			currentTime: parseTime("2025-01-01T12:00:00Z"),
			feed: model.Feed{
				Failure:             ptr.To("dummy previous error"),
				Suspended:           ptr.To(false),
				UpdatedAt:           parseTime("2025-01-01T09:10:00Z"), // 170 minutes before current time
				ConsecutiveFailures: 3,
			},
			expectedAction:     pull.ActionSkipUpdate,
			expectedSkipReason: &pull.SkipReasonCoolingOff,
		},
		{
			description: "failed feed with 3 consecutive failures should be updated after 174 minutes",
			currentTime: parseTime("2025-01-01T12:00:00Z"),
			feed: model.Feed{
				Failure:             ptr.To("dummy previous error"),
				Suspended:           ptr.To(false),
				UpdatedAt:           parseTime("2025-01-01T09:06:00Z"), // 174 minutes before current time
				ConsecutiveFailures: 3,
			},
			expectedAction:     pull.ActionFetchUpdate,
			expectedSkipReason: nil,
		},
		{
			description: "failed feed with many consecutive failures should not exceed maximum wait time of 7 days",
			currentTime: parseTime("2025-01-01T12:00:00Z"),
			feed: model.Feed{
				Failure:             ptr.To("dummy previous error"),
				Suspended:           ptr.To(false),
				UpdatedAt:           parseTime("2024-12-30T12:00:00Z"), // 2 days before current time
				ConsecutiveFailures: 10,
			},
			expectedAction:     pull.ActionSkipUpdate,
			expectedSkipReason: &pull.SkipReasonCoolingOff,
		},
		{
			description: "failed feed with many consecutive failures should be updated after maximum wait time of 7 days",
			currentTime: parseTime("2025-01-01T12:00:00Z"),
			feed: model.Feed{
				Failure:             ptr.To("dummy previous error"),
				Suspended:           ptr.To(false),
				UpdatedAt:           parseTime("2024-12-25T12:00:00Z"), // 7 days before current time
				ConsecutiveFailures: math.MaxUint,
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
