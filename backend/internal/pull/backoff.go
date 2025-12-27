package pull

import (
	"math"
	"time"

	"github.com/0x2E/fusion/internal/model"
)

const backoffBase = 1.8

// CalculateBackoff computes exponential backoff duration.
// Formula: interval × (1.8 ^ failures), capped at maxBackoff.
func CalculateBackoff(interval time.Duration, failures int64, maxBackoff time.Duration) time.Duration {
	if failures == 0 {
		return 0
	}

	backoff := float64(interval) * math.Pow(backoffBase, float64(failures))
	duration := time.Duration(backoff)

	if duration > maxBackoff {
		return maxBackoff
	}
	return duration
}

// ShouldSkip determines if feed fetch should be skipped based on backoff.
// Returns true if feed is suspended or still in backoff period.
func ShouldSkip(feed *model.Feed, interval, maxBackoff time.Duration) bool {
	now := time.Now().Unix()

	// Skip if suspended
	if feed.Suspended {
		return true
	}

	// Skip if in backoff period
	if feed.Failures > 0 {
		backoff := CalculateBackoff(interval, feed.Failures, maxBackoff)
		nextPull := feed.LastBuild + int64(backoff.Seconds())
		if now < nextPull {
			return true
		}
	}

	// Skip if recently updated (within interval)
	if now-feed.LastBuild < int64(interval.Seconds()) {
		return true
	}

	return false
}
