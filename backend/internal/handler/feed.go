package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/0x2E/feedfinder"
	"github.com/0x2E/fusion/internal/pkg/httpc"
	"github.com/0x2E/fusion/internal/pull"
	"github.com/0x2E/fusion/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/mmcdole/gofeed"
	"golang.org/x/sync/errgroup"
)

type createFeedRequest struct {
	GroupID                int64  `json:"group_id" binding:"required"`
	Name                   string `json:"name" binding:"required"`
	Link                   string `json:"link" binding:"required"`
	SiteURL                string `json:"site_url"`
	Proxy                  string `json:"proxy"`
	RefreshIntervalSeconds *int64 `json:"refresh_interval_seconds"`
}

type updateFeedRequest struct {
	GroupID                *int64  `json:"group_id"`
	Name                   *string `json:"name"`
	Link                   *string `json:"link"`
	SiteURL                *string `json:"site_url"`
	Suspended              *bool   `json:"suspended"`
	Proxy                  *string `json:"proxy"`
	RefreshIntervalSeconds *int64  `json:"refresh_interval_seconds"`
}

type validateFeedRequest struct {
	URL   string `json:"url" binding:"required"`
	Proxy string `json:"proxy"`
}

type discoveredFeed struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

type validateFeedResponse struct {
	Feeds []discoveredFeed `json:"feeds"`
}

// Notice codes served by GET /api/app. They are stable identifiers; all copy
// lives in the frontend.
const noticeOIDCAllowedUserUnset = "oidc_allowed_user_unset"

type appInfoResponse struct {
	PullInterval            int      `json:"pull_interval"`
	AllowedRefreshIntervals []int64  `json:"allowed_refresh_intervals"`
	Notices                 []string `json:"notices"`
}

type batchCreateFeedsRequest struct {
	Feeds []batchCreateFeedItem `json:"feeds" binding:"required,dive"`
}

type batchCreateFeedItem struct {
	GroupID int64  `json:"group_id" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Link    string `json:"link" binding:"required"`
	SiteURL string `json:"site_url"`
}

// allowedRefreshIntervals is the closed set of per-feed refresh overrides.
// It is exposed through GET /api/app so the frontend renders the same list
// the backend validates against.
var allowedRefreshIntervals = []int64{900, 1800, 3600, 7200, 21600, 43200, 86400}

func isAllowedRefreshInterval(v int64) bool {
	for _, allowed := range allowedRefreshIntervals {
		if v == allowed {
			return true
		}
	}
	return false
}

func invalidRefreshIntervalError() error {
	return fmt.Errorf("invalid refresh_interval_seconds: must be one of %v", allowedRefreshIntervals)
}

// normalizeCreateRefreshInterval maps the request semantics to storage: nil or
// 0 means "use the global interval" and is stored as NULL; any other value
// must be whitelisted.
func normalizeCreateRefreshInterval(v *int64) (*int64, error) {
	if v == nil || *v == 0 {
		return nil, nil
	}
	if !isAllowedRefreshInterval(*v) {
		return nil, invalidRefreshIntervalError()
	}
	return v, nil
}

func (h *Handler) getAppInfo(c *gin.Context) {
	dataResponse(c, appInfoResponse{
		PullInterval:            h.config.PullInterval,
		AllowedRefreshIntervals: allowedRefreshIntervals,
		Notices:                 h.notices,
	})
}

const refreshAllTimeout = 30 * time.Minute

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
	if err := httpc.ValidateRequestURL(c.Request.Context(), req.Link, h.allowPrivateFeeds()); err != nil {
		badRequestError(c, fmt.Sprintf("invalid URL: %v", err))
		return
	}
	refreshInterval, err := normalizeCreateRefreshInterval(req.RefreshIntervalSeconds)
	if err != nil {
		badRequestError(c, err.Error())
		return
	}

	feed, err := h.store.CreateFeed(req.GroupID, req.Name, req.Link, req.SiteURL, req.Proxy, refreshInterval)
	if err != nil {
		internalError(c, err, "create feed")
		return
	}

	h.startInitialPull(feed.ID)

	dataResponse(c, feed)
}

// startInitialPull warms a newly created feed without making the request wait
// for it. The puller wakes the scheduler when the pull finishes.
func (h *Handler) startInitialPull(feedID int64) {
	refreshTimeout := time.Duration(h.config.PullTimeout) * time.Second
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), refreshTimeout)
		defer cancel()
		if err := h.puller.RefreshFeed(ctx, feedID); err != nil {
			slog.Warn("initial feed pull failed", "feed_id", feedID, "error", err)
		}
	}()
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
	if req.Link != nil {
		if err := httpc.ValidateRequestURL(c.Request.Context(), *req.Link, h.allowPrivateFeeds()); err != nil {
			badRequestError(c, fmt.Sprintf("invalid URL: %v", err))
			return
		}
		params.Link = req.Link
	}
	if req.SiteURL != nil {
		params.SiteURL = req.SiteURL
	}
	if req.Suspended != nil {
		params.Suspended = req.Suspended
	}
	if req.Proxy != nil {
		params.Proxy = req.Proxy
	}
	// refresh_interval_seconds on update: absent = keep, 0 = clear back to the
	// global interval, other values must be whitelisted.
	if req.RefreshIntervalSeconds != nil {
		if *req.RefreshIntervalSeconds == 0 {
			params.ClearRefreshInterval = true
		} else if !isAllowedRefreshInterval(*req.RefreshIntervalSeconds) {
			badRequestError(c, invalidRefreshIntervalError().Error())
			return
		} else {
			params.RefreshIntervalSeconds = req.RefreshIntervalSeconds
		}
	}

	if err := h.store.UpdateFeed(id, params); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFoundError(c, "feed")
			return
		}
		internalError(c, err, "update feed")
		return
	}
	h.puller.Wake()

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
	h.puller.Wake()

	c.Status(http.StatusNoContent)
}

// newFeedDiscoveryClient builds the HTTP client used for feed discovery. It
// is a variable so tests can route a fake public address to a local server.
var newFeedDiscoveryClient = httpc.NewClient

// discoverFeeds resolves a user-supplied URL to feeds. Invalid input (URL or
// proxy) is returned as an error; a failed discovery is not, it just yields no
// feeds. Shared by the add-feed dialog and the GReader quickadd endpoint.
func discoverFeeds(ctx context.Context, target, proxy string, allowPrivateFeeds bool) ([]discoveredFeed, error) {
	if err := httpc.ValidateRequestURL(ctx, target, allowPrivateFeeds); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	client, err := newFeedDiscoveryClient(30*time.Second, proxy, allowPrivateFeeds)
	if err != nil {
		return nil, errors.New("invalid proxy")
	}

	found, err := feedfinder.Find(ctx, target, &feedfinder.Options{HTTPClient: client})
	if err != nil {
		slog.Warn("feed discovery failed", "url", target, "error", err)
	}

	feeds := normalizeDiscoveredFeeds(found)
	if len(feeds) == 0 {
		title, parseErr := parseFeedTitle(ctx, client, target)
		if parseErr == nil {
			feeds = append(feeds, discoveredFeed{Title: title, Link: target})
		}
	}

	return feeds, nil
}

func (h *Handler) validateFeed(c *gin.Context) {
	var req validateFeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestError(c, "invalid request")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	feeds, err := discoverFeeds(ctx, strings.TrimSpace(req.URL), strings.TrimSpace(req.Proxy), h.allowPrivateFeeds())
	if err != nil {
		badRequestError(c, err.Error())
		return
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
		// feedfinder mixes gofeed-parsed feed titles with page <title>/<link>
		// text the HTML finder already decoded once; Feed carries no source
		// marker, so page-derived names take one extra decode here. That is
		// the same bounded residual as item titles and keeps both name sources
		// on one form.
		result = append(result, discoveredFeed{
			Title: pull.NormalizeTitle(feed.Title),
			Link:  link,
		})
	}

	return result
}

func parseFeedTitle(ctx context.Context, client *http.Client, target string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return "", err
	}
	httpc.SetDefaultHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("feed fetch failed")
	}

	parsedFeed, err := gofeed.NewParser().Parse(resp.Body)
	if err != nil {
		return "", err
	}

	if parsedFeed == nil {
		return "", nil
	}

	return pull.NormalizeTitle(parsedFeed.Title), nil
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

	c.Status(http.StatusAccepted)
}

func (h *Handler) refreshAllFeeds(c *gin.Context) {
	if !h.tryStartRefreshAll() {
		c.Status(http.StatusAccepted)
		return
	}

	// Run in background so the HTTP response returns immediately.
	go func() {
		defer h.finishRefreshAll()

		ctx, cancel := context.WithTimeout(context.Background(), refreshAllTimeout)
		defer cancel()

		if count, err := h.puller.RefreshAll(ctx); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			slog.Warn("refresh all feeds failed", "refreshed", count, "error", err)
		}
	}()

	c.Status(http.StatusAccepted)
}

func (h *Handler) tryStartRefreshAll() bool {
	h.refreshAllMu.Lock()
	defer h.refreshAllMu.Unlock()

	if h.refreshAllRunning {
		return false
	}

	h.refreshAllRunning = true
	return true
}

func (h *Handler) finishRefreshAll() {
	h.refreshAllMu.Lock()
	h.refreshAllRunning = false
	h.refreshAllMu.Unlock()
}

// batchCreateFeedsValidationConcurrency bounds concurrent URL validation so a
// large OPML import cannot flood the resolver or exhaust file descriptors.
const batchCreateFeedsValidationConcurrency = 8

// batchCreateFeedsValidationBudget caps the whole validation pass below the
// server's 30s WriteTimeout. Without it, enough stalled DNS lookups make the
// client see a dropped connection while the valid feeds still get committed.
// Feeds not validated in time are reported as failed and can be re-imported.
const batchCreateFeedsValidationBudget = 20 * time.Second

func (h *Handler) batchCreateFeeds(c *gin.Context) {
	var req batchCreateFeedsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestError(c, "invalid request")
		return
	}

	// Validate each feed independently: a dead domain in an OPML export must
	// not block the feeds that are still reachable. The checks run concurrently
	// because each DNS lookup can stall for seconds, and a serial pass over a
	// large export can outlive the server write timeout. Results are stored by
	// index so the response stays deterministic.
	validateCtx, cancel := context.WithTimeout(c.Request.Context(), batchCreateFeedsValidationBudget)
	defer cancel()
	validationErrs := make([]error, len(req.Feeds))
	var g errgroup.Group
	g.SetLimit(batchCreateFeedsValidationConcurrency)
	for i, f := range req.Feeds {
		g.Go(func() error {
			validationErrs[i] = httpc.ValidateRequestURL(validateCtx, f.Link, h.allowPrivateFeeds())
			return nil
		})
	}
	_ = g.Wait()

	inputs := make([]store.BatchCreateFeedsInput, 0, len(req.Feeds))
	var feedErrors []string
	for i, f := range req.Feeds {
		if err := validationErrs[i]; err != nil {
			feedErrors = append(feedErrors, fmt.Sprintf("invalid URL %s: %v", f.Link, err))
			continue
		}
		inputs = append(inputs, store.BatchCreateFeedsInput{
			GroupID: f.GroupID,
			Name:    f.Name,
			Link:    f.Link,
			SiteURL: f.SiteURL,
		})
	}

	result, err := h.store.BatchCreateFeeds(inputs)
	if err != nil {
		internalError(c, err, "batch create feeds")
		return
	}
	feedErrors = append(feedErrors, result.Errors...)

	for _, id := range result.CreatedIDs {
		h.startInitialPull(id)
	}

	dataResponse(c, gin.H{
		"created": result.Created,
		"failed":  len(feedErrors),
		"errors":  feedErrors,
	})
}
