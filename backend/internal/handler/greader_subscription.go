package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/0x2E/fusion/internal/model"
	"github.com/0x2E/fusion/internal/pkg/httpc"
	"github.com/0x2E/fusion/internal/store"
	"github.com/gin-gonic/gin"
)

// greaderDiscoveryTimeout matches the feed discovery budget used by
// validateFeed so both entry points behave the same behind a proxy.
const greaderDiscoveryTimeout = 30 * time.Second

type greaderQuickaddResponse struct {
	NumResults int    `json:"numResults"`
	Query      string `json:"query"`
	StreamID   string `json:"streamId,omitempty"`
	StreamName string `json:"streamName,omitempty"`
}

func (h *Handler) greaderSubscriptionEdit(c *gin.Context) {
	form, ok := greaderForm(c)
	if !ok {
		return
	}

	refs := form["s"]
	if len(refs) == 0 {
		greaderBadRequest(c, "missing stream id")
		return
	}

	switch strings.ToLower(strings.TrimSpace(form.Get("ac"))) {
	case "subscribe":
		h.greaderSubscribe(c, form, refs)
	case "unsubscribe":
		h.greaderUnsubscribe(c, refs)
	case "edit":
		h.greaderEditSubscription(c, form, refs)
	default:
		greaderBadRequest(c, "invalid ac")
	}
}

func (h *Handler) greaderSubscribe(c *gin.Context, form url.Values, refs []string) {
	title := strings.TrimSpace(form.Get("t"))

	// Validate every reference before touching the database: a rejected
	// subscribe must not create a feed or an empty label.
	type action struct {
		feed *model.Feed
		link string // set only for a link that is not subscribed yet
	}
	actions := make([]action, 0, len(refs))

	for _, raw := range refs {
		feedID, link, err := parseGReaderFeedRef(raw)
		if err != nil {
			greaderBadRequest(c, err.Error())
			return
		}

		if link == "" {
			feed, err := h.store.GetFeed(feedID)
			if err != nil {
				greaderStoreError(c, err, "feed not found", "get greader feed")
				return
			}
			actions = append(actions, action{feed: feed})
			continue
		}

		feed, err := h.store.GetFeedByLink(link)
		if err == nil {
			actions = append(actions, action{feed: feed})
			continue
		}
		if !errors.Is(err, store.ErrNotFound) {
			internalError(c, err, "look up greader subscription")
			return
		}
		if err := httpc.ValidateRequestURL(c.Request.Context(), link, h.allowPrivateFeeds()); err != nil {
			greaderBadRequest(c, fmt.Sprintf("invalid URL: %v", err))
			return
		}
		actions = append(actions, action{link: link})
	}

	groupID, labelGiven, err := h.greaderSubscriptionGroup(form)
	if err != nil {
		greaderSubscriptionGroupError(c, err)
		return
	}

	for _, action := range actions {
		if action.feed != nil {
			// Subscribing again is idempotent: apply the title and label the
			// client asked for instead of duplicating the feed.
			if err := h.updateGReaderSubscription(action.feed.ID, title, groupID, labelGiven); err != nil {
				internalError(c, err, "update greader subscription")
				return
			}
			continue
		}

		name := title
		if name == "" {
			name = action.link
		}
		if _, err := h.createGReaderFeed(groupID, name, action.link); err != nil {
			internalError(c, err, "create greader feed")
			return
		}
	}

	c.String(http.StatusOK, "OK")
}

func (h *Handler) greaderUnsubscribe(c *gin.Context, refs []string) {
	for _, raw := range refs {
		feed, err := h.resolveGReaderFeed(raw)
		if err != nil {
			if errors.Is(err, errGReaderInvalidStream) {
				greaderBadRequest(c, err.Error())
				return
			}
			// Unsubscribing a feed that is already gone is not an error.
			if errors.Is(err, store.ErrNotFound) {
				continue
			}
			internalError(c, err, "get greader feed")
			return
		}

		if err := h.store.DeleteFeed(feed.ID); err != nil {
			internalError(c, err, "unsubscribe greader feed")
			return
		}
		h.puller.Wake()
	}

	c.String(http.StatusOK, "OK")
}

func (h *Handler) greaderEditSubscription(c *gin.Context, form url.Values, refs []string) {
	title := strings.TrimSpace(form.Get("t"))
	// In the protocol, r on its own removes the subscription from its label;
	// with a single label per feed that means falling back to Default. An empty
	// r value is not a removal.
	hasRemove := false
	for _, raw := range form["r"] {
		if strings.TrimSpace(raw) != "" {
			hasRemove = true
		}
	}

	feeds := make([]*model.Feed, 0, len(refs))
	for _, raw := range refs {
		feed, err := h.resolveGReaderFeed(raw)
		if err != nil {
			if errors.Is(err, errGReaderInvalidStream) {
				greaderBadRequest(c, err.Error())
				return
			}
			greaderStoreError(c, err, "feed not found", "get greader feed")
			return
		}
		feeds = append(feeds, feed)
	}

	groupID, labelGiven, err := h.greaderSubscriptionGroup(form)
	if err != nil {
		greaderSubscriptionGroupError(c, err)
		return
	}

	for _, feed := range feeds {
		if err := h.updateGReaderSubscription(feed.ID, title, groupID, labelGiven || hasRemove); err != nil {
			internalError(c, err, "update greader subscription")
			return
		}
	}

	c.String(http.StatusOK, "OK")
}

func (h *Handler) greaderSubscriptionQuickadd(c *gin.Context) {
	form, ok := greaderForm(c)
	if !ok {
		return
	}

	target := strings.TrimSpace(form.Get("quickadd"))
	if target == "" {
		greaderBadRequest(c, "missing url")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), greaderDiscoveryTimeout)
	defer cancel()

	found, err := discoverFeeds(ctx, target, "", h.allowPrivateFeeds())
	if err != nil {
		greaderBadRequest(c, err.Error())
		return
	}

	response := greaderQuickaddResponse{Query: target}
	if len(found) > 0 {
		// Quickadd is idempotent: an already-subscribed feed is returned as is.
		feed, err := h.store.GetFeedByLink(found[0].Link)
		if errors.Is(err, store.ErrNotFound) {
			if err := httpc.ValidateRequestURL(ctx, found[0].Link, h.allowPrivateFeeds()); err != nil {
				greaderBadRequest(c, fmt.Sprintf("invalid URL: %v", err))
				return
			}
			name := found[0].Title
			if name == "" {
				name = found[0].Link
			}
			feed, err = h.createGReaderFeed(store.DefaultGroupID, name, found[0].Link)
		}
		if err != nil {
			internalError(c, err, "create greader feed")
			return
		}

		response.NumResults = 1
		response.StreamID = greaderFeedStreamID(feed.ID)
		response.StreamName = feed.Name
	}

	c.JSON(http.StatusOK, response)
}

// greaderSubscriptionGroup resolves the label of a subscription edit and
// creates its group. Every non-empty a value must be a label stream: silently
// falling back to Default on a malformed label would move feeds the client
// never asked to move.
func (h *Handler) greaderSubscriptionGroup(form url.Values) (groupID int64, labelGiven bool, err error) {
	label := ""
	for _, raw := range form["a"] {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		name, ok := parseGReaderLabelTag(raw)
		if !ok {
			return 0, false, errGReaderInvalidTag
		}
		if label == "" {
			label = name
		}
	}

	if label == "" {
		return store.DefaultGroupID, false, nil
	}

	group, err := h.getOrCreateGReaderGroup(label)
	if err != nil {
		return 0, false, err
	}
	return group.ID, true, nil
}

func greaderSubscriptionGroupError(c *gin.Context, err error) {
	if errors.Is(err, errGReaderInvalidTag) {
		greaderBadRequest(c, err.Error())
		return
	}
	internalError(c, err, "resolve greader subscription label")
}

func (h *Handler) getOrCreateGReaderGroup(name string) (*model.Group, error) {
	group, err := h.store.GetGroupByName(name)
	if err == nil {
		return group, nil
	}
	if !errors.Is(err, store.ErrNotFound) {
		return nil, err
	}
	return h.store.CreateGroup(name)
}

// resolveGReaderFeed looks a feed up by feed/<id> or feed/<url>. Parse failures
// wrap errGReaderInvalidStream; missing feeds wrap store.ErrNotFound.
func (h *Handler) resolveGReaderFeed(raw string) (*model.Feed, error) {
	feedID, link, err := parseGReaderFeedRef(raw)
	if err != nil {
		return nil, err
	}
	if link == "" {
		return h.store.GetFeed(feedID)
	}
	return h.store.GetFeedByLink(link)
}

func (h *Handler) updateGReaderSubscription(feedID int64, title string, groupID int64, move bool) error {
	params := store.UpdateFeedParams{}
	if title != "" {
		params.Name = &title
	}
	if move {
		params.GroupID = &groupID
	}
	if params.Name == nil && params.GroupID == nil {
		return nil
	}

	if err := h.store.UpdateFeed(feedID, params); err != nil {
		return err
	}
	h.puller.Wake()
	return nil
}

// createGReaderFeed creates a feed whose link was already validated, mirroring
// createFeed, and starts its initial pull in the background.
func (h *Handler) createGReaderFeed(groupID int64, name, link string) (*model.Feed, error) {
	feed, err := h.store.CreateFeed(groupID, name, link, "", "", nil)
	if err != nil {
		return nil, err
	}

	h.startInitialPull(feed.ID)
	return feed, nil
}

func (h *Handler) greaderRenameTag(c *gin.Context) {
	form, ok := greaderForm(c)
	if !ok {
		return
	}

	source, ok := parseGReaderLabelTag(form.Get("s"))
	if !ok {
		greaderBadRequest(c, "invalid source tag")
		return
	}
	dest, ok := parseGReaderLabelTag(form.Get("dest"))
	if !ok {
		greaderBadRequest(c, "invalid dest tag")
		return
	}
	if source == dest {
		c.String(http.StatusOK, "OK")
		return
	}

	sourceGroup, err := h.store.GetGroupByName(source)
	if err != nil {
		greaderStoreError(c, err, "label not found", "get greader label")
		return
	}

	if _, err := h.store.GetGroupByName(dest); err == nil {
		greaderBadRequest(c, "label already exists")
		return
	} else if !errors.Is(err, store.ErrNotFound) {
		internalError(c, err, "get greader label")
		return
	}

	if err := h.store.UpdateGroup(sourceGroup.ID, dest); err != nil {
		internalError(c, err, "rename greader label")
		return
	}

	c.String(http.StatusOK, "OK")
}

func (h *Handler) greaderDisableTag(c *gin.Context) {
	form, ok := greaderForm(c)
	if !ok {
		return
	}

	rawTags := form["s"]
	if len(rawTags) == 0 {
		greaderBadRequest(c, "missing tag")
		return
	}

	// Validate every tag before deleting any, so a bad tag in the request
	// cannot leave the labels half-removed.
	toDelete := make([]int64, 0, len(rawTags))
	seen := make(map[int64]bool, len(rawTags))
	for _, raw := range rawTags {
		label, ok := parseGReaderLabelTag(raw)
		if !ok {
			greaderBadRequest(c, "invalid tag")
			return
		}
		group, err := h.store.GetGroupByName(label)
		if err != nil {
			greaderStoreError(c, err, "label not found", "get greader label")
			return
		}
		if group.ID == store.DefaultGroupID {
			greaderBadRequest(c, "cannot delete default label")
			return
		}
		if seen[group.ID] {
			continue
		}
		seen[group.ID] = true
		toDelete = append(toDelete, group.ID)
	}

	for _, groupID := range toDelete {
		if err := h.store.DeleteGroup(groupID); err != nil {
			internalError(c, err, "delete greader label")
			return
		}
	}

	c.String(http.StatusOK, "OK")
}
