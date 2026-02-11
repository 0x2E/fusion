package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func mustJSONBody(t *testing.T, payload interface{}) *bytes.Reader {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	return bytes.NewReader(body)
}

func performRequest(
	r http.Handler,
	method, target string,
	body io.Reader,
	headers map[string]string,
	cookies ...*http.Cookie,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, body)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
