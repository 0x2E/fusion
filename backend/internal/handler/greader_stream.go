package handler

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/0x2E/fusion/internal/store"
	"github.com/gin-gonic/gin"
)

// registerGReaderRoutes mounts the Google Reader API next to Fever, outside
// /api. Read endpoints answer both GET and POST because clients disagree about
// which method to use; write endpoints are POST only, as in the protocol.
func (h *Handler) registerGReaderRoutes(r *gin.Engine) {
	reader := r.Group("/reader/api/0")
	reader.Use(h.greaderAuthMiddleware())
	{
		for _, method := range []string{http.MethodGet, http.MethodPost} {
			reader.Handle(method, "/user-info", h.greaderUserInfo)
			reader.Handle(method, "/token", h.greaderTokenHandler)
			reader.Handle(method, "/subscription/list", h.greaderSubscriptionList)
			reader.Handle(method, "/tag/list", h.greaderTagList)
			reader.Handle(method, "/unread-count", h.greaderUnreadCount)
			reader.Handle(method, "/stream/items/ids", h.greaderStreamItemIDs)
			reader.Handle(method, "/stream/contents/*streamID", h.greaderStreamContents)
			reader.Handle(method, "/stream/items/contents", h.greaderStreamItemsContents)
		}

		reader.POST("/edit-tag", h.greaderEditTag)
		reader.POST("/mark-all-as-read", h.greaderMarkAllAsRead)
		reader.POST("/subscription/edit", h.greaderSubscriptionEdit)
		reader.POST("/subscription/quickadd", h.greaderSubscriptionQuickadd)
		reader.POST("/rename-tag", h.greaderRenameTag)
		reader.POST("/disable-tag", h.greaderDisableTag)
	}
}

type greaderItemContent struct {
	Direction string `json:"direction"`
	Content   string `json:"content"`
}

type greaderItemLink struct {
	Href string `json:"href"`
	Type string `json:"type,omitempty"`
}

type greaderItemOrigin struct {
	StreamID string `json:"streamId"`
	Title    string `json:"title"`
	HTMLURL  string `json:"htmlUrl"`
}

type greaderItem struct {
	ID            string             `json:"id"`
	CrawlTimeMsec string             `json:"crawlTimeMsec"`
	TimestampUsec string             `json:"timestampUsec"`
	Published     int64              `json:"published"`
	Updated       int64              `json:"updated"`
	Title         string             `json:"title"`
	Author        string             `json:"author"`
	Canonical     []greaderItemLink  `json:"canonical"`
	Alternate     []greaderItemLink  `json:"alternate"`
	Summary       greaderItemContent `json:"summary"`
	Categories    []string           `json:"categories"`
	Origin        greaderItemOrigin  `json:"origin"`
}

type greaderItemsPayload struct {
	ID           string        `json:"id"`
	Updated      int64         `json:"updated"`
	Items        []greaderItem `json:"items"`
	Continuation string        `json:"continuation,omitempty"`
}

type greaderItemRef struct {
	ID              string   `json:"id"`
	DirectStreamIDs []string `json:"directStreamIds"`
	TimestampUsec   string   `json:"timestampUsec"`
}

type greaderItemRefsPayload struct {
	ItemRefs     []greaderItemRef `json:"itemRefs"`
	Continuation string           `json:"continuation,omitempty"`
}

func greaderItemPayload(item *store.GReaderItem) greaderItem {
	// Clients sort by published and de-duplicate on timestampUsec, so keep both
	// on the same fallback: pub_date, then crawl time for items without one.
	published := item.PubDate
	if published == 0 {
		published = item.CreatedAt
	}

	categories := []string{
		greaderReadingList,
		greaderLabelState(item.GroupName),
	}
	if !item.Unread {
		categories = append(categories, greaderReadState)
	}
	if item.Starred {
		categories = append(categories, greaderStarredState)
	}

	return greaderItem{
		ID:            formatGReaderItemID(item.ID),
		CrawlTimeMsec: strconv.FormatInt(item.CreatedAt*1000, 10),
		TimestampUsec: greaderUsec(item.CreatedAt),
		Published:     published,
		Updated:       published,
		Title:         item.Title,
		Author:        "",
		Canonical:     []greaderItemLink{{Href: item.Link}},
		Alternate:     []greaderItemLink{{Href: item.Link, Type: "text/html"}},
		Summary:       greaderItemContent{Direction: "ltr", Content: item.Content},
		Categories:    categories,
		Origin: greaderItemOrigin{
			StreamID: greaderFeedStreamID(item.FeedID),
			Title:    item.FeedName,
			HTMLURL:  item.FeedSiteURL,
		},
	}
}

func buildGReaderItemsPayload(streamID string, items []*store.GReaderItem, continuation string) greaderItemsPayload {
	payload := greaderItemsPayload{
		ID:           streamID,
		Updated:      time.Now().Unix(),
		Items:        make([]greaderItem, 0, len(items)),
		Continuation: continuation,
	}
	for _, item := range items {
		payload.Items = append(payload.Items, greaderItemPayload(item))
	}
	return payload
}

// greaderPage trims the extra row used to detect another page and returns the
// continuation cursor, which is just the last returned item id.
func greaderPage[T any](rows []T, limit int, id func(T) int64) ([]T, string) {
	if len(rows) <= limit {
		return rows, ""
	}

	rows = rows[:limit]
	return rows, strconv.FormatInt(id(rows[len(rows)-1]), 10)
}

// prepareGReaderStream resolves a stream id into item filters. On failure it
// writes the error response and returns ok=false.
func (h *Handler) prepareGReaderStream(c *gin.Context, rawStream string, maxLimit int, form url.Values) (store.GReaderItemsParams, bool) {
	stream, err := parseGReaderStreamID(rawStream)
	if err != nil {
		greaderBadRequest(c, err.Error())
		return store.GReaderItemsParams{}, false
	}

	params := store.GReaderItemsParams{}
	switch stream.kind {
	case greaderStreamReadingList:
	case greaderStreamRead:
		params.ReadOnly = true
	case greaderStreamStarred:
		params.Starred = true
	case greaderStreamKeptUnread:
		// Fusion has one read flag, so the kept-unread stream is the unread
		// stream; clients that sync it get the items edit-tag would keep.
		params.UnreadOnly = true
	case greaderStreamFeed:
		if _, err := h.store.GetFeed(stream.feedID); err != nil {
			greaderStoreError(c, err, "feed not found", "get greader feed")
			return store.GReaderItemsParams{}, false
		}
		feedID := stream.feedID
		params.FeedID = &feedID
	case greaderStreamLabel:
		group, err := h.store.GetGroupByName(stream.label)
		if err != nil {
			greaderStoreError(c, err, "label not found", "get greader label")
			return store.GReaderItemsParams{}, false
		}
		params.GroupID = &group.ID
	default:
		greaderBadRequest(c, "unsupported stream")
		return store.GReaderItemsParams{}, false
	}

	if err := applyGReaderStreamQuery(form, &params, maxLimit); err != nil {
		greaderBadRequest(c, err.Error())
		return store.GReaderItemsParams{}, false
	}
	return params, true
}

func applyGReaderStreamQuery(form url.Values, params *store.GReaderItemsParams, maxLimit int) error {
	limit, err := parseGReaderLimit(form.Get("n"), maxLimit)
	if err != nil {
		return err
	}
	// Fetch one extra row so the handler knows whether a continuation exists.
	params.Limit = limit + 1

	// Clients send n (or d, as FreshRSS answers) for newest first and o for
	// oldest first; an absent r means newest first.
	switch strings.TrimSpace(form.Get("r")) {
	case "", "n", "d":
		params.Ascending = false
	case "o":
		params.Ascending = true
	default:
		return errors.New("invalid r")
	}

	if raw := strings.TrimSpace(form.Get("c")); raw != "" {
		cursor, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || cursor <= 0 {
			return errors.New("invalid continuation")
		}
		if params.Ascending {
			params.AfterID = &cursor
		} else {
			params.BeforeID = &cursor
		}
	}

	if params.CreatedAfter, err = parseGReaderOptionalSeconds(form.Get("nt")); err != nil {
		return errors.New("invalid nt")
	}
	if params.CreatedBefore, err = parseGReaderOptionalSeconds(form.Get("ot")); err != nil {
		return errors.New("invalid ot")
	}

	// it/xt only carry tags Fusion can express as item filters; the rest are
	// ignored because Fusion has no per-item labels or broadcasting. The
	// filters intersect with the stream's own state, so a contradictory pair
	// yields an empty page instead of silently overriding one side.
	for _, tag := range form["it"] {
		stream, err := parseGReaderStreamID(tag)
		if err != nil {
			continue
		}
		switch stream.kind {
		case greaderStreamRead:
			params.ReadOnly = true
		case greaderStreamStarred:
			params.Starred = true
		}
	}
	for _, tag := range form["xt"] {
		stream, err := parseGReaderStreamID(tag)
		if err != nil {
			continue
		}
		switch stream.kind {
		case greaderStreamRead:
			params.UnreadOnly = true
		case greaderStreamStarred:
			params.NotStarred = true
		}
	}

	return nil
}

func parseGReaderLimit(raw string, maxLimit int) (int, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return min(greaderDefaultItemLimit, maxLimit), nil
	}

	limit, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New("invalid n")
	}
	return min(max(limit, 1), maxLimit), nil
}

func parseGReaderOptionalSeconds(raw string) (*int64, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}

	seconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, errors.New("invalid timestamp")
	}
	// Zero and negative bounds mean "no bound"; clients send nt=0 on the first
	// sync, which must return everything rather than fail.
	if seconds <= 0 {
		return nil, nil
	}
	return &seconds, nil
}

func (h *Handler) greaderStreamItemIDs(c *gin.Context) {
	form, ok := greaderForm(c)
	if !ok {
		return
	}

	rawStream := strings.TrimSpace(form.Get("s"))
	if rawStream == "" {
		greaderBadRequest(c, "missing stream id")
		return
	}

	params, ok := h.prepareGReaderStream(c, rawStream, greaderMaxItemIDsLimit, form)
	if !ok {
		return
	}

	refs, err := h.store.ListGReaderItemRefs(params)
	if err != nil {
		internalError(c, err, "list greader item ids")
		return
	}

	refs, continuation := greaderPage(refs, params.Limit-1, func(ref store.GReaderItemRef) int64 {
		return ref.ID
	})

	itemRefs := make([]greaderItemRef, 0, len(refs))
	for _, ref := range refs {
		itemRefs = append(itemRefs, greaderItemRef{
			ID:              strconv.FormatInt(ref.ID, 10),
			DirectStreamIDs: []string{},
			TimestampUsec:   greaderUsec(ref.CreatedAt),
		})
	}

	c.JSON(http.StatusOK, greaderItemRefsPayload{ItemRefs: itemRefs, Continuation: continuation})
}

func (h *Handler) greaderStreamContents(c *gin.Context) {
	form, ok := greaderForm(c)
	if !ok {
		return
	}

	rawStream := strings.TrimPrefix(c.Param("streamID"), "/")
	params, ok := h.prepareGReaderStream(c, rawStream, greaderMaxContentsLimit, form)
	if !ok {
		return
	}

	items, err := h.store.ListGReaderItems(params)
	if err != nil {
		internalError(c, err, "list greader stream contents")
		return
	}

	items, continuation := greaderPage(items, params.Limit-1, func(item *store.GReaderItem) int64 {
		return item.ID
	})

	c.JSON(http.StatusOK, buildGReaderItemsPayload(rawStream, items, continuation))
}

func (h *Handler) greaderStreamItemsContents(c *gin.Context) {
	form, ok := greaderForm(c)
	if !ok {
		return
	}

	ids, err := parseGReaderItemIDs(form["i"])
	if err != nil {
		greaderBadRequest(c, err.Error())
		return
	}
	if len(ids) > greaderMaxRequestIDs {
		greaderBadRequest(c, "too many item ids")
		return
	}

	items, err := h.store.ListGReaderItemsByIDs(ids)
	if err != nil {
		internalError(c, err, "list greader items by ids")
		return
	}

	byID := make(map[int64]*store.GReaderItem, len(items))
	for _, item := range items {
		byID[item.ID] = item
	}
	ordered := make([]*store.GReaderItem, 0, len(items))
	for _, id := range ids {
		if item, ok := byID[id]; ok {
			ordered = append(ordered, item)
		}
	}

	c.JSON(http.StatusOK, buildGReaderItemsPayload(greaderReadingList, ordered, ""))
}

func (h *Handler) greaderEditTag(c *gin.Context) {
	form, ok := greaderForm(c)
	if !ok {
		return
	}

	ids, err := parseGReaderItemIDs(form["i"])
	if err != nil {
		greaderBadRequest(c, err.Error())
		return
	}

	if err := h.applyGReaderTagEdits(ids, form); err != nil {
		internalError(c, err, "edit greader tags")
		return
	}

	c.String(http.StatusOK, "OK")
}

// applyGReaderTagEdits translates add/remove tag lists onto Fusion state.
// Fusion has a single read flag and a single starred flag, so read-state tags
// win in this order: kept-unread add and read remove mark unread; read add and
// kept-unread remove mark read. Every other tag is ignored.
func (h *Handler) applyGReaderTagEdits(ids []int64, form url.Values) error {
	var markRead, markUnread, star, unstar bool

	for _, tag := range form["a"] {
		stream, err := parseGReaderStreamID(tag)
		if err != nil {
			continue
		}
		switch stream.kind {
		case greaderStreamRead:
			markRead = true
		case greaderStreamKeptUnread:
			markUnread = true
		case greaderStreamStarred:
			star = true
		}
	}
	for _, tag := range form["r"] {
		stream, err := parseGReaderStreamID(tag)
		if err != nil {
			continue
		}
		switch stream.kind {
		case greaderStreamRead:
			markUnread = true
		case greaderStreamKeptUnread:
			markRead = true
		case greaderStreamStarred:
			unstar = true
		}
	}

	if len(ids) == 0 {
		return nil
	}

	switch {
	case markUnread:
		if err := h.store.BatchUpdateItemsUnread(ids, true); err != nil {
			return err
		}
	case markRead:
		if err := h.store.BatchUpdateItemsUnread(ids, false); err != nil {
			return err
		}
	}

	if star {
		for _, id := range ids {
			if err := h.markItemSaved(id); err != nil {
				return err
			}
		}
	}
	if unstar {
		for _, id := range ids {
			if err := h.markItemUnsaved(id); err != nil {
				return err
			}
		}
	}

	return nil
}

func (h *Handler) greaderMarkAllAsRead(c *gin.Context) {
	form, ok := greaderForm(c)
	if !ok {
		return
	}

	rawStream := strings.TrimSpace(form.Get("s"))
	if rawStream == "" {
		greaderBadRequest(c, "missing stream id")
		return
	}

	stream, err := parseGReaderStreamID(rawStream)
	if err != nil {
		greaderBadRequest(c, err.Error())
		return
	}

	params := store.MarkGReaderItemsReadParams{CreatedAt: time.Now().Unix()}
	switch stream.kind {
	case greaderStreamReadingList:
	case greaderStreamFeed:
		if _, err := h.store.GetFeed(stream.feedID); err != nil {
			greaderStoreError(c, err, "feed not found", "get greader feed")
			return
		}
		params.FeedID = &stream.feedID
	case greaderStreamLabel:
		group, err := h.store.GetGroupByName(stream.label)
		if err != nil {
			greaderStoreError(c, err, "label not found", "get greader label")
			return
		}
		params.GroupID = &group.ID
	default:
		greaderBadRequest(c, "unsupported stream")
		return
	}

	if raw := strings.TrimSpace(form.Get("ts")); raw != "" {
		usec, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || usec < 0 {
			greaderBadRequest(c, "invalid ts")
			return
		}
		params.CreatedAt = usec / 1_000_000
	}

	if err := h.store.MarkGReaderItemsRead(params); err != nil {
		internalError(c, err, "mark greader items read")
		return
	}

	c.String(http.StatusOK, "OK")
}
