package pullpolicy

import (
	"math"
	"strconv"
	"strings"
	"time"
)

const backoffBase = 1.8

type FeedRuntimeState struct {
	Suspended           bool
	RetryAfterUntil     int64
	NextCheckAt         int64
	ConsecutiveFailures int64
	LastErrorAt         int64
	LastCheckedAt       int64
}

func ShouldSkip(now int64, state FeedRuntimeState, interval, maxBackoff time.Duration) bool {
	if state.Suspended {
		return true
	}

	if state.RetryAfterUntil > now {
		return true
	}

	if state.NextCheckAt > now {
		return true
	}

	if state.NextCheckAt > 0 {
		return false
	}

	if state.ConsecutiveFailures > 0 {
		backoff := CalculateBackoff(interval, state.ConsecutiveFailures, maxBackoff)
		base := state.LastErrorAt
		if base <= 0 {
			base = state.LastCheckedAt
		}
		nextPull := base + int64(backoff.Seconds())
		if now < nextPull {
			return true
		}
	}

	if now-state.LastCheckedAt < int64(interval.Seconds()) {
		return true
	}

	return false
}

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

func ComputeNextCheckAt(
	now int64,
	interval, maxBackoff time.Duration,
	consecutiveFailures int64,
	retryAfterUntil int64,
	cacheControl string,
	expiresAt int64,
) int64 {
	return ComputeNextCheckAtSeconds(
		now,
		int64(interval.Seconds()),
		int64(maxBackoff.Seconds()),
		consecutiveFailures,
		retryAfterUntil,
		cacheControl,
		expiresAt,
	)
}

func ComputeNextCheckAtSeconds(
	now int64,
	intervalSeconds int64,
	maxBackoffSeconds int64,
	consecutiveFailures int64,
	retryAfterUntil int64,
	cacheControl string,
	expiresAt int64,
) int64 {
	if intervalSeconds <= 0 {
		intervalSeconds = 1
	}
	if maxBackoffSeconds <= 0 {
		maxBackoffSeconds = intervalSeconds
	}

	branchDelay := intervalSeconds

	if retryAfterDelay := retryAfterUntil - now; retryAfterDelay > branchDelay {
		branchDelay = retryAfterDelay
	}

	if cacheMaxAge := parseCacheControlMaxAgeSeconds(cacheControl); cacheMaxAge > 0 {
		if cacheMaxAge > branchDelay {
			branchDelay = cacheMaxAge
		}
	}

	if expiresDelay := expiresAt - now; expiresDelay > branchDelay {
		branchDelay = expiresDelay
	}

	if consecutiveFailures > 0 {
		backoffSeconds := calculateBackoffSeconds(intervalSeconds, consecutiveFailures, maxBackoffSeconds)
		if backoffSeconds > branchDelay {
			branchDelay = backoffSeconds
		}
	}

	if branchDelay < 0 {
		branchDelay = 0
	}

	if branchDelay > maxBackoffSeconds {
		branchDelay = maxBackoffSeconds
	}

	return now + branchDelay
}

func calculateBackoffSeconds(intervalSeconds, failures, maxBackoffSeconds int64) int64 {
	if failures <= 0 {
		return 0
	}

	backoff := float64(intervalSeconds) * math.Pow(backoffBase, float64(failures))
	seconds := int64(backoff)
	if seconds > maxBackoffSeconds {
		return maxBackoffSeconds
	}

	return seconds
}

func parseCacheControlMaxAgeSeconds(cacheControl string) int64 {
	if strings.TrimSpace(cacheControl) == "" {
		return 0
	}

	for _, part := range strings.Split(cacheControl, ",") {
		token := strings.TrimSpace(strings.ToLower(part))
		if !strings.HasPrefix(token, "max-age=") {
			continue
		}

		raw := strings.TrimSpace(strings.TrimPrefix(token, "max-age="))
		seconds, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || seconds <= 0 {
			return 0
		}

		return seconds
	}

	return 0
}
