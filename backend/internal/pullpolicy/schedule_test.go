package pullpolicy

import (
	"testing"
	"time"
)

func TestComputeNextCheckAtSecondsPicksStrictestDelay(t *testing.T) {
	now := int64(1000)
	got := ComputeNextCheckAtSeconds(
		now,
		60,
		86400,
		2,
		1120,
		"public, max-age=300",
		1250,
	)

	// max(now+interval=1060, retry_after=1120, expires=1250, backoff=1194, cache=1300)
	want := int64(1300)
	if got != want {
		t.Fatalf("ComputeNextCheckAtSeconds() = %d, want %d", got, want)
	}
}

func TestComputeNextCheckAtSecondsBackoffCapped(t *testing.T) {
	now := int64(200)
	got := ComputeNextCheckAtSeconds(
		now,
		60,
		500,
		20,
		0,
		"",
		0,
	)

	want := now + 500
	if got != want {
		t.Fatalf("ComputeNextCheckAtSeconds() = %d, want %d", got, want)
	}
}

func TestComputeNextCheckAtSecondsSuccessBranchCappedByGlobalMax(t *testing.T) {
	now := int64(1000)
	got := ComputeNextCheckAtSeconds(
		now,
		60,
		120,
		0,
		0,
		"public, max-age=600",
		0,
	)

	want := now + 120
	if got != want {
		t.Fatalf("ComputeNextCheckAtSeconds() = %d, want %d", got, want)
	}
}

func TestComputeNextCheckAtSecondsRetryAfterCappedByGlobalMax(t *testing.T) {
	now := int64(2000)
	got := ComputeNextCheckAtSeconds(
		now,
		60,
		300,
		0,
		now+3600,
		"",
		0,
	)

	want := now + 300
	if got != want {
		t.Fatalf("ComputeNextCheckAtSeconds() = %d, want %d", got, want)
	}
}

func TestComputeNextCheckAtSecondsUsesSafeDefaults(t *testing.T) {
	now := int64(123)
	got := ComputeNextCheckAtSeconds(
		now,
		0,
		0,
		0,
		0,
		"",
		0,
	)

	want := now + 1
	if got != want {
		t.Fatalf("ComputeNextCheckAtSeconds() = %d, want %d", got, want)
	}
}

func TestParseCacheControlMaxAgeSeconds(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int64
	}{
		{name: "normal", input: "public, max-age=600", want: 600},
		{name: "uppercase", input: "MAX-AGE=120", want: 120},
		{name: "invalid", input: "max-age=abc", want: 0},
		{name: "missing", input: "no-store", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCacheControlMaxAgeSeconds(tt.input)
			if got != tt.want {
				t.Fatalf("parseCacheControlMaxAgeSeconds(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestComputeNextCheckAtDurationWrapper(t *testing.T) {
	now := int64(100)
	got := ComputeNextCheckAt(
		now,
		30*time.Minute,
		7*24*time.Hour,
		1,
		0,
		"",
		0,
	)

	if got <= now {
		t.Fatalf("ComputeNextCheckAt() = %d, want > %d", got, now)
	}
}

func TestShouldSkip(t *testing.T) {
	interval := 30 * time.Minute
	maxBackoff := 7 * 24 * time.Hour
	now := time.Now().Unix()

	tests := []struct {
		name  string
		state FeedRuntimeState
		want  bool
	}{
		{
			name: "suspended feed",
			state: FeedRuntimeState{
				Suspended:     true,
				LastCheckedAt: now - 3600,
			},
			want: true,
		},
		{
			name: "recently updated (10 min ago)",
			state: FeedRuntimeState{
				LastCheckedAt: now - 600,
			},
			want: true,
		},
		{
			name: "ready to pull (40 min ago)",
			state: FeedRuntimeState{
				LastCheckedAt: now - 2400,
			},
			want: false,
		},
		{
			name: "1 failure, in backoff period",
			state: FeedRuntimeState{
				LastCheckedAt:       now - 1800,
				ConsecutiveFailures: 1,
				LastErrorAt:         now - 1800,
			},
			want: true,
		},
		{
			name: "1 failure, backoff expired",
			state: FeedRuntimeState{
				LastCheckedAt:       now - 3600,
				ConsecutiveFailures: 1,
				LastErrorAt:         now - 3600,
			},
			want: false,
		},
		{
			name:  "never pulled before",
			state: FeedRuntimeState{},
			want:  false,
		},
		{
			name: "explicit next_check_at in future",
			state: FeedRuntimeState{
				NextCheckAt: now + 300,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldSkip(now, tt.state, interval, maxBackoff)
			if got != tt.want {
				t.Errorf("ShouldSkip() = %v, want %v", got, tt.want)
			}
		})
	}
}
