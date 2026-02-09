package handler

import (
	"net/http"
	"sync"
	"time"

	"github.com/0x2E/fusion/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	sessionTTL           = 30 * 24 * time.Hour
	sessionSweepInterval = 60 * time.Second
)

type loginState struct {
	windowStart int64
	failures    int
	blockedTill int64
}

type loginLimiter struct {
	mu           sync.Mutex
	states       map[string]loginState
	limit        int
	windowSecs   int64
	blockSecs    int64
	lastSweepSec int64
}

func newLoginLimiter(limit, windowSecs, blockSecs int) *loginLimiter {
	return &loginLimiter{
		states:     make(map[string]loginState),
		limit:      limit,
		windowSecs: int64(windowSecs),
		blockSecs:  int64(blockSecs),
	}
}

func (l *loginLimiter) allow(ip string, now time.Time) (bool, int64) {
	nowSec := now.Unix()

	l.mu.Lock()
	defer l.mu.Unlock()

	l.sweep(nowSec)

	state, ok := l.states[ip]
	if !ok {
		return true, 0
	}
	if state.blockedTill > nowSec {
		return false, state.blockedTill - nowSec
	}

	return true, 0
}

func (l *loginLimiter) recordFailure(ip string, now time.Time) {
	nowSec := now.Unix()

	l.mu.Lock()
	defer l.mu.Unlock()

	l.sweep(nowSec)

	state := l.states[ip]
	if state.windowStart == 0 || nowSec-state.windowStart >= l.windowSecs {
		state.windowStart = nowSec
		state.failures = 0
	}

	state.failures++
	if state.failures >= l.limit {
		state.blockedTill = nowSec + l.blockSecs
		state.windowStart = nowSec
		state.failures = 0
	}

	l.states[ip] = state
}

func (l *loginLimiter) recordSuccess(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.states, ip)
}

func (l *loginLimiter) sweep(nowSec int64) {
	if nowSec-l.lastSweepSec < 60 {
		return
	}
	l.lastSweepSec = nowSec

	for ip, state := range l.states {
		windowExpired := state.windowStart > 0 && nowSec-state.windowStart >= l.windowSecs
		unblocked := state.blockedTill > 0 && state.blockedTill <= nowSec
		if (state.blockedTill == 0 && windowExpired) || unblocked {
			delete(l.states, ip)
		}
	}
}

func isSecureRequest(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return r.Header.Get("X-Forwarded-Proto") == "https"
}

type loginRequest struct {
	Password *string `json:"password"`
}

func (h *Handler) login(c *gin.Context) {
	ip := c.ClientIP()
	allowed, retryAfter := h.limiter.allow(ip, time.Now())
	if !allowed {
		tooManyRequestsError(c, retryAfter)
		return
	}

	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestError(c, "invalid request")
		return
	}
	if req.Password == nil {
		badRequestError(c, "invalid request")
		return
	}

	if err := auth.CheckPassword(h.passwordHash, *req.Password); err != nil {
		h.limiter.recordFailure(ip, time.Now())
		unauthorizedError(c)
		return
	}
	h.limiter.recordSuccess(ip)

	h.createSession(c)
	dataResponse(c, gin.H{"message": "logged in"})
}

func (h *Handler) isSessionValid(sessionID string) bool {
	nowSec := time.Now().Unix()

	h.mu.Lock()
	defer h.mu.Unlock()

	h.sweepExpiredSessionsLocked(nowSec)

	expiresAt, ok := h.sessions[sessionID]
	if !ok {
		return false
	}

	if expiresAt <= nowSec {
		delete(h.sessions, sessionID)
		return false
	}

	return true
}

func (h *Handler) sweepExpiredSessionsLocked(nowSec int64) {
	if nowSec-h.lastSweep < int64(sessionSweepInterval.Seconds()) {
		return
	}
	h.lastSweep = nowSec

	for sessionID, expiresAt := range h.sessions {
		if expiresAt <= nowSec {
			delete(h.sessions, sessionID)
		}
	}
}

// createSession generates a new session ID, stores it, and sets the session cookie.
func (h *Handler) createSession(c *gin.Context) {
	now := time.Now()
	expiresAt := now.Add(sessionTTL).Unix()
	sessionID := uuid.New().String()

	h.mu.Lock()
	h.sweepExpiredSessionsLocked(now.Unix())
	h.sessions[sessionID] = expiresAt
	h.mu.Unlock()

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   int(sessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   isSecureRequest(c.Request),
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) logout(c *gin.Context) {
	sessionID, err := c.Cookie("session")
	if err == nil {
		h.mu.Lock()
		delete(h.sessions, sessionID)
		h.mu.Unlock()
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isSecureRequest(c.Request),
		SameSite: http.SameSiteLaxMode,
	})

	dataResponse(c, gin.H{"message": "logged out"})
}
