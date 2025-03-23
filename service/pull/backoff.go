package pull

import (
	"math"
	"time"
)

// maxBackoff is the maximum time to wait before checking a feed due to past
// errors.
const maxBackoff = 7 * 24 * time.Hour

// CalculateBackoffTime calculates the exponential backoff time based on the
// number of consecutive failures.
// The formula is: interval * (1.8 ^ consecutiveFailures), capped at maxBackoff.
func CalculateBackoffTime(consecutiveFailures uint) time.Duration {
	// If no failures, no backoff needed
	if consecutiveFailures == 0 {
		return 0
	}

	intervalMinutes := float64(interval.Minutes())
	backoffMinutes := intervalMinutes * math.Pow(1.8, float64(consecutiveFailures))

	if math.IsInf(backoffMinutes, 0) || backoffMinutes > maxBackoff.Minutes() {
		return maxBackoff
	}

	return time.Duration(backoffMinutes) * time.Minute
}
