package handler

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestListSavedBookmarkRefs(t *testing.T) {
	h, st := newFeverTestHandler(t)

	group, err := st.CreateGroup("Tech")
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	feed, err := st.CreateFeed(group.ID, "Fusion Feed", "https://example.com/rss.xml", "https://example.com", "")
	if err != nil {
		t.Fatalf("create feed: %v", err)
	}
	item, err := st.CreateItem(feed.ID, "guid-1", "Entry 1", "https://example.com/entry-1", "<p>Hello</p>", 1700000000)
	if err != nil {
		t.Fatalf("create item: %v", err)
	}
	bookmark, err := st.CreateBookmark(&item.ID, item.Link, item.Title, item.Content, item.PubDate, feed.Name)
	if err != nil {
		t.Fatalf("create bookmark: %v", err)
	}
	if _, err := st.CreateBookmark(nil, "https://example.com/orphan", "Orphan", "<p>orphan</p>", 1700000001, feed.Name); err != nil {
		t.Fatalf("create orphan bookmark: %v", err)
	}

	r := newTestRouter()
	r.GET("/api/bookmarks/-/items", h.listSavedBookmarkRefs)

	w := performRequest(r, http.MethodGet, "/api/bookmarks/-/items", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var payload struct {
		Data []struct {
			BookmarkID int64 `json:"bookmark_id"`
			ItemID     int64 `json:"item_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(payload.Data) != 1 {
		t.Fatalf("expected 1 saved bookmark ref, got %d", len(payload.Data))
	}
	if payload.Data[0].BookmarkID != bookmark.ID || payload.Data[0].ItemID != item.ID {
		t.Fatalf("unexpected saved bookmark ref: %#v", payload.Data[0])
	}
}
