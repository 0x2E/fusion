package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/0x2E/fusion/internal/auth"
	"github.com/gin-gonic/gin"
)

func newTestSessionHandler(t *testing.T, password string) *Handler {
	t.Helper()

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	return &Handler{
		passwordHash: hash,
		sessions:     make(map[string]bool),
		limiter:      newLoginLimiter(10, 60, 300),
	}
}

func TestLoginRejectsMissingPasswordField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestSessionHandler(t, "secret")

	r := gin.New()
	r.POST("/api/sessions", h.login)

	req := httptest.NewRequest(http.MethodPost, "/api/sessions", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestLoginAcceptsEmptyPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestSessionHandler(t, "")

	r := gin.New()
	r.POST("/api/sessions", h.login)

	req := httptest.NewRequest(http.MethodPost, "/api/sessions", strings.NewReader(`{"password":""}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if cookie := w.Header().Get("Set-Cookie"); !strings.Contains(cookie, "session=") {
		t.Fatalf("expected session cookie to be set, got %q", cookie)
	}
}
