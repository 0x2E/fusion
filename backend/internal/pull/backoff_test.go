package pull

import (
	"testing"
	"time"

	"github.com/0x2E/fusion/internal/model"
)

func TestCalculateBackoff(t *testing.T) {
	interval := 30 * time.Minute
	maxBackoff := 7 * 24 * time.Hour

	tests := []struct {
		name       string
		failures   int64
		wantApprox time.Duration // Approximate expected value
	}{
		{
			name:       "no failures",
			failures:   0,
			wantApprox: 0,
		},
		{
			name:       "1 failure",
			failures:   1,
			wantApprox: 54 * time.Minute, // 30 * 1.8 = 54
		},
		{
			name:       "2 failures",
			failures:   2,
			wantApprox: 97 * time.Minute, // 30 * 1.8^2 = 97.2
		},
		{
			name:       "3 failures",
			failures:   3,
			wantApprox: 175 * time.Minute, // 30 * 1.8^3 = 174.96
		},
		{
			name:       "5 failures",
			failures:   5,
			wantApprox: 9*time.Hour + 30*time.Minute, // 30 * 1.8^5 = 567.65 min ≈ 9.5h
		},
		{
			name:       "100 failures (exceeds max)",
			failures:   100,
			wantApprox: maxBackoff,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateBackoff(interval, tt.failures, maxBackoff)

			if tt.failures == 0 {
				if got != 0 {
					t.Errorf("CalculateBackoff() = %v, want 0", got)
				}
				return
			}

			if tt.failures == 100 {
				if got != maxBackoff {
					t.Errorf("CalculateBackoff() = %v, want max backoff %v", got, maxBackoff)
				}
				return
			}

			// Allow 10% tolerance for floating point calculations
			tolerance := float64(tt.wantApprox) * 0.1
			diff := float64(got - tt.wantApprox)
			if diff < 0 {
				diff = -diff
			}

			if diff > tolerance {
				t.Errorf("CalculateBackoff() = %v, want approximately %v (tolerance: %.0f)", got, tt.wantApprox, tolerance)
			}
		})
	}
}

func TestShouldSkip(t *testing.T) {
	interval := 30 * time.Minute
	maxBackoff := 7 * 24 * time.Hour
	now := time.Now().Unix()

	tests := []struct {
		name string
		feed *model.Feed
		want bool
	}{
		{
			name: "suspended feed",
			feed: &model.Feed{
				Suspended: true,
				LastBuild: now - 3600,
				Failures:  0,
			},
			want: true,
		},
		{
			name: "recently updated (10 min ago)",
			feed: &model.Feed{
				Suspended: false,
				LastBuild: now - 600, // 10 minutes ago
				Failures:  0,
			},
			want: true,
		},
		{
			name: "ready to pull (40 min ago)",
			feed: &model.Feed{
				Suspended: false,
				LastBuild: now - 2400, // 40 minutes ago
				Failures:  0,
			},
			want: false,
		},
		{
			name: "1 failure, in backoff period",
			feed: &model.Feed{
				Suspended: false,
				LastBuild: now - 1800, // 30 minutes ago (backoff is 54 min)
				Failures:  1,
			},
			want: true,
		},
		{
			name: "1 failure, backoff expired",
			feed: &model.Feed{
				Suspended: false,
				LastBuild: now - 3600, // 60 minutes ago (backoff is 54 min)
				Failures:  1,
			},
			want: false,
		},
		{
			name: "never pulled before",
			feed: &model.Feed{
				Suspended: false,
				LastBuild: 0,
				Failures:  0,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldSkip(tt.feed, interval, maxBackoff)
			if got != tt.want {
				t.Errorf("ShouldSkip() = %v, want %v", got, tt.want)
			}
		})
	}
}
