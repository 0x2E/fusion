package handler

import (
	"crypto/md5"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/0x2E/fusion/internal/store"
	"github.com/gin-gonic/gin"
)

const (
	feverAPIVersion         = 3
	feverItemsLimit         = 50
	feverTransparentGIFData = "image/gif;base64,R0lGODlhAQABAIAAAObm5gAAACH5BAEAAAAALAAAAAABAAEAAAICRAEAOw=="
)

type feverGroup struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type feverFeedsGroup struct {
	GroupID int64  `json:"group_id"`
	FeedIDs string `json:"feed_ids"`
}

type feverFeed struct {
	ID                int64  `json:"id"`
	FaviconID         int64  `json:"favicon_id"`
	Title             string `json:"title"`
	URL               string `json:"url"`
	SiteURL           string `json:"site_url"`
	IsSpark           int    `json:"is_spark"`
	LastUpdatedOnTime int64  `json:"last_updated_on_time"`
}

type feverFavicon struct {
	ID   int64  `json:"id"`
	Data string `json:"data"`
}

type feverItem struct {
	ID            int64  `json:"id"`
	FeedID        int64  `json:"feed_id"`
	Title         string `json:"title"`
	Author        string `json:"author"`
	HTML          string `json:"html"`
	URL           string `json:"url"`
	IsSaved       int    `json:"is_saved"`
	IsRead        int    `json:"is_read"`
	CreatedOnTime int64  `json:"created_on_time"`
}

type feverMarkResult struct {
	IncludeUnreadItemIDs bool
	IncludeSavedItemIDs  bool
}

func deriveFeverAPIKey(username, password string) string {
	sum := md5.Sum([]byte(strings.TrimSpace(username) + ":" + password))
	return hex.EncodeToString(sum[:])
}

func (h *Handler) fever(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		badRequestError(c, "invalid request")
		return
	}

	if !hasFeverFlag(c.Request.Form, "api") {
		badRequestError(c, "invalid request")
		return
	}

	if !h.verifyFeverAPIKey(c.Request.Form.Get("api_key")) {
		c.JSON(200, gin.H{"api_version": feverAPIVersion, "auth": 0})
		return
	}

	response := gin.H{
		"auth":                   1,
		"api_version":            feverAPIVersion,
		"last_refreshed_on_time": time.Now().Unix(),
	}

	markResult, msg, err := h.handleFeverMark(c.Request.Form)
	if err != nil {
		internalError(c, err, "mark fever item")
		return
	}
	if msg != "" {
		badRequestError(c, msg)
		return
	}

	if markResult.IncludeUnreadItemIDs {
		ids, err := h.store.ListUnreadItemIDs()
		if err != nil {
			internalError(c, err, "list fever unread item ids")
			return
		}
		response["unread_item_ids"] = joinInt64CSV(ids)
	}

	if markResult.IncludeSavedItemIDs {
		ids, err := h.store.ListSavedItemIDs()
		if err != nil {
			internalError(c, err, "list fever saved item ids")
			return
		}
		response["saved_item_ids"] = joinInt64CSV(ids)
	}

	if hasFeverFlag(c.Request.Form, "groups") {
		groups, feedsGroups, err := h.buildFeverGroupsPayload()
		if err != nil {
			internalError(c, err, "list fever groups")
			return
		}
		response["groups"] = groups
		response["feeds_groups"] = feedsGroups
	}

	if hasFeverFlag(c.Request.Form, "feeds") {
		feeds, feedsGroups, err := h.buildFeverFeedsPayload()
		if err != nil {
			internalError(c, err, "list fever feeds")
			return
		}
		response["feeds"] = feeds
		response["feeds_groups"] = feedsGroups
	}

	if hasFeverFlag(c.Request.Form, "favicons") {
		favicons, err := h.buildFeverFaviconsPayload()
		if err != nil {
			internalError(c, err, "list fever favicons")
			return
		}
		response["favicons"] = favicons
	}

	if hasFeverFlag(c.Request.Form, "items") {
		items, totalItems, err := h.buildFeverItemsPayload(c.Request.Form)
		if err != nil {
			if errors.Is(err, strconv.ErrSyntax) || errors.Is(err, strconv.ErrRange) {
				badRequestError(c, "invalid item filter")
				return
			}
			internalError(c, err, "list fever items")
			return
		}
		response["items"] = items
		response["total_items"] = totalItems
	}

	if hasFeverFlag(c.Request.Form, "unread_item_ids") {
		ids, err := h.store.ListUnreadItemIDs()
		if err != nil {
			internalError(c, err, "list fever unread item ids")
			return
		}
		response["unread_item_ids"] = joinInt64CSV(ids)
	}

	if hasFeverFlag(c.Request.Form, "saved_item_ids") {
		ids, err := h.store.ListSavedItemIDs()
		if err != nil {
			internalError(c, err, "list fever saved item ids")
			return
		}
		response["saved_item_ids"] = joinInt64CSV(ids)
	}

	c.JSON(200, response)
}

func (h *Handler) verifyFeverAPIKey(apiKey string) bool {
	provided := strings.ToLower(strings.TrimSpace(apiKey))
	expected := strings.ToLower(strings.TrimSpace(h.feverAPIKey))
	if provided == "" || expected == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

func (h *Handler) handleFeverMark(form url.Values) (feverMarkResult, string, error) {
	mark := strings.ToLower(strings.TrimSpace(form.Get("mark")))
	if mark == "" {
		return feverMarkResult{}, "", nil
	}

	as := strings.ToLower(strings.TrimSpace(form.Get("as")))

	switch mark {
	case "item":
		ids, err := parseFeverItemIDs(form["id"])
		if err != nil {
			return feverMarkResult{}, "invalid id", nil
		}

		switch as {
		case "read":
			if err := h.store.BatchUpdateItemsUnread(ids, false); err != nil {
				return feverMarkResult{}, "", err
			}
			return feverMarkResult{IncludeUnreadItemIDs: true}, "", nil
		case "unread":
			if err := h.store.BatchUpdateItemsUnread(ids, true); err != nil {
				return feverMarkResult{}, "", err
			}
			return feverMarkResult{IncludeUnreadItemIDs: true}, "", nil
		case "saved":
			for _, id := range ids {
				if err := h.markItemSaved(id); err != nil {
					return feverMarkResult{}, "", err
				}
			}
			return feverMarkResult{IncludeSavedItemIDs: true}, "", nil
		case "unsaved":
			for _, id := range ids {
				if err := h.markItemUnsaved(id); err != nil {
					return feverMarkResult{}, "", err
				}
			}
			return feverMarkResult{IncludeSavedItemIDs: true}, "", nil
		default:
			return feverMarkResult{}, "invalid as", nil
		}
	case "feed":
		id, err := parseFeverRequiredID(form.Get("id"))
		if err != nil {
			return feverMarkResult{}, "invalid id", nil
		}
		before, err := parseFeverBefore(form.Get("before"))
		if err != nil {
			return feverMarkResult{}, "invalid before", nil
		}
		if as != "read" {
			return feverMarkResult{}, "invalid as", nil
		}
		if err := h.store.MarkFeedAsReadBefore(id, before); err != nil {
			return feverMarkResult{}, "", err
		}
		return feverMarkResult{IncludeUnreadItemIDs: true}, "", nil
	case "group":
		id, err := parseFeverRequiredID(form.Get("id"))
		if err != nil {
			return feverMarkResult{}, "invalid id", nil
		}
		before, err := parseFeverBefore(form.Get("before"))
		if err != nil {
			return feverMarkResult{}, "invalid before", nil
		}
		if as != "read" {
			return feverMarkResult{}, "invalid as", nil
		}
		if id == 0 {
			if err := h.store.MarkAllAsReadBefore(before); err != nil {
				return feverMarkResult{}, "", err
			}
			return feverMarkResult{IncludeUnreadItemIDs: true}, "", nil
		}
		if err := h.store.MarkGroupAsReadBefore(id, before); err != nil {
			return feverMarkResult{}, "", err
		}
		return feverMarkResult{IncludeUnreadItemIDs: true}, "", nil
	default:
		return feverMarkResult{}, "invalid mark", nil
	}
}

func (h *Handler) markItemSaved(itemID int64) error {
	item, err := h.store.GetItem(itemID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil
		}
		return err
	}

	exists, err := h.store.BookmarkExists(item.Link)
	if err != nil {
		return err
	}
	if exists {
		return h.store.UpdateBookmarkItemIDByLink(item.ID, item.Link)
	}

	feed, err := h.store.GetFeed(item.FeedID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil
		}
		return err
	}

	_, err = h.store.CreateBookmark(&item.ID, item.Link, item.Title, item.Content, item.PubDate, feed.Name)
	return err
}

func (h *Handler) markItemUnsaved(itemID int64) error {
	item, err := h.store.GetItem(itemID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil
		}
		return err
	}

	return h.store.DeleteBookmarkByLink(item.Link)
}

func (h *Handler) buildFeverGroupsPayload() ([]feverGroup, []feverFeedsGroup, error) {
	groups, err := h.store.ListGroups()
	if err != nil {
		return nil, nil, err
	}

	feeds, err := h.store.ListFeeds()
	if err != nil {
		return nil, nil, err
	}

	resultGroups := make([]feverGroup, 0, len(groups))
	for _, group := range groups {
		resultGroups = append(resultGroups, feverGroup{ID: group.ID, Title: group.Name})
	}

	groupToFeedIDs := make(map[int64][]int64)
	for _, feed := range feeds {
		groupToFeedIDs[feed.GroupID] = append(groupToFeedIDs[feed.GroupID], feed.ID)
	}

	resultFeedGroups := make([]feverFeedsGroup, 0, len(groups))
	for _, group := range groups {
		resultFeedGroups = append(resultFeedGroups, feverFeedsGroup{
			GroupID: group.ID,
			FeedIDs: joinInt64CSV(groupToFeedIDs[group.ID]),
		})
	}

	return resultGroups, resultFeedGroups, nil
}

func (h *Handler) buildFeverFeedsPayload() ([]feverFeed, []feverFeedsGroup, error) {
	feeds, err := h.store.ListFeeds()
	if err != nil {
		return nil, nil, err
	}

	result := make([]feverFeed, 0, len(feeds))
	groupToFeedIDs := make(map[int64][]int64)
	for _, feed := range feeds {
		lastUpdatedOnTime := feed.FetchState.LastSuccessAt
		if lastUpdatedOnTime == 0 {
			lastUpdatedOnTime = feed.UpdatedAt
		}

		groupToFeedIDs[feed.GroupID] = append(groupToFeedIDs[feed.GroupID], feed.ID)

		result = append(result, feverFeed{
			ID:                feed.ID,
			FaviconID:         feed.ID,
			Title:             feed.Name,
			URL:               feed.Link,
			SiteURL:           feed.SiteURL,
			IsSpark:           0,
			LastUpdatedOnTime: lastUpdatedOnTime,
		})
	}

	feedsGroups := make([]feverFeedsGroup, 0, len(groupToFeedIDs))
	for groupID, feedIDs := range groupToFeedIDs {
		feedsGroups = append(feedsGroups, feverFeedsGroup{
			GroupID: groupID,
			FeedIDs: joinInt64CSV(feedIDs),
		})
	}

	return result, feedsGroups, nil
}

func (h *Handler) buildFeverFaviconsPayload() ([]feverFavicon, error) {
	feeds, err := h.store.ListFeeds()
	if err != nil {
		return nil, err
	}

	result := make([]feverFavicon, 0, len(feeds))
	for _, feed := range feeds {
		result = append(result, feverFavicon{ID: feed.ID, Data: feverTransparentGIFData})
	}

	return result, nil
}

func (h *Handler) buildFeverItemsPayload(form url.Values) ([]feverItem, int, error) {
	params, err := parseListFeverItemsParams(form)
	if err != nil {
		return nil, 0, err
	}

	items, err := h.store.ListFeverItems(params)
	if err != nil {
		return nil, 0, err
	}

	totalItems, err := h.store.CountItems(store.ListItemsParams{})
	if err != nil {
		return nil, 0, err
	}

	savedIDs, err := h.store.ListSavedItemIDs()
	if err != nil {
		return nil, 0, err
	}
	savedSet := make(map[int64]struct{}, len(savedIDs))
	for _, id := range savedIDs {
		savedSet[id] = struct{}{}
	}

	result := make([]feverItem, 0, len(items))
	for _, item := range items {
		createdOnTime := item.PubDate
		if createdOnTime == 0 {
			createdOnTime = item.CreatedAt
		}

		_, isSaved := savedSet[item.ID]
		result = append(result, feverItem{
			ID:            item.ID,
			FeedID:        item.FeedID,
			Title:         item.Title,
			Author:        "",
			HTML:          item.Content,
			URL:           item.Link,
			IsSaved:       boolToFeverInt(isSaved),
			IsRead:        boolToFeverInt(!item.Unread),
			CreatedOnTime: createdOnTime,
		})
	}

	return result, totalItems, nil
}

func parseListFeverItemsParams(form url.Values) (store.ListFeverItemsParams, error) {
	withIDs, err := parseFeverCSVInt64(form.Get("with_ids"))
	if err != nil {
		return store.ListFeverItemsParams{}, err
	}

	sinceID, err := parseFeverOptionalInt64(form.Get("since_id"))
	if err != nil {
		return store.ListFeverItemsParams{}, err
	}

	maxID, err := parseFeverOptionalInt64(form.Get("max_id"))
	if err != nil {
		return store.ListFeverItemsParams{}, err
	}

	limit := feverItemsLimit
	if len(withIDs) > 0 {
		if len(withIDs) > feverItemsLimit {
			withIDs = withIDs[:feverItemsLimit]
		}
		limit = 0
	}

	return store.ListFeverItemsParams{
		WithIDs: withIDs,
		SinceID: sinceID,
		MaxID:   maxID,
		Limit:   limit,
		SortAsc: sinceID != nil,
	}, nil
}

func parseFeverRequiredID(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("missing id")
	}

	return strconv.ParseInt(value, 10, 64)
}

func parseFeverBefore(value string) (int64, error) {
	parsed, err := parseFeverOptionalInt64(value)
	if err != nil {
		return 0, err
	}
	if parsed == nil {
		return time.Now().Unix(), nil
	}

	return *parsed, nil
}

func parseFeverItemIDs(values []string) ([]int64, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("missing id")
	}

	ids := []int64{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		parts := strings.Split(value, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			id, err := strconv.ParseInt(part, 10, 64)
			if err != nil {
				return nil, err
			}
			if id <= 0 {
				return nil, fmt.Errorf("invalid id")
			}
			ids = append(ids, id)
		}
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("missing id")
	}

	return ids, nil
}

func parseFeverOptionalInt64(value string) (*int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, err
	}
	if parsed <= 0 {
		return nil, nil
	}

	return &parsed, nil
}

func parseFeverCSVInt64(value string) ([]int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parts := strings.Split(value, ",")
	ids := make([]int64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		id, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func hasFeverFlag(form url.Values, key string) bool {
	if _, exists := form[key]; !exists {
		return false
	}

	value := strings.ToLower(strings.TrimSpace(form.Get(key)))
	if value == "" {
		return true
	}

	switch value {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

func joinInt64CSV(ids []int64) string {
	if len(ids) == 0 {
		return ""
	}

	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.FormatInt(id, 10)
	}

	return strings.Join(parts, ",")
}

func boolToFeverInt(value bool) int {
	if value {
		return 1
	}

	return 0
}
