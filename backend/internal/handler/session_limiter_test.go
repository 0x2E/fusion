package handler

import "testing"

func TestLoginLimiterSweep(t *testing.T) {
	tests := []struct {
		name       string
		windowSecs int
		blockSecs  int
		ip         string
		state      loginState
		nowSec     int64
		wantExists bool
	}{
		{
			name:       "deletes expired window without block",
			windowSecs: 10,
			blockSecs:  30,
			ip:         "1.1.1.1",
			state:      loginState{windowStart: 10, failures: 1},
			nowSec:     120,
			wantExists: false,
		},
		{
			name:       "deletes unblocked state",
			windowSecs: 60,
			blockSecs:  30,
			ip:         "2.2.2.2",
			state:      loginState{windowStart: 100, blockedTill: 110},
			nowSec:     120,
			wantExists: false,
		},
		{
			name:       "keeps active blocked state",
			windowSecs: 60,
			blockSecs:  30,
			ip:         "3.3.3.3",
			state:      loginState{windowStart: 100, blockedTill: 170},
			nowSec:     120,
			wantExists: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := newLoginLimiter(3, tt.windowSecs, tt.blockSecs)
			limiter.states[tt.ip] = tt.state

			limiter.sweep(tt.nowSec)

			_, ok := limiter.states[tt.ip]
			if ok != tt.wantExists {
				t.Fatalf("state exists = %v, want %v", ok, tt.wantExists)
			}
		})
	}
}
