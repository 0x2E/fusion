package handler

import (
	"strconv"

	"github.com/0x2E/fusion/internal/store"
	"github.com/gin-gonic/gin"
)

type createFeedRequest struct {
	GroupID int64  `json:"group_id" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Link    string `json:"link" binding:"required"`
	SiteURL string `json:"site_url"`
	Proxy   string `json:"proxy"`
}

type updateFeedRequest struct {
	GroupID *int64  `json:"group_id"`
	Name    *string `json:"name"`
	SiteURL *string `json:"site_url"`
	Proxy   *string `json:"proxy"` // Empty string clears proxy
}

type validateFeedRequest struct {
	URL string `json:"url" binding:"required"`
}

func (h *Handler) listFeeds(c *gin.Context) {
	feeds, err := h.store.ListFeeds()
	if err != nil {
		errorResponse(c, 500, err.Error())
		return
	}

	listResponse(c, feeds, len(feeds))
}

func (h *Handler) getFeed(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errorResponse(c, 400, "invalid id")
		return
	}

	feed, err := h.store.GetFeed(id)
	if err != nil {
		errorResponse(c, 404, "feed not found")
		return
	}

	dataResponse(c, feed)
}

func (h *Handler) createFeed(c *gin.Context) {
	var req createFeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, 400, "invalid request")
		return
	}

	feed, err := h.store.CreateFeed(req.GroupID, req.Name, req.Link, req.SiteURL, req.Proxy)
	if err != nil {
		errorResponse(c, 500, err.Error())
		return
	}

	dataResponse(c, feed)
}

func (h *Handler) updateFeed(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errorResponse(c, 400, "invalid id")
		return
	}

	var req updateFeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, 400, "invalid request")
		return
	}

	params := store.UpdateFeedParams{}
	if req.GroupID != nil {
		params.GroupID = req.GroupID
	}
	if req.Name != nil {
		params.Name = req.Name
	}
	if req.SiteURL != nil {
		params.SiteURL = req.SiteURL
	}
	if req.Proxy != nil {
		params.Proxy = req.Proxy
	}

	if err := h.store.UpdateFeed(id, params); err != nil {
		errorResponse(c, 500, err.Error())
		return
	}

	feed, err := h.store.GetFeed(id)
	if err != nil {
		errorResponse(c, 500, err.Error())
		return
	}

	dataResponse(c, feed)
}

func (h *Handler) deleteFeed(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errorResponse(c, 400, "invalid id")
		return
	}

	if err := h.store.DeleteFeed(id); err != nil {
		errorResponse(c, 500, err.Error())
		return
	}

	dataResponse(c, gin.H{"message": "feed deleted"})
}

func (h *Handler) validateFeed(c *gin.Context) {
	var req validateFeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, 400, "invalid request")
		return
	}

	// TODO implement feed validation (HTTP fetch + RSS/Atom parsing)
	dataResponse(c, gin.H{"valid": true})
}

func (h *Handler) refreshFeed(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errorResponse(c, 400, "invalid id")
		return
	}

	if _, err := h.store.GetFeed(id); err != nil {
		errorResponse(c, 404, "feed not found")
		return
	}

	// Trigger refresh in background
	go h.puller.RefreshFeed(c.Request.Context(), id)

	dataResponse(c, gin.H{"message": "refresh triggered"})
}
