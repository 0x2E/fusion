package handler

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSetupRouterServesEmbeddedFrontend(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{}
	r := h.SetupRouter()

	t.Run("serves index for root", func(t *testing.T) {
		w := performRequest(r, http.MethodGet, "/", nil, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		if contentType := w.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
			t.Fatalf("expected text/html content type, got %q", contentType)
		}
	})

	t.Run("serves index for client-side route", func(t *testing.T) {
		w := performRequest(r, http.MethodGet, "/feeds", nil, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		if contentType := w.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
			t.Fatalf("expected text/html content type, got %q", contentType)
		}
	})

	t.Run("returns 404 for unknown api path", func(t *testing.T) {
		w := performRequest(r, http.MethodGet, "/api/not-found", nil, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("returns 404 for missing asset", func(t *testing.T) {
		w := performRequest(r, http.MethodGet, "/assets/does-not-exist-0db74db9a5.js", nil, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}
	})
}
