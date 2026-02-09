package handler

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMarkItemsBatchValidation(t *testing.T) {
	ids := make([]int64, maxBatchUpdateIDs+1)
	for i := range ids {
		ids[i] = int64(i + 1)
	}

	tests := []struct {
		name    string
		path    string
		handler gin.HandlerFunc
		body    interface{}
	}{
		{name: "read rejects too many ids", path: "/api/items/-/read", handler: (&Handler{}).markItemsRead, body: gin.H{"ids": ids}},
		{name: "unread rejects empty ids", path: "/api/items/-/unread", handler: (&Handler{}).markItemsUnread, body: gin.H{"ids": []int64{}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRouter()
			r.PATCH(tt.path, tt.handler)

			w := performRequest(
				r,
				http.MethodPatch,
				tt.path,
				mustJSONBody(t, tt.body),
				map[string]string{"Content-Type": "application/json"},
			)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected status 400, got %d", w.Code)
			}
		})
	}
}
