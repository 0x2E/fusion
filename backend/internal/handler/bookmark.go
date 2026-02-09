package handler

import (
	"errors"
	"strconv"

	"github.com/0x2E/fusion/internal/store"
	"github.com/gin-gonic/gin"
)

type createBookmarkRequest struct {
	ItemID   *int64 `json:"item_id"`
	Link     string `json:"link"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	PubDate  int64  `json:"pub_date"`
	FeedName string `json:"feed_name"`
}

func (h *Handler) listBookmarks(c *gin.Context) {
	limit := 50
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		val, err := strconv.Atoi(limitStr)
		if err != nil || val <= 0 {
			badRequestError(c, "invalid limit")
			return
		}
		if val > maxListLimit {
			val = maxListLimit
		}
		limit = val
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		val, err := strconv.Atoi(offsetStr)
		if err != nil || val < 0 {
			badRequestError(c, "invalid offset")
			return
		}
		offset = val
	}

	bookmarks, err := h.store.ListBookmarks(limit, offset)
	if err != nil {
		internalError(c, err, "list bookmarks")
		return
	}

	total, err := h.store.CountBookmarks()
	if err != nil {
		internalError(c, err, "count bookmarks")
		return
	}

	listResponse(c, bookmarks, total)
}

func (h *Handler) getBookmark(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestError(c, "invalid id")
		return
	}

	bookmark, err := h.store.GetBookmark(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFoundError(c, "bookmark")
			return
		}
		internalError(c, err, "get bookmark")
		return
	}

	dataResponse(c, bookmark)
}

func (h *Handler) createBookmark(c *gin.Context) {
	var req createBookmarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestError(c, "invalid request")
		return
	}

	var link, title, content, feedName string
	var pubDate int64

	// If item_id provided, auto-fill bookmark fields from item
	if req.ItemID != nil {
		item, err := h.store.GetItem(*req.ItemID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				notFoundError(c, "item")
				return
			}
			internalError(c, err, "get item for bookmark")
			return
		}

		feed, err := h.store.GetFeed(item.FeedID)
		if err != nil {
			internalError(c, err, "get feed for bookmark")
			return
		}

		link = item.Link
		title = item.Title
		content = item.Content
		pubDate = item.PubDate
		feedName = feed.Name
	} else {
		if req.Link == "" || req.Title == "" || req.Content == "" || req.FeedName == "" {
			badRequestError(c, "missing required fields")
			return
		}
		link = req.Link
		title = req.Title
		content = req.Content
		pubDate = req.PubDate
		feedName = req.FeedName
	}

	bookmark, err := h.store.CreateBookmark(req.ItemID, link, title, content, pubDate, feedName)
	if err != nil {
		internalError(c, err, "create bookmark")
		return
	}

	dataResponse(c, bookmark)
}

func (h *Handler) deleteBookmark(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestError(c, "invalid id")
		return
	}

	if err := h.store.DeleteBookmark(id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFoundError(c, "bookmark")
			return
		}
		internalError(c, err, "delete bookmark")
		return
	}

	dataResponse(c, gin.H{"message": "bookmark deleted"})
}
