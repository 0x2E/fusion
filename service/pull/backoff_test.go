package pull_test

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/0x2e/fusion/service/pull"
)

func TestCalculateBackoffTime(t *testing.T) {
	for _, tt := range []struct {
		name                string
		consecutiveFailures uint
		expectedBackoff     time.Duration
	}{
		{
			name:                "no failures",
			consecutiveFailures: 0,
			expectedBackoff:     0,
		},
		{
			name:                "one failure",
			consecutiveFailures: 1,
			expectedBackoff:     54 * time.Minute, // 30 * (1.8^1) = 54 minutes
		},
		{
			name:                "two failures",
			consecutiveFailures: 2,
			expectedBackoff:     97 * time.Minute, // 30 * (1.8^2) = 97.2 minutes ≈ 97 minutes
		},
		{
			name:                "three failures",
			consecutiveFailures: 3,
			expectedBackoff:     174 * time.Minute, // 30 * (1.8^3) = 174.96 minutes ≈ 174 minutes
		},
		{
			name:                "many failures",
			consecutiveFailures: 10000,
			expectedBackoff:     7 * 24 * time.Hour, // Maximum backoff (7 days)
		},
		{
			name:                "maximum failures",
			consecutiveFailures: math.MaxUint,
			expectedBackoff:     7 * 24 * time.Hour, // Maximum backoff (7 days)
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			backoff := pull.CalculateBackoffTime(tt.consecutiveFailures)
			assert.Equal(t, tt.expectedBackoff, backoff)
		})
	}
}
