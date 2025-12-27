package handler

import (
	"strconv"

	"github.com/0x2E/fusion/internal/store"
	"github.com/gin-gonic/gin"
)

type markItemsReadRequest struct {
	IDs []int64 `json:"ids" binding:"required"`
}

func (h *Handler) listItems(c *gin.Context) {
	params := store.ListItemsParams{}

	if feedID := c.Query("feed_id"); feedID != "" {
		id, err := strconv.ParseInt(feedID, 10, 64)
		if err != nil {
			errorResponse(c, 400, "invalid feed_id")
			return
		}
		params.FeedID = &id
	}

	if unread := c.Query("unread"); unread != "" {
		val, err := strconv.ParseBool(unread)
		if err != nil {
			errorResponse(c, 400, "invalid unread")
			return
		}
		params.Unread = &val
	}

	if limit := c.Query("limit"); limit != "" {
		val, err := strconv.Atoi(limit)
		if err != nil {
			errorResponse(c, 400, "invalid limit")
			return
		}
		params.Limit = val
	} else {
		params.Limit = 50
	}

	if offset := c.Query("offset"); offset != "" {
		val, err := strconv.Atoi(offset)
		if err != nil {
			errorResponse(c, 400, "invalid offset")
			return
		}
		params.Offset = val
	}

	if orderBy := c.Query("order_by"); orderBy != "" {
		params.OrderBy = orderBy
	} else {
		params.OrderBy = "pub_date"
	}

	items, err := h.store.ListItems(params)
	if err != nil {
		errorResponse(c, 500, err.Error())
		return
	}

	listResponse(c, items, len(items))
}

func (h *Handler) getItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errorResponse(c, 400, "invalid id")
		return
	}

	item, err := h.store.GetItem(id)
	if err != nil {
		errorResponse(c, 404, "item not found")
		return
	}

	dataResponse(c, item)
}

func (h *Handler) markItemsRead(c *gin.Context) {
	var req markItemsReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, 400, "invalid request")
		return
	}

	if err := h.store.BatchUpdateItemsUnread(req.IDs, false); err != nil {
		errorResponse(c, 500, err.Error())
		return
	}

	dataResponse(c, gin.H{"message": "items marked as read"})
}
