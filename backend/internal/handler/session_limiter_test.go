package handler

import "testing"

func TestLoginLimiterSweepDeletesExpiredWindowWithoutBlock(t *testing.T) {
	limiter := newLoginLimiter(3, 10, 30)
	limiter.states["1.1.1.1"] = loginState{windowStart: 10, failures: 1}

	limiter.sweep(120)

	if _, ok := limiter.states["1.1.1.1"]; ok {
		t.Fatal("expected expired non-blocked state to be deleted")
	}
}

func TestLoginLimiterSweepDeletesUnblockedState(t *testing.T) {
	limiter := newLoginLimiter(3, 60, 30)
	limiter.states["2.2.2.2"] = loginState{windowStart: 100, blockedTill: 110}

	limiter.sweep(120)

	if _, ok := limiter.states["2.2.2.2"]; ok {
		t.Fatal("expected unblocked state to be deleted")
	}
}

func TestLoginLimiterSweepKeepsActiveBlockedState(t *testing.T) {
	limiter := newLoginLimiter(3, 60, 30)
	limiter.states["3.3.3.3"] = loginState{windowStart: 100, blockedTill: 170}

	limiter.sweep(120)

	if _, ok := limiter.states["3.3.3.3"]; !ok {
		t.Fatal("expected active blocked state to remain")
	}
}
