package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/0x2E/feedfinder"
	"github.com/0x2E/fusion/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/mmcdole/gofeed"
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

type discoveredFeed struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

type validateFeedResponse struct {
	Feeds []discoveredFeed `json:"feeds"`
}

type batchCreateFeedsRequest struct {
	Feeds []batchCreateFeedItem `json:"feeds" binding:"required"`
}

type batchCreateFeedItem struct {
	GroupID int64  `json:"group_id" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Link    string `json:"link" binding:"required"`
}

func (h *Handler) listFeeds(c *gin.Context) {
	feeds, err := h.store.ListFeeds()
	if err != nil {
		internalError(c, err, "list feeds")
		return
	}

	listResponse(c, feeds, len(feeds))
}

func (h *Handler) getFeed(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestError(c, "invalid id")
		return
	}

	feed, err := h.store.GetFeed(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFoundError(c, "feed")
			return
		}
		internalError(c, err, "get feed")
		return
	}

	dataResponse(c, feed)
}

func (h *Handler) createFeed(c *gin.Context) {
	var req createFeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestError(c, "invalid request")
		return
	}

	feed, err := h.store.CreateFeed(req.GroupID, req.Name, req.Link, req.SiteURL, req.Proxy)
	if err != nil {
		internalError(c, err, "create feed")
		return
	}

	// Trigger initial pull in background.
	refreshTimeout := time.Duration(h.config.PullTimeout) * time.Second
	go func(feedID int64) {
		ctx, cancel := context.WithTimeout(context.Background(), refreshTimeout)
		defer cancel()
		if err := h.puller.RefreshFeed(ctx, feedID); err != nil {
			slog.Warn("initial feed pull failed", "feed_id", feedID, "error", err)
		}
	}(feed.ID)

	dataResponse(c, feed)
}

func (h *Handler) updateFeed(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestError(c, "invalid id")
		return
	}

	var req updateFeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestError(c, "invalid request")
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
		if errors.Is(err, store.ErrNotFound) {
			notFoundError(c, "feed")
			return
		}
		internalError(c, err, "update feed")
		return
	}

	feed, err := h.store.GetFeed(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFoundError(c, "feed")
			return
		}
		internalError(c, err, "get updated feed")
		return
	}

	dataResponse(c, feed)
}

func (h *Handler) deleteFeed(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestError(c, "invalid id")
		return
	}

	if err := h.store.DeleteFeed(id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFoundError(c, "feed")
			return
		}
		internalError(c, err, "delete feed")
		return
	}

	dataResponse(c, gin.H{"message": "feed deleted"})
}

func (h *Handler) validateFeed(c *gin.Context) {
	var req validateFeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestError(c, "invalid request")
		return
	}

	target := strings.TrimSpace(req.URL)
	parsedURL, err := url.ParseRequestURI(target)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		badRequestError(c, "invalid url")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	found, err := feedfinder.Find(ctx, target, nil)
	if err != nil {
		slog.Warn("feed discovery failed", "url", target, "error", err)
	}

	feeds := normalizeDiscoveredFeeds(found)
	if len(feeds) == 0 {
		parser := gofeed.NewParser()
		parsedFeed, parseErr := parser.ParseURLWithContext(target, ctx)
		if parseErr == nil {
			title := ""
			if parsedFeed != nil {
				title = strings.TrimSpace(parsedFeed.Title)
			}
			feeds = append(feeds, discoveredFeed{Title: title, Link: target})
		}
	}

	dataResponse(c, validateFeedResponse{Feeds: feeds})
}

func normalizeDiscoveredFeeds(found []feedfinder.Feed) []discoveredFeed {
	result := make([]discoveredFeed, 0, len(found))
	seen := make(map[string]struct{}, len(found))

	for _, feed := range found {
		link := strings.TrimSpace(feed.Link)
		if link == "" {
			continue
		}
		if _, exists := seen[link]; exists {
			continue
		}

		seen[link] = struct{}{}
		result = append(result, discoveredFeed{
			Title: strings.TrimSpace(feed.Title),
			Link:  link,
		})
	}

	return result
}

func (h *Handler) refreshFeed(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestError(c, "invalid id")
		return
	}

	if _, err := h.store.GetFeed(id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFoundError(c, "feed")
			return
		}
		internalError(c, err, "get feed for refresh")
		return
	}

	// Trigger refresh in background.
	// Do not use the request context here: once the handler returns, it may be cancelled.
	refreshTimeout := time.Duration(h.config.PullTimeout) * time.Second
	go func(feedID int64) {
		ctx, cancel := context.WithTimeout(context.Background(), refreshTimeout)
		defer cancel()
		if err := h.puller.RefreshFeed(ctx, feedID); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			slog.Warn("refresh feed failed", "feed_id", feedID, "error", err)
		}
	}(id)

	dataResponse(c, gin.H{"message": "refresh triggered"})
}

func (h *Handler) batchCreateFeeds(c *gin.Context) {
	var req batchCreateFeedsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestError(c, "invalid request")
		return
	}

	inputs := make([]store.BatchCreateFeedsInput, len(req.Feeds))
	for i, f := range req.Feeds {
		inputs[i] = store.BatchCreateFeedsInput{
			GroupID: f.GroupID,
			Name:    f.Name,
			Link:    f.Link,
		}
	}

	result, err := h.store.BatchCreateFeeds(inputs)
	if err != nil {
		internalError(c, err, "batch create feeds")
		return
	}

	// Trigger initial pull for each new feed in background.
	refreshTimeout := time.Duration(h.config.PullTimeout) * time.Second
	for _, id := range result.CreatedIDs {
		go func(feedID int64) {
			ctx, cancel := context.WithTimeout(context.Background(), refreshTimeout)
			defer cancel()
			if err := h.puller.RefreshFeed(ctx, feedID); err != nil {
				slog.Warn("initial feed pull failed", "feed_id", feedID, "error", err)
			}
		}(id)
	}

	dataResponse(c, gin.H{
		"created": result.Created,
		"failed":  len(result.Errors),
		"errors":  result.Errors,
	})
}
