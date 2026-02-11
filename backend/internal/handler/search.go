package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) search(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		badRequestError(c, "q parameter is required")
		return
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		parsed, err := strconv.Atoi(l)
		if err != nil || parsed < 1 {
			badRequestError(c, "invalid limit")
			return
		}
		if parsed > maxListLimit {
			parsed = maxListLimit
		}
		limit = parsed
	}

	feeds, err := h.store.SearchFeeds(q)
	if err != nil {
		internalError(c, err, "search feeds")
		return
	}

	items, err := h.store.SearchItems(q, limit)
	if err != nil {
		internalError(c, err, "search items")
		return
	}

	dataResponse(c, gin.H{
		"feeds": feeds,
		"items": items,
	})
}
