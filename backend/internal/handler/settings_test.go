package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type settingsEnvelope struct {
	Data struct {
		Locale          *string `json:"locale"`
		ArticlePageSize *int64  `json:"article_page_size"`
		Theme           *string `json:"theme"`
		UpdatedAt       int64   `json:"updated_at"`
	} `json:"data"`
}

func decodeSettings(t *testing.T, w *httptest.ResponseRecorder) settingsEnvelope {
	t.Helper()

	var envelope settingsEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode settings response failed: %v (body: %s)", err, w.Body.String())
	}
	return envelope
}

// newSettingsTestRouter returns the full router plus a valid session cookie.
func newSettingsTestRouter(t *testing.T) (*gin.Engine, *http.Cookie) {
	t.Helper()

	h, _ := newFeverTestHandler(t) // password: "secret"
	r := h.SetupRouter()

	w := performRequest(r, http.MethodPost, "/api/sessions", mustJSONBody(t, map[string]string{"password": "secret"}), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", w.Code, w.Body.String())
	}

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected a session cookie after login")
	}
	return r, cookies[0]
}

func TestGetSettingsEmpty(t *testing.T) {
	r, cookie := newSettingsTestRouter(t)

	w := performRequest(r, http.MethodGet, "/api/settings", nil, nil, cookie)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	settings := decodeSettings(t, w).Data
	if settings.Locale != nil || settings.ArticlePageSize != nil || settings.Theme != nil {
		t.Errorf("expected all-null settings, got %+v", settings)
	}

	// Pointer decoding cannot distinguish a missing key from a JSON null, so
	// assert on the raw body that all three keys are explicitly present.
	for _, key := range []string{`"locale":null`, `"article_page_size":null`, `"theme":null`} {
		if !strings.Contains(w.Body.String(), key) {
			t.Errorf("expected body to contain %s, got %s", key, w.Body.String())
		}
	}
}

func TestUpdateSettingsExplicitNullLeavesValueUnchanged(t *testing.T) {
	r, cookie := newSettingsTestRouter(t)

	w := performRequest(r, http.MethodPatch, "/api/settings", strings.NewReader(`{"locale":"de"}`), nil, cookie)
	if w.Code != http.StatusOK {
		t.Fatalf("PATCH locale: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	w = performRequest(r, http.MethodPatch, "/api/settings", strings.NewReader(`{"locale":null,"theme":"dark"}`), nil, cookie)
	if w.Code != http.StatusOK {
		t.Fatalf("PATCH null locale: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	settings := decodeSettings(t, w).Data
	if settings.Locale == nil || *settings.Locale != "de" {
		t.Errorf("expected explicit null to leave locale unchanged, got %+v", settings)
	}
	if settings.Theme == nil || *settings.Theme != "dark" {
		t.Errorf("expected theme=dark, got %+v", settings)
	}
}

func TestSettingsRequireAuth(t *testing.T) {
	r, _ := newSettingsTestRouter(t)

	for _, method := range []string{http.MethodGet, http.MethodPatch} {
		w := performRequest(r, method, "/api/settings", nil, nil)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s without cookie: expected 401, got %d", method, w.Code)
		}
	}
}

func TestUpdateSettingsSingleField(t *testing.T) {
	r, cookie := newSettingsTestRouter(t)

	w := performRequest(r, http.MethodPatch, "/api/settings", strings.NewReader(`{"locale":"de"}`), nil, cookie)
	if w.Code != http.StatusOK {
		t.Fatalf("PATCH locale: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	settings := decodeSettings(t, w).Data
	if settings.Locale == nil || *settings.Locale != "de" {
		t.Errorf("PATCH locale: expected locale=de in response, got %+v", settings)
	}

	w = performRequest(r, http.MethodPatch, "/api/settings", strings.NewReader(`{"theme":"dark"}`), nil, cookie)
	if w.Code != http.StatusOK {
		t.Fatalf("PATCH theme: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	settings = decodeSettings(t, w).Data
	if settings.Locale == nil || *settings.Locale != "de" {
		t.Errorf("PATCH theme: expected locale to survive, got %+v", settings)
	}
	if settings.Theme == nil || *settings.Theme != "dark" {
		t.Errorf("PATCH theme: expected theme=dark, got %+v", settings)
	}

	w = performRequest(r, http.MethodGet, "/api/settings", nil, nil, cookie)
	if w.Code != http.StatusOK {
		t.Fatalf("GET after updates: expected 200, got %d", w.Code)
	}
	settings = decodeSettings(t, w).Data
	if settings.UpdatedAt == 0 {
		t.Error("expected updated_at to be set once a preference is stored")
	}
}

func TestUpdateSettingsValidation(t *testing.T) {
	r, cookie := newSettingsTestRouter(t)

	tests := []struct {
		name string
		body string
		want string
	}{
		{"invalid locale", `{"locale":"xx"}`, "invalid locale"},
		{"empty locale", `{"locale":""}`, "invalid locale"},
		{"locale wrong type", `{"locale":5}`, "invalid request"},
		{"page size below range", `{"article_page_size":0}`, "invalid article_page_size"},
		{"page size above range", `{"article_page_size":101}`, "invalid article_page_size"},
		{"negative page size", `{"article_page_size":-10}`, "invalid article_page_size"},
		{"page size as string", `{"article_page_size":"20"}`, "invalid request"},
		{"invalid theme", `{"theme":"blue"}`, "invalid theme"},
		{"empty theme", `{"theme":""}`, "invalid theme"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(r, http.MethodPatch, "/api/settings", strings.NewReader(tt.body), nil, cookie)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
			}
			var errBody struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &errBody); err != nil {
				t.Fatalf("decode error body failed: %v", err)
			}
			if errBody.Error != tt.want {
				t.Errorf("expected error %q, got %q", tt.want, errBody.Error)
			}
		})
	}
}

func TestUpdateSettingsCustomPageSize(t *testing.T) {
	r, cookie := newSettingsTestRouter(t)

	// The five-option list is a frontend UI choice; any value the items API
	// can serve (1..maxListLimit) must roundtrip.
	w := performRequest(r, http.MethodPatch, "/api/settings", strings.NewReader(`{"article_page_size":25}`), nil, cookie)
	if w.Code != http.StatusOK {
		t.Fatalf("PATCH custom page size: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	settings := decodeSettings(t, w).Data
	if settings.ArticlePageSize == nil || *settings.ArticlePageSize != 25 {
		t.Errorf("expected article_page_size=25, got %+v", settings)
	}

	w = performRequest(r, http.MethodGet, "/api/settings", nil, nil, cookie)
	settings = decodeSettings(t, w).Data
	if settings.ArticlePageSize == nil || *settings.ArticlePageSize != 25 {
		t.Errorf("expected article_page_size=25 after GET, got %+v", settings)
	}
}

func TestUpdateSettingsEmptyBody(t *testing.T) {
	r, cookie := newSettingsTestRouter(t)

	w := performRequest(r, http.MethodPatch, "/api/settings", strings.NewReader(`{}`), nil, cookie)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	settings := decodeSettings(t, w).Data
	if settings.Locale != nil || settings.ArticlePageSize != nil || settings.Theme != nil {
		t.Errorf("expected empty PATCH to be a no-op, got %+v", settings)
	}
	if settings.UpdatedAt != 0 {
		t.Errorf("expected empty PATCH not to create the row, got updated_at=%d", settings.UpdatedAt)
	}
}
