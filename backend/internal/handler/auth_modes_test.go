package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/0x2E/fusion/internal/config"
	"github.com/0x2E/fusion/internal/store"
)

// newOIDCDiscoveryServer serves only the discovery document so tests can build
// an OIDC-backed handler without reaching an external provider.
func newOIDCDiscoveryServer(t *testing.T) *httptest.Server {
	t.Helper()

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/openid-configuration" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"issuer":                 server.URL,
			"authorization_endpoint": server.URL + "/authorize",
			"token_endpoint":         server.URL + "/token",
			"jwks_uri":               server.URL + "/jwks",
		})
	}))
	t.Cleanup(server.Close)

	return server
}

func newAuthModeTestHandler(t *testing.T, cfg *config.Config) *Handler {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	st, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}

	if cfg.FeverUsername == "" {
		cfg.FeverUsername = "fusion"
	}
	if cfg.LoginRateLimit == 0 {
		cfg.LoginRateLimit = 10
	}
	if cfg.LoginWindow == 0 {
		cfg.LoginWindow = 60
	}
	if cfg.LoginBlock == 0 {
		cfg.LoginBlock = 300
	}
	if cfg.PullTimeout == 0 {
		cfg.PullTimeout = 30
	}

	h, err := New(st, cfg, noopPuller{})
	if err != nil {
		_ = st.Close()
		t.Fatalf("new handler: %v", err)
	}

	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("close store: %v", err)
		}
	})

	return h
}

func decodeFeverAuth(t *testing.T, w *httptest.ResponseRecorder) float64 {
	t.Helper()

	var payload struct {
		Auth float64 `json:"auth"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal fever response: %v (body: %s)", err, w.Body.String())
	}

	return payload.Auth
}

func TestOIDCOnlyDisablesPasswordLoginAndFever(t *testing.T) {
	discovery := newOIDCDiscoveryServer(t)

	tests := []struct {
		name     string
		password string
	}{
		{name: "empty password", password: ""},
		{name: "whitespace-only password", password: "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newAuthModeTestHandler(t, &config.Config{
				Password:         tt.password,
				OIDCIssuer:       discovery.URL,
				OIDCClientID:     "fusion",
				OIDCClientSecret: "x",
				OIDCRedirectURI:  discovery.URL + "/api/oidc/callback",
			})
			router := h.SetupRouter()

			// Password login must not be registered in OIDC-only mode.
			w := performRequest(
				router,
				http.MethodPost,
				"/api/sessions",
				mustJSONBody(t, map[string]string{"password": tt.password}),
				map[string]string{"Content-Type": "application/json"},
			)
			if w.Code != http.StatusNotFound {
				t.Fatalf("expected POST /api/sessions to be unregistered, got %d", w.Code)
			}
			if cookie := w.Header().Get("Set-Cookie"); cookie != "" {
				t.Fatalf("expected no session cookie, got %q", cookie)
			}

			// Session status and logout remain available for OIDC sessions.
			w = performRequest(router, http.MethodGet, "/api/sessions", nil, nil)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected GET /api/sessions to be registered, got %d", w.Code)
			}
			w = performRequest(router, http.MethodDelete, "/api/sessions", nil, nil)
			if w.Code != http.StatusNoContent {
				t.Fatalf("expected DELETE /api/sessions to be registered, got %d", w.Code)
			}

			w = performRequest(router, http.MethodGet, "/api/feeds", nil, nil)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected protected API to require a session, got %d", w.Code)
			}

			for _, password := range []string{"", tt.password} {
				w = performRequest(
					router,
					http.MethodPost,
					"/fever",
					strings.NewReader(feverRequestBody(deriveFeverAPIKey("fusion", password), nil)),
					map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
				)
				if got := decodeFeverAuth(t, w); got != 0 {
					t.Fatalf("expected Fever auth=0 for password %q, got %v", password, got)
				}
			}
		})
	}
}

func TestPasswordConfiguredLoginAndFever(t *testing.T) {
	h := newAuthModeTestHandler(t, &config.Config{Password: "secret"})
	router := h.SetupRouter()

	for _, body := range []string{`{"password":"wrong"}`, `{"password":""}`} {
		w := performRequest(
			router,
			http.MethodPost,
			"/api/sessions",
			strings.NewReader(body),
			map[string]string{"Content-Type": "application/json"},
		)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected %s to be rejected, got %d", body, w.Code)
		}
	}

	w := performRequest(
		router,
		http.MethodPost,
		"/api/sessions",
		strings.NewReader(`{"password":"secret"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected correct password to log in, got %d: %s", w.Code, w.Body.String())
	}
	if cookie := w.Header().Get("Set-Cookie"); !strings.Contains(cookie, "session=") {
		t.Fatalf("expected session cookie, got %q", cookie)
	}

	w = performRequest(
		router,
		http.MethodPost,
		"/fever",
		strings.NewReader(feverRequestBody(deriveFeverAPIKey("fusion", "secret"), nil)),
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
	)
	if got := decodeFeverAuth(t, w); got != 1 {
		t.Fatalf("expected Fever auth=1, got %v", got)
	}
}

func TestAnonymousModeKeepsAPIAndFeverOpen(t *testing.T) {
	h := newAuthModeTestHandler(t, &config.Config{})
	router := h.SetupRouter()

	w := performRequest(router, http.MethodGet, "/api/feeds", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected anonymous API access, got %d", w.Code)
	}

	w = performRequest(
		router,
		http.MethodPost,
		"/fever",
		strings.NewReader(feverRequestBody(deriveFeverAPIKey("fusion", ""), nil)),
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
	)
	if got := decodeFeverAuth(t, w); got != 1 {
		t.Fatalf("expected Fever auth=1, got %v", got)
	}
}
