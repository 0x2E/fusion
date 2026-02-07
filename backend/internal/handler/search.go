package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) search(c *gin.Context) {
	q := c.Query("q")
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
		limit = parsed
	}

	feeds, err := h.store.SearchFeeds(q)
	if err != nil {
		c.JSON(500, gin.H{"error": "search feeds: " + err.Error()})
		return
	}

	items, err := h.store.SearchItems(q, limit)
	if err != nil {
		c.JSON(500, gin.H{"error": "search items: " + err.Error()})
		return
	}

	dataResponse(c, gin.H{
		"feeds": feeds,
		"items": items,
	})
}
