package handler

import (
	"net/http"
	"testing"
)

func TestSearchRejectsWhitespaceOnlyQuery(t *testing.T) {
	h := &Handler{}

	r := newTestRouter()
	r.GET("/api/search", h.search)
	w := performRequest(r, http.MethodGet, "/api/search?q=%20%20%20", nil, nil)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}
