package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/0x2E/fusion/internal/store"
	"github.com/gin-gonic/gin"
)

// Google Reader (GReader) identifiers. The long item id embeds the int64
// items.id as 16 lowercase hex digits; the short form is the decimal id.
var (
	errGReaderInvalidStream = errors.New("invalid stream id")
	errGReaderInvalidTag    = errors.New("invalid label tag")
)

const (
	greaderItemIDPrefix    = "tag:google.com,2005:reader/item/"
	greaderReadingList     = "user/-/state/com.google/reading-list"
	greaderReadState       = "user/-/state/com.google/read"
	greaderStarredState    = "user/-/state/com.google/starred"
	greaderKeptUnreadState = "user/-/state/com.google/kept-unread"
	greaderLabelPrefix     = "user/-/label/"
	greaderFeedPrefix      = "feed/"
	greaderUserID          = "1"

	greaderDefaultItemLimit = 20
	greaderMaxItemIDsLimit  = 10000
	greaderMaxContentsLimit = 1000
	greaderMaxRequestIDs    = 1000
)

// greaderStreamKind is the normalized target of a GReader stream id.
type greaderStreamKind int

const (
	greaderStreamReadingList greaderStreamKind = iota
	greaderStreamRead
	greaderStreamStarred
	greaderStreamKeptUnread
	greaderStreamFeed
	greaderStreamLabel
)

type greaderStream struct {
	kind   greaderStreamKind
	feedID int64
	label  string
}

// deriveGReaderToken derives the stateless auth token from the configured
// password. Changing the password invalidates every client token; restarts do
// not, which clients rely on to stay signed in.
func deriveGReaderToken(username, password string) string {
	mac := hmac.New(sha256.New, []byte(password))
	_, _ = mac.Write([]byte("greader:" + strings.TrimSpace(username)))
	return hex.EncodeToString(mac.Sum(nil))
}

// constantTimeStringEqual compares two secrets without exposing where they
// differ. Length differences still leak through the length check, which is
// unavoidable for variable-length strings and harmless for fixed-width tokens.
func constantTimeStringEqual(provided, expected string) bool {
	if provided == "" || expected == "" || len(provided) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

// greaderAuthToken extracts the token from an "Authorization: GoogleLogin
// auth=<token>" header, the credential format every GReader client sends.
func greaderAuthToken(header string) string {
	scheme, params, ok := strings.Cut(strings.TrimSpace(header), " ")
	if !ok || !strings.EqualFold(scheme, "GoogleLogin") {
		return ""
	}

	for _, param := range strings.Fields(params) {
		key, value, ok := strings.Cut(param, "=")
		if ok && strings.EqualFold(key, "auth") {
			return value
		}
	}
	return ""
}

func (h *Handler) greaderAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !constantTimeStringEqual(h.greaderRequestToken(c), h.greaderToken) {
			greaderUnauthorized(c)
			c.Abort()
			return
		}
		c.Next()
	}
}

// greaderRequestToken returns the credential a request carries. The header is
// preferred; a POST without it may send the token as the T form value instead,
// because some clients only do that on writes (Miniflux accepts nothing else
// on POST). Both carry the same stateless token.
func (h *Handler) greaderRequestToken(c *gin.Context) string {
	if token := greaderAuthToken(c.GetHeader("Authorization")); token != "" {
		return token
	}
	if c.Request.Method != http.MethodPost {
		return ""
	}
	if err := c.Request.ParseForm(); err != nil {
		return ""
	}
	return c.Request.PostForm.Get("T")
}

// greaderUnimplemented answers endpoints Fusion does not implement, such as
// preference/list or friend/list. Clients probe them during setup and treat a
// 404 as a broken server, so an authenticated request gets an empty 200 JSON
// list, as Miniflux answers.
func (h *Handler) greaderUnimplemented(c *gin.Context) {
	if !constantTimeStringEqual(h.greaderRequestToken(c), h.greaderToken) {
		greaderUnauthorized(c)
		return
	}
	c.JSON(http.StatusOK, []any{})
}

// checkGReaderCredentials validates a ClientLogin request. The password is
// checked by deriving the token from it, so only one secret comparison exists.
// In anonymous mode every credential is accepted, matching the open API.
func (h *Handler) checkGReaderCredentials(email, password string) bool {
	if h.greaderToken == "" {
		return false
	}
	if h.allowAnonAPI {
		return true
	}
	if strings.TrimSpace(email) != h.greaderUsername {
		return false
	}
	return constantTimeStringEqual(deriveGReaderToken(h.greaderUsername, password), h.greaderToken)
}

type greaderClientLoginResponse struct {
	SID  string `json:"SID"`
	LSID string `json:"LSID"`
	Auth string `json:"Auth"`
}

func (h *Handler) greaderClientLogin(c *gin.Context) {
	ip := c.ClientIP()
	allowed, retryAfter := h.limiter.allow(ip, time.Now())
	if !allowed {
		tooManyRequestsError(c, retryAfter)
		return
	}

	if err := c.Request.ParseForm(); err != nil {
		greaderBadRequest(c, "invalid request")
		return
	}

	if !h.checkGReaderCredentials(c.Request.Form.Get("Email"), c.Request.Form.Get("Passwd")) {
		h.limiter.recordFailure(ip, time.Now())
		c.String(http.StatusUnauthorized, "Error=BadAuthentication\n")
		return
	}
	h.limiter.recordSuccess(ip)

	if strings.EqualFold(strings.TrimSpace(c.Request.Form.Get("output")), "json") {
		c.JSON(http.StatusOK, greaderClientLoginResponse{
			SID:  h.greaderToken,
			LSID: h.greaderToken,
			Auth: h.greaderToken,
		})
		return
	}

	c.String(http.StatusOK, "SID=%s\nLSID=%s\nAuth=%s\n", h.greaderToken, h.greaderToken, h.greaderToken)
}

type greaderUserInfoResponse struct {
	UserID        string `json:"userId"`
	UserName      string `json:"userName"`
	UserProfileID string `json:"userProfileId"`
	UserEmail     string `json:"userEmail"`
}

func (h *Handler) greaderUserInfo(c *gin.Context) {
	c.JSON(http.StatusOK, greaderUserInfoResponse{
		UserID:        greaderUserID,
		UserName:      h.greaderUsername,
		UserProfileID: greaderUserID,
		UserEmail:     "",
	})
}

// greaderTokenHandler answers the token endpoint GReader clients call after
// login. Fusion's token is stateless, so it is the same value as ClientLogin.
func (h *Handler) greaderTokenHandler(c *gin.Context) {
	c.String(http.StatusOK, h.greaderToken)
}

type greaderSubscriptionCategory struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type greaderSubscription struct {
	ID         string                        `json:"id"`
	Title      string                        `json:"title"`
	Categories []greaderSubscriptionCategory `json:"categories"`
	URL        string                        `json:"url"`
	HTMLURL    string                        `json:"htmlUrl"`
	IconURL    string                        `json:"iconUrl"`
}

func greaderLabelState(label string) string {
	return greaderLabelPrefix + label
}

func greaderFeedStreamID(feedID int64) string {
	return greaderFeedPrefix + strconv.FormatInt(feedID, 10)
}

func (h *Handler) greaderSubscriptionList(c *gin.Context) {
	feeds, err := h.store.ListFeeds()
	if err != nil {
		internalError(c, err, "list greader subscriptions")
		return
	}
	groups, err := h.store.ListGroups()
	if err != nil {
		internalError(c, err, "list greader groups")
		return
	}

	groupNames := make(map[int64]string, len(groups))
	for _, group := range groups {
		groupNames[group.ID] = group.Name
	}

	subscriptions := make([]greaderSubscription, 0, len(feeds))
	for _, feed := range feeds {
		categories := make([]greaderSubscriptionCategory, 0, 1)
		if name, ok := groupNames[feed.GroupID]; ok {
			categories = append(categories, greaderSubscriptionCategory{
				ID:    greaderLabelState(name),
				Label: name,
			})
		}
		subscriptions = append(subscriptions, greaderSubscription{
			ID:         greaderFeedStreamID(feed.ID),
			Title:      feed.Name,
			Categories: categories,
			URL:        feed.Link,
			HTMLURL:    feed.SiteURL,
			IconURL:    "",
		})
	}

	c.JSON(http.StatusOK, gin.H{"subscriptions": subscriptions})
}

type greaderTag struct {
	ID   string `json:"id"`
	Type string `json:"type,omitempty"`
}

func (h *Handler) greaderTagList(c *gin.Context) {
	groups, err := h.store.ListGroups()
	if err != nil {
		internalError(c, err, "list greader tags")
		return
	}

	tags := make([]greaderTag, 0, len(groups)+1)
	tags = append(tags, greaderTag{ID: greaderStarredState})
	for _, group := range groups {
		tags = append(tags, greaderTag{ID: greaderLabelState(group.Name), Type: "folder"})
	}

	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

type greaderUnreadCount struct {
	ID                      string `json:"id"`
	Count                   int64  `json:"count"`
	NewestItemTimestampUsec string `json:"newestItemTimestampUsec"`
}

func greaderUsec(seconds int64) string {
	if seconds <= 0 {
		return "0"
	}
	return strconv.FormatInt(seconds*1_000_000, 10)
}

func (h *Handler) greaderUnreadCount(c *gin.Context) {
	feeds, err := h.store.ListFeeds()
	if err != nil {
		internalError(c, err, "list greader unread feeds")
		return
	}
	groups, err := h.store.ListGroups()
	if err != nil {
		internalError(c, err, "list greader unread groups")
		return
	}
	counts, err := h.store.GReaderUnreadCounts()
	if err != nil {
		internalError(c, err, "count greader unread items")
		return
	}

	unreadCounts := make([]greaderUnreadCount, 0, len(feeds)+len(groups)+1)
	for _, feed := range feeds {
		summary := counts.ByFeed[feed.ID]
		unreadCounts = append(unreadCounts, greaderUnreadCount{
			ID:                      greaderFeedStreamID(feed.ID),
			Count:                   summary.Count,
			NewestItemTimestampUsec: greaderUsec(summary.NewestCreatedAt),
		})
	}
	for _, group := range groups {
		summary := counts.ByGroup[group.ID]
		unreadCounts = append(unreadCounts, greaderUnreadCount{
			ID:                      greaderLabelState(group.Name),
			Count:                   summary.Count,
			NewestItemTimestampUsec: greaderUsec(summary.NewestCreatedAt),
		})
	}
	unreadCounts = append(unreadCounts, greaderUnreadCount{
		ID:                      greaderReadingList,
		Count:                   counts.Total.Count,
		NewestItemTimestampUsec: greaderUsec(counts.Total.NewestCreatedAt),
	})

	c.JSON(http.StatusOK, gin.H{
		"max":          counts.Total.Count,
		"unreadcounts": unreadCounts,
	})
}

// parseGReaderStreamID normalizes a stream id. Clients sometimes send their
// numeric user id instead of "-", so the user segment is treated as opaque.
func parseGReaderStreamID(raw string) (greaderStream, error) {
	id := strings.TrimPrefix(strings.TrimSpace(raw), "/")
	if id == "" {
		return greaderStream{}, errors.New("missing stream id")
	}

	if rest, ok := strings.CutPrefix(id, greaderFeedPrefix); ok {
		feedID, err := strconv.ParseInt(rest, 10, 64)
		if err != nil || feedID <= 0 {
			return greaderStream{}, errors.New("invalid stream id")
		}
		return greaderStream{kind: greaderStreamFeed, feedID: feedID}, nil
	}

	rest, ok := strings.CutPrefix(id, "user/")
	if !ok {
		return greaderStream{}, errors.New("invalid stream id")
	}
	_, rest, ok = strings.Cut(rest, "/")
	if !ok {
		return greaderStream{}, errors.New("invalid stream id")
	}

	switch {
	case rest == "state/com.google/reading-list":
		return greaderStream{kind: greaderStreamReadingList}, nil
	case rest == "state/com.google/read":
		return greaderStream{kind: greaderStreamRead}, nil
	case rest == "state/com.google/starred":
		return greaderStream{kind: greaderStreamStarred}, nil
	case rest == "state/com.google/kept-unread":
		return greaderStream{kind: greaderStreamKeptUnread}, nil
	}

	if label, ok := strings.CutPrefix(rest, "label/"); ok && label != "" {
		return greaderStream{kind: greaderStreamLabel, label: label}, nil
	}

	return greaderStream{}, errors.New("invalid stream id")
}

func parseGReaderLabelTag(raw string) (string, bool) {
	stream, err := parseGReaderStreamID(raw)
	if err != nil || stream.kind != greaderStreamLabel {
		return "", false
	}
	return stream.label, true
}

// parseGReaderFeedRef accepts the subscription stream shapes: feed/<id> and
// feed/<url>.
func parseGReaderFeedRef(raw string) (feedID int64, link string, err error) {
	id := strings.TrimPrefix(strings.TrimSpace(raw), "/")
	rest, ok := strings.CutPrefix(id, greaderFeedPrefix)
	if !ok || rest == "" {
		return 0, "", errGReaderInvalidStream
	}

	if feedID, err := strconv.ParseInt(rest, 10, 64); err == nil {
		if feedID <= 0 {
			return 0, "", errGReaderInvalidStream
		}
		return feedID, "", nil
	}

	return 0, rest, nil
}

// parseGReaderItemID accepts the long prefixed form, the bare 16-digit hex
// form some clients strip it to, and the short decimal form. A 16-character
// value is always hex: decimal ids that long would need 10^15 items.
func parseGReaderItemID(raw string) (int64, error) {
	value := strings.TrimSpace(raw)
	hexID, prefixed := strings.CutPrefix(value, greaderItemIDPrefix)
	if prefixed || len(value) == 16 {
		id, err := strconv.ParseUint(hexID, 16, 64)
		if err != nil || id == 0 || id > math.MaxInt64 {
			return 0, errors.New("invalid item id")
		}
		return int64(id), nil
	}

	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid item id")
	}
	return id, nil
}

func formatGReaderItemID(id int64) string {
	return fmt.Sprintf("%s%016x", greaderItemIDPrefix, id)
}

func parseGReaderItemIDs(values []string) ([]int64, error) {
	ids := make([]int64, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		id, err := parseGReaderItemID(value)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, errors.New("missing item id")
	}
	return ids, nil
}

// greaderForm parses query and body parameters together: clients send the same
// parameters either way, and POST read endpoints may carry them in the body.
func greaderForm(c *gin.Context) (url.Values, bool) {
	if err := c.Request.ParseForm(); err != nil {
		greaderBadRequest(c, "invalid request")
		return nil, false
	}
	return c.Request.Form, true
}

func greaderBadRequest(c *gin.Context, message string) {
	c.String(http.StatusBadRequest, message)
}

func greaderNotFound(c *gin.Context, message string) {
	c.String(http.StatusNotFound, message)
}

func greaderUnauthorized(c *gin.Context) {
	c.String(http.StatusUnauthorized, "Unauthorized")
}

func (h *Handler) allowPrivateFeeds() bool {
	return h.config != nil && h.config.AllowPrivateFeeds
}

// greaderStoreError maps store failures onto the GReader error contract:
// unknown resources are 404 text, everything else goes through internalError.
func greaderStoreError(c *gin.Context, err error, message, context string) {
	if errors.Is(err, store.ErrNotFound) {
		greaderNotFound(c, message)
		return
	}
	internalError(c, err, context)
}
