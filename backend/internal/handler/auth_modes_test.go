package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/0x2E/fusion/internal/config"
	"github.com/0x2E/fusion/internal/store"
)

// oidcTestServer serves the discovery document and a token endpoint that
// records the exchange request, so tests can exercise OIDC flows offline.
type oidcTestServer struct {
	*httptest.Server

	mu           sync.Mutex
	tokenRequest url.Values
}

// newOIDCTestServer serves the discovery document and a token endpoint that
// records its last request, so tests can build an OIDC-backed handler and
// inspect the token exchange without reaching an external provider.
func newOIDCTestServer(t *testing.T) *oidcTestServer {
	t.Helper()

	s := &oidcTestServer{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"issuer":                 s.URL,
				"authorization_endpoint": s.URL + "/authorize",
				"token_endpoint":         s.URL + "/token",
				"jwks_uri":               s.URL + "/jwks",
			})
		case "/token":
			// The exchange is expected to fail after the request is recorded; no
			// tokens or JWKS are needed for that.
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			s.mu.Lock()
			s.tokenRequest = r.PostForm
			s.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(s.Close)

	return s
}

func (s *oidcTestServer) lastTokenRequest() url.Values {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.tokenRequest
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

func assertAuthMethods(t *testing.T, router http.Handler, wantPassword, wantOIDC bool) {
	t.Helper()

	w := performRequest(router, http.MethodGet, "/api/auth/methods", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected GET /api/auth/methods to return 200, got %d", w.Code)
	}

	var payload struct {
		Data struct {
			Password bool `json:"password"`
			OIDC     bool `json:"oidc"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal auth methods response: %v (body: %s)", err, w.Body.String())
	}
	if payload.Data.Password != wantPassword || payload.Data.OIDC != wantOIDC {
		t.Fatalf(
			"expected auth methods password=%v oidc=%v, got password=%v oidc=%v",
			wantPassword, wantOIDC, payload.Data.Password, payload.Data.OIDC,
		)
	}

	w = performRequest(router, http.MethodGet, "/api/oidc/enabled", nil, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected GET /api/oidc/enabled to be removed, got %d", w.Code)
	}
}

func TestOIDCOnlyDisablesPasswordLoginAndFever(t *testing.T) {
	discovery := newOIDCTestServer(t)

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
			assertAuthMethods(t, router, false, true)

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
	assertAuthMethods(t, router, true, false)

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

func TestPasswordAndOIDCConfigured(t *testing.T) {
	discovery := newOIDCTestServer(t)
	h := newAuthModeTestHandler(t, &config.Config{
		Password:         "secret",
		OIDCIssuer:       discovery.URL,
		OIDCClientID:     "fusion",
		OIDCClientSecret: "x",
		OIDCRedirectURI:  discovery.URL + "/api/oidc/callback",
	})
	router := h.SetupRouter()

	assertAuthMethods(t, router, true, true)
}

func TestAnonymousModeKeepsAPIAndFeverOpen(t *testing.T) {
	h := newAuthModeTestHandler(t, &config.Config{})
	router := h.SetupRouter()
	assertAuthMethods(t, router, false, false)

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

func loginSessionCookie(t *testing.T, router http.Handler, password string) *http.Cookie {
	t.Helper()

	w := performRequest(
		router,
		http.MethodPost,
		"/api/sessions",
		strings.NewReader(`{"password":"`+password+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected login to succeed, got %d: %s", w.Code, w.Body.String())
	}

	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == "session" {
			return cookie
		}
	}

	t.Fatalf("expected a session cookie, got %q", w.Header().Get("Set-Cookie"))
	return nil
}

func TestAppInfoNotices(t *testing.T) {
	discovery := newOIDCTestServer(t)

	tests := []struct {
		name        string
		oidcIssuer  string
		allowedUser string
		want        []string
	}{
		{name: "password only", want: []string{}},
		{name: "OIDC with allowed user", oidcIssuer: discovery.URL, allowedUser: "owner@example.com", want: []string{}},
		{name: "OIDC without allowed user", oidcIssuer: discovery.URL, want: []string{"oidc_allowed_user_unset"}},
		{name: "OIDC with blank allowed user", oidcIssuer: discovery.URL, allowedUser: "   ", want: []string{"oidc_allowed_user_unset"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newAuthModeTestHandler(t, &config.Config{
				Password:         "secret",
				OIDCIssuer:       tt.oidcIssuer,
				OIDCClientID:     "fusion",
				OIDCClientSecret: "x",
				OIDCAllowedUser:  tt.allowedUser,
			})
			router := h.SetupRouter()

			w := performRequest(router, http.MethodGet, "/api/app", nil, nil, loginSessionCookie(t, router, "secret"))
			if w.Code != http.StatusOK {
				t.Fatalf("expected GET /api/app to return 200, got %d: %s", w.Code, w.Body.String())
			}

			var payload struct {
				Data struct {
					Notices json.RawMessage `json:"notices"`
				} `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
				t.Fatalf("unmarshal app response: %v (body: %s)", err, w.Body.String())
			}
			if string(payload.Data.Notices) == "null" {
				t.Fatal("expected notices to always be a JSON array, got null")
			}

			var got []string
			if err := json.Unmarshal(payload.Data.Notices, &got); err != nil {
				t.Fatalf("unmarshal notices: %v", err)
			}
			if !slices.Equal(got, tt.want) {
				t.Fatalf("expected notices %v, got %v", tt.want, got)
			}
		})
	}
}

func TestOIDCLoginRedirectURI(t *testing.T) {
	discovery := newOIDCTestServer(t)

	tests := []struct {
		name        string
		redirectURI string
		headers     map[string]string
		want        string
	}{
		{
			name: "derived from request",
			want: "http://example.com/api/oidc/callback",
		},
		{
			name:    "derived from forwarded proto",
			headers: map[string]string{"X-Forwarded-Proto": "https"},
			want:    "https://example.com/api/oidc/callback",
		},
		{
			name:        "configured redirect URI wins",
			redirectURI: "https://fusion.example.com/api/oidc/callback",
			headers:     map[string]string{"X-Forwarded-Proto": "https"},
			want:        "https://fusion.example.com/api/oidc/callback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// OIDC-only configuration without a redirect URI must start.
			h := newAuthModeTestHandler(t, &config.Config{
				OIDCIssuer:       discovery.URL,
				OIDCClientID:     "fusion",
				OIDCClientSecret: "x",
				OIDCRedirectURI:  tt.redirectURI,
			})
			router := h.SetupRouter()

			w := performRequest(router, http.MethodGet, "/api/oidc/login", nil, tt.headers)
			if w.Code != http.StatusOK {
				t.Fatalf("expected GET /api/oidc/login to return 200, got %d: %s", w.Code, w.Body.String())
			}

			var payload struct {
				Data struct {
					AuthURL string `json:"auth_url"`
				} `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
				t.Fatalf("unmarshal login response: %v (body: %s)", err, w.Body.String())
			}

			authURL, err := url.Parse(payload.Data.AuthURL)
			if err != nil {
				t.Fatalf("parse auth url: %v", err)
			}
			if got := authURL.Query().Get("redirect_uri"); got != tt.want {
				t.Fatalf("expected redirect_uri %q, got %q", tt.want, got)
			}
		})
	}
}

func TestOIDCCallbackSendsLoginRedirectURI(t *testing.T) {
	discovery := newOIDCTestServer(t)
	h := newAuthModeTestHandler(t, &config.Config{
		OIDCIssuer:       discovery.URL,
		OIDCClientID:     "fusion",
		OIDCClientSecret: "x",
	})
	router := h.SetupRouter()

	w := performRequest(
		router,
		http.MethodGet,
		"/api/oidc/login",
		nil,
		map[string]string{"X-Forwarded-Proto": "https"},
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected GET /api/oidc/login to return 200, got %d: %s", w.Code, w.Body.String())
	}

	var payload struct {
		Data struct {
			AuthURL string `json:"auth_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal login response: %v (body: %s)", err, w.Body.String())
	}

	authURL, err := url.Parse(payload.Data.AuthURL)
	if err != nil {
		t.Fatalf("parse auth url: %v", err)
	}
	wantRedirectURI := authURL.Query().Get("redirect_uri")
	if wantRedirectURI != "https://example.com/api/oidc/callback" {
		t.Fatalf("unexpected login redirect_uri %q", wantRedirectURI)
	}

	// The token endpoint rejects the exchange, but only after recording it.
	w = performRequest(
		router,
		http.MethodGet,
		"/api/oidc/callback?state="+url.QueryEscape(authURL.Query().Get("state"))+"&code=test-code",
		nil,
		nil,
	)
	if w.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected callback to redirect, got %d: %s", w.Code, w.Body.String())
	}
	if location := w.Header().Get("Location"); location != "/login?error=oidc_failed" {
		t.Fatalf("expected failed-login redirect, got %q", location)
	}

	tokenRequest := discovery.lastTokenRequest()
	if tokenRequest == nil {
		t.Fatal("expected the token endpoint to be called")
	}
	if got := tokenRequest.Get("redirect_uri"); got != wantRedirectURI {
		t.Fatalf("expected token exchange redirect_uri %q, got %q", wantRedirectURI, got)
	}
	if got := tokenRequest.Get("code_verifier"); got == "" {
		t.Fatal("expected the token exchange to send the PKCE code_verifier")
	}
	if got := tokenRequest.Get("code"); got != "test-code" {
		t.Fatalf("expected token exchange code %q, got %q", "test-code", got)
	}
}
