package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/0x2E/fusion/internal/config"
	"github.com/gin-gonic/gin"
)

func newCORSTestRouter(h *Handler) *gin.Engine {
	r := newTestRouter()
	r.Use(h.corsMiddleware())
	r.Any("/api/test", func(c *gin.Context) {
		c.String(http.StatusOK, "handler reached")
	})
	return r
}

func requireNoCORSHeaders(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()

	for _, header := range []string{
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Credentials",
		"Access-Control-Allow-Headers",
		"Access-Control-Allow-Methods",
	} {
		if got := w.Header().Get(header); got != "" {
			t.Errorf("expected no %s header, got %q", header, got)
		}
	}
}

func TestCORSMiddlewareDoesNotExposeCrossOriginReadsWhenNotConfigured(t *testing.T) {
	h := &Handler{config: &config.Config{}}

	w := performRequest(
		newCORSTestRouter(h),
		http.MethodGet,
		"/api/test",
		nil,
		map[string]string{"Origin": "https://evil.example.com"},
	)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	requireNoCORSHeaders(t, w)
}

func TestCORSMiddlewareRejectsCrossOriginWritesWhenNotConfigured(t *testing.T) {
	h := &Handler{config: &config.Config{}}

	tests := []struct {
		name    string
		headers map[string]string
	}{
		{
			name:    "sibling subdomain origin",
			headers: map[string]string{"Origin": "https://evil.example.com"},
		},
		{
			name:    "same-site fetch metadata",
			headers: map[string]string{"Origin": "https://evil.example.com", "Sec-Fetch-Site": "same-site"},
		},
		{
			name:    "cross-site fetch metadata without origin",
			headers: map[string]string{"Sec-Fetch-Site": "cross-site"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// A simple text/plain POST needs no preflight, so the header set must
			// be rejected by the request itself, not only by preflight handling.
			headers := map[string]string{"Content-Type": "text/plain"}
			for key, value := range tt.headers {
				headers[key] = value
			}

			w := performRequest(
				newCORSTestRouter(h),
				http.MethodPost,
				"/api/test",
				strings.NewReader(`{"feed_id": 1}`),
				headers,
			)

			if w.Code != http.StatusForbidden {
				t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
			}
			if got := w.Body.String(); got != "" {
				t.Fatalf("expected empty body, got %q", got)
			}
			requireNoCORSHeaders(t, w)
		})
	}
}

func TestCORSMiddlewareAllowsSameOriginWrites(t *testing.T) {
	h := &Handler{config: &config.Config{}}

	tests := []struct {
		name    string
		headers map[string]string
	}{
		{
			name:    "fetch metadata from same origin",
			headers: map[string]string{"Origin": "http://example.com", "Sec-Fetch-Site": "same-origin"},
		},
		{
			name:    "origin matching host without fetch metadata",
			headers: map[string]string{"Origin": "http://example.com"},
		},
		{
			name: "non-browser client such as curl or a Fever client",
		},
		{
			name:    "browser navigation",
			headers: map[string]string{"Sec-Fetch-Site": "none"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(
				newCORSTestRouter(h),
				http.MethodPost,
				"/api/test",
				strings.NewReader(`{"feed_id": 1}`),
				tt.headers,
			)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
			}
			if got := w.Body.String(); got != "handler reached" {
				t.Fatalf("expected the request to reach the handler, got body %q", got)
			}
		})
	}
}

// Fever clients are not browsers: they send no Origin or Sec-Fetch-Site header
// and must keep working through the globally installed middleware.
func TestCORSMiddlewareKeepsFeverClientsWorking(t *testing.T) {
	h, _ := newFeverTestHandler(t)

	w := performRequest(
		h.SetupRouter(),
		http.MethodPost,
		"/fever",
		strings.NewReader(feverRequestBody(deriveFeverAPIKey("fusion", "secret"), nil)),
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
	)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload["auth"] != float64(1) {
		t.Fatalf("expected fever auth=1, got %#v", payload["auth"])
	}
}

func requireAllowedOriginHeaders(t *testing.T, w *httptest.ResponseRecorder, wantOrigin string) {
	t.Helper()

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != wantOrigin {
		t.Errorf("expected allow-origin header %q, got %q", wantOrigin, got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("expected allow-credentials header \"true\", got %q", got)
	}
	if got := w.Header().Get("Vary"); got != "Origin" {
		t.Errorf("expected Vary header %q, got %q", "Origin", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Error("expected allow-headers header to be set")
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("expected allow-methods header to be set")
	}
}

func TestCORSMiddlewareGrantsConfiguredOrigin(t *testing.T) {
	// Configuration that differs from a browser Origin only by case and a
	// trailing slash must still match.
	h := &Handler{config: &config.Config{CORSAllowedOrigins: []string{"https://App.Example.com/"}}}
	r := newCORSTestRouter(h)

	t.Run("read", func(t *testing.T) {
		w := performRequest(r, http.MethodGet, "/api/test", nil, map[string]string{"Origin": "https://app.example.com"})

		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
		}
		requireAllowedOriginHeaders(t, w, "https://app.example.com")
	})

	t.Run("write", func(t *testing.T) {
		w := performRequest(
			r,
			http.MethodPost,
			"/api/test",
			strings.NewReader(`{"feed_id": 1}`),
			map[string]string{
				"Content-Type":   "application/json",
				"Origin":         "https://app.example.com",
				"Sec-Fetch-Site": "cross-site",
			},
		)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
		}
		if got := w.Body.String(); got != "handler reached" {
			t.Fatalf("expected the request to reach the handler, got body %q", got)
		}
		requireAllowedOriginHeaders(t, w, "https://app.example.com")
	})

	t.Run("preflight", func(t *testing.T) {
		w := performRequest(
			r,
			http.MethodOptions,
			"/api/test",
			nil,
			map[string]string{
				"Origin":                        "https://app.example.com",
				"Access-Control-Request-Method": "POST",
			},
		)

		if w.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d", http.StatusNoContent, w.Code)
		}
		requireAllowedOriginHeaders(t, w, "https://app.example.com")
	})
}

func TestCORSMiddlewareDoesNotGrantPreflightToOtherOrigins(t *testing.T) {
	h := &Handler{config: &config.Config{CORSAllowedOrigins: []string{"https://app.example.com"}}}

	w := performRequest(
		newCORSTestRouter(h),
		http.MethodOptions,
		"/api/test",
		nil,
		map[string]string{
			"Origin":                        "https://evil.example.com",
			"Access-Control-Request-Method": "POST",
		},
	)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
	requireNoCORSHeaders(t, w)
}

func TestCORSMiddlewareTreatsWildcardAsAnOrdinaryEntry(t *testing.T) {
	h := &Handler{config: &config.Config{CORSAllowedOrigins: []string{"*"}}}
	r := newCORSTestRouter(h)

	t.Run("read", func(t *testing.T) {
		w := performRequest(r, http.MethodGet, "/api/test", nil, map[string]string{"Origin": "https://evil.example.com"})

		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
		}
		requireNoCORSHeaders(t, w)
	})

	t.Run("write", func(t *testing.T) {
		w := performRequest(
			r,
			http.MethodPost,
			"/api/test",
			strings.NewReader(`{"feed_id": 1}`),
			map[string]string{"Origin": "https://evil.example.com", "Content-Type": "text/plain"},
		)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
		}
		requireNoCORSHeaders(t, w)
	})
}
