package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/0x2E/fusion/internal/config"
	"github.com/0x2E/fusion/internal/model"
	"github.com/0x2E/fusion/internal/store"
	"github.com/gin-gonic/gin"
)

func newGReaderTestHandler(t *testing.T) (*Handler, *store.Store) {
	t.Helper()

	return newGReaderTestHandlerWithConfig(t, &config.Config{})
}

func newGReaderTestHandlerWithConfig(t *testing.T, cfg *config.Config) (*Handler, *store.Store) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	st, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}

	if cfg.Password == "" && cfg.OIDCIssuer == "" {
		cfg.Password = "secret"
	}
	if cfg.GReaderUsername == "" {
		cfg.GReaderUsername = "fusion"
	}
	if cfg.PullTimeout == 0 {
		cfg.PullTimeout = 30
	}
	if cfg.LoginRateLimit == 0 {
		cfg.LoginRateLimit = 10
	}
	if cfg.LoginWindow == 0 {
		cfg.LoginWindow = 60
	}
	if cfg.LoginBlock == 0 {
		cfg.LoginBlock = 300
	}

	h, err := New(st, cfg, noopPuller{})
	if err != nil {
		_ = st.Close()
		t.Fatalf("new handler: %v", err)
	}

	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("close store: %v", err)
		}
	})

	return h, st
}

func parseClientLoginAuth(t *testing.T, body string) string {
	t.Helper()

	for _, line := range strings.Split(body, "\n") {
		if token, ok := strings.CutPrefix(line, "Auth="); ok {
			return token
		}
	}

	t.Fatalf("expected an Auth line, got %q", body)
	return ""
}

func greaderClientLogin(t *testing.T, r http.Handler, email, password string) string {
	t.Helper()

	w := performRequest(
		r,
		http.MethodPost,
		"/accounts/ClientLogin",
		strings.NewReader(url.Values{"Email": {email}, "Passwd": {password}}.Encode()),
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected ClientLogin to succeed, got %d: %s", w.Code, w.Body.String())
	}

	return parseClientLoginAuth(t, w.Body.String())
}

// performGReaderRequest sends a form request with the GoogleLogin header when
// token is not empty. values may be nil for GET requests without parameters.
func performGReaderRequest(
	r http.Handler,
	method, target, token string,
	values url.Values,
) *httptest.ResponseRecorder {
	headers := map[string]string{}
	var body io.Reader
	if values != nil {
		body = strings.NewReader(values.Encode())
		headers["Content-Type"] = "application/x-www-form-urlencoded"
	}
	if token != "" {
		headers["Authorization"] = "GoogleLogin auth=" + token
	}

	return performRequest(r, method, target, body, headers)
}

func TestGReaderClientLoginAndIdentity(t *testing.T) {
	h, _ := newGReaderTestHandler(t)
	router := h.SetupRouter()

	t.Run("text response carries the same token three times", func(t *testing.T) {
		w := performRequest(
			router,
			http.MethodPost,
			"/accounts/ClientLogin",
			strings.NewReader(url.Values{"Email": {"fusion"}, "Passwd": {"secret"}}.Encode()),
			map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		if contentType := w.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "text/plain") {
			t.Fatalf("expected text/plain, got %q", contentType)
		}

		body := w.Body.String()
		sid := strings.TrimPrefix(strings.SplitN(body, "\n", 2)[0], "SID=")
		if sid == "" || sid == body {
			t.Fatalf("expected a SID line, got %q", body)
		}
		for _, prefix := range []string{"SID=", "LSID=", "Auth="} {
			if !strings.Contains(body, prefix+sid) {
				t.Fatalf("expected %s to carry %q, got %q", prefix, sid, body)
			}
		}
	})

	t.Run("query parameters are accepted", func(t *testing.T) {
		w := performRequest(
			router,
			http.MethodPost,
			"/accounts/ClientLogin?Email=fusion&Passwd=secret",
			nil,
			nil,
		)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("output=json", func(t *testing.T) {
		w := performRequest(
			router,
			http.MethodPost,
			"/accounts/ClientLogin",
			strings.NewReader(url.Values{"Email": {"fusion"}, "Passwd": {"secret"}, "output": {"json"}}.Encode()),
			map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var payload greaderClientLoginResponse
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatalf("unmarshal response: %v (body: %s)", err, w.Body.String())
		}
		if payload.SID == "" || payload.SID != payload.LSID || payload.SID != payload.Auth {
			t.Fatalf("expected identical tokens, got %+v", payload)
		}
	})

	t.Run("user-info", func(t *testing.T) {
		token := greaderClientLogin(t, router, "fusion", "secret")
		w := performGReaderRequest(router, http.MethodGet, "/reader/api/0/user-info", token, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var payload greaderUserInfoResponse
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if payload.UserID != "1" || payload.UserProfileID != "1" || payload.UserName != "fusion" || payload.UserEmail != "" {
			t.Fatalf("unexpected user-info payload %+v", payload)
		}
	})

	t.Run("token endpoint returns the same token", func(t *testing.T) {
		token := greaderClientLogin(t, router, "fusion", "secret")
		w := performGReaderRequest(router, http.MethodGet, "/reader/api/0/token", token, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		if got := w.Body.String(); got != token {
			t.Fatalf("expected token %q, got %q", token, got)
		}
	})

	t.Run("bad credentials", func(t *testing.T) {
		for _, creds := range []url.Values{
			{"Email": {"fusion"}, "Passwd": {"wrong"}},
			{"Email": {"wrong"}, "Passwd": {"secret"}},
			{},
		} {
			w := performRequest(
				router,
				http.MethodPost,
				"/accounts/ClientLogin",
				strings.NewReader(creds.Encode()),
				map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected status 401 for %v, got %d", creds, w.Code)
			}
			if got := w.Body.String(); got != "Error=BadAuthentication\n" {
				t.Fatalf("expected BadAuthentication body, got %q", got)
			}
		}
	})
}

func TestGReaderRejectsMissingOrInvalidToken(t *testing.T) {
	h, _ := newGReaderTestHandler(t)
	router := h.SetupRouter()
	token := greaderClientLogin(t, router, "fusion", "secret")

	tests := []struct {
		name    string
		method  string
		target  string
		headers map[string]string
	}{
		{name: "no header", method: http.MethodGet, target: "/reader/api/0/user-info"},
		{name: "wrong scheme", method: http.MethodGet, target: "/reader/api/0/user-info", headers: map[string]string{"Authorization": "Bearer " + token}},
		{name: "wrong token", method: http.MethodGet, target: "/reader/api/0/user-info", headers: map[string]string{"Authorization": "GoogleLogin auth=nope"}},
		{name: "stream", method: http.MethodGet, target: "/reader/api/0/stream/items/ids?s=user/-/state/com.google/reading-list"},
		{name: "write", method: http.MethodPost, target: "/reader/api/0/edit-tag"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(router, tt.method, tt.target, nil, tt.headers)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected status 401, got %d", w.Code)
			}
			if got := w.Body.String(); got != "Unauthorized" {
				t.Fatalf("expected Unauthorized body, got %q", got)
			}
		})
	}
}

func TestGReaderRateLimitsFailedClientLogin(t *testing.T) {
	h, _ := newGReaderTestHandler(t)
	router := h.SetupRouter()

	login := func() *httptest.ResponseRecorder {
		return performRequest(
			router,
			http.MethodPost,
			"/accounts/ClientLogin",
			strings.NewReader(url.Values{"Email": {"fusion"}, "Passwd": {"wrong"}}.Encode()),
			map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		)
	}

	for i := 0; i < 10; i++ {
		if w := login(); w.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: expected 401, got %d", i+1, w.Code)
		}
	}

	w := login()
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after the limit, got %d", w.Code)
	}
	if retryAfter := w.Header().Get("Retry-After"); retryAfter == "" {
		t.Fatal("expected a Retry-After header")
	}
}

func TestGReaderOIDCOnlyRejectsEverything(t *testing.T) {
	discovery := newOIDCTestServer(t)
	h := newAuthModeTestHandler(t, &config.Config{
		OIDCIssuer:       discovery.URL,
		OIDCClientID:     "fusion",
		OIDCClientSecret: "x",
	})
	router := h.SetupRouter()

	for _, password := range []string{"", "secret"} {
		w := performRequest(
			router,
			http.MethodPost,
			"/accounts/ClientLogin",
			strings.NewReader(url.Values{"Email": {"fusion"}, "Passwd": {password}}.Encode()),
			map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for password %q, got %d", password, w.Code)
		}
	}

	// No token can exist, so even a well-formed header is rejected.
	w := performRequest(
		router,
		http.MethodGet,
		"/reader/api/0/user-info",
		nil,
		map[string]string{"Authorization": "GoogleLogin auth=" + deriveGReaderToken("fusion", "")},
	)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected reader endpoints to be disabled, got %d", w.Code)
	}
}

func TestGReaderAnonymousModeAcceptsAnyCredentials(t *testing.T) {
	h := newAuthModeTestHandler(t, &config.Config{})
	router := h.SetupRouter()

	for _, creds := range []url.Values{
		{"Email": {"fusion"}, "Passwd": {""}},
		{"Email": {"someone"}, "Passwd": {"anything"}},
		{},
	} {
		w := performRequest(
			router,
			http.MethodPost,
			"/accounts/ClientLogin",
			strings.NewReader(creds.Encode()),
			map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		)
		if w.Code != http.StatusOK {
			t.Fatalf("expected anonymous ClientLogin to accept %v, got %d: %s", creds, w.Code, w.Body.String())
		}
		if got := parseClientLoginAuth(t, w.Body.String()); got != deriveGReaderToken("fusion", "") {
			t.Fatalf("expected the empty-password token, got %q", got)
		}
	}

	token := deriveGReaderToken("fusion", "")
	w := performGReaderRequest(router, http.MethodGet, "/reader/api/0/user-info", token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected anonymous reader access, got %d", w.Code)
	}
}

func mustCreateGReaderGroup(t *testing.T, st *store.Store, name string) *model.Group {
	t.Helper()

	group, err := st.CreateGroup(name)
	if err != nil {
		t.Fatalf("create group %q: %v", name, err)
	}
	return group
}

func mustCreateGReaderFeed(t *testing.T, st *store.Store, groupID int64, name, link, siteURL string) *model.Feed {
	t.Helper()

	feed, err := st.CreateFeed(groupID, name, link, siteURL, "", nil)
	if err != nil {
		t.Fatalf("create feed %q: %v", name, err)
	}
	return feed
}

func mustCreateGReaderItem(t *testing.T, st *store.Store, feedID int64, title string, pubDate int64) *model.Item {
	t.Helper()

	link := "https://example.com/" + strings.ReplaceAll(title, " ", "-")
	item, err := st.CreateItem(feedID, "guid-"+title, title, link, "<p>"+title+"</p>", pubDate)
	if err != nil {
		t.Fatalf("create item %q: %v", title, err)
	}
	return item
}

func TestGReaderSubscriptionAndTagLists(t *testing.T) {
	h, st := newGReaderTestHandler(t)
	router := h.SetupRouter()
	token := greaderClientLogin(t, router, "fusion", "secret")

	tech := mustCreateGReaderGroup(t, st, "Tech")
	defaultGroup, err := st.GetGroup(store.DefaultGroupID)
	if err != nil {
		t.Fatalf("get default group: %v", err)
	}
	feed1 := mustCreateGReaderFeed(t, st, tech.ID, "Fusion Feed", "https://example.com/rss.xml", "https://example.com")
	feed2 := mustCreateGReaderFeed(t, st, defaultGroup.ID, "Default Feed", "https://example.org/rss.xml", "https://example.org")

	t.Run("subscription list", func(t *testing.T) {
		w := performGReaderRequest(router, http.MethodGet, "/reader/api/0/subscription/list", token, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var payload struct {
			Subscriptions []greaderSubscription `json:"subscriptions"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if len(payload.Subscriptions) != 2 {
			t.Fatalf("expected 2 subscriptions, got %d", len(payload.Subscriptions))
		}

		first := payload.Subscriptions[0]
		if first.ID != greaderFeedStreamID(feed1.ID) || first.Title != feed1.Name {
			t.Fatalf("unexpected subscription %+v", first)
		}
		if first.URL != feed1.Link || first.HTMLURL != feed1.SiteURL || first.IconURL != "" {
			t.Fatalf("unexpected subscription urls %+v", first)
		}
		if len(first.Categories) != 1 || first.Categories[0].ID != "user/-/label/Tech" || first.Categories[0].Label != "Tech" {
			t.Fatalf("unexpected categories %+v", first.Categories)
		}

		// Read endpoints also accept POST.
		w = performGReaderRequest(router, http.MethodPost, "/reader/api/0/subscription/list", token, url.Values{})
		if w.Code != http.StatusOK {
			t.Fatalf("expected POST status 200, got %d", w.Code)
		}
	})

	t.Run("tag list includes every label, even empty ones", func(t *testing.T) {
		mustCreateGReaderGroup(t, st, "Archive")

		w := performGReaderRequest(router, http.MethodGet, "/reader/api/0/tag/list", token, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var payload struct {
			Tags []greaderTag `json:"tags"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}

		labels := map[string]string{}
		hasStarred := false
		for _, tag := range payload.Tags {
			if tag.ID == greaderStarredState {
				hasStarred = true
				continue
			}
			labels[tag.ID] = tag.Type
		}
		if !hasStarred {
			t.Fatal("expected the starred tag")
		}
		for _, name := range []string{"Default", "Tech", "Archive"} {
			if got, ok := labels["user/-/label/"+name]; !ok || got != "folder" {
				t.Fatalf("expected folder tag for %q, got %q (present=%v)", name, got, ok)
			}
		}
		_ = feed2
	})
}

func TestGReaderUnreadCount(t *testing.T) {
	h, st := newGReaderTestHandler(t)
	router := h.SetupRouter()
	token := greaderClientLogin(t, router, "fusion", "secret")

	tech := mustCreateGReaderGroup(t, st, "Tech")
	feed1 := mustCreateGReaderFeed(t, st, tech.ID, "Feed 1", "https://example.com/1.xml", "https://example.com")
	feed2 := mustCreateGReaderFeed(t, st, tech.ID, "Feed 2", "https://example.com/2.xml", "https://example.com")

	mustCreateGReaderItem(t, st, feed1.ID, "One", 100)
	item2 := mustCreateGReaderItem(t, st, feed1.ID, "Two", 200)
	if err := st.BatchUpdateItemsUnread([]int64{item2.ID}, false); err != nil {
		t.Fatalf("mark item read: %v", err)
	}

	w := performGReaderRequest(router, http.MethodGet, "/reader/api/0/unread-count", token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var payload struct {
		Max          int64                `json:"max"`
		UnreadCounts []greaderUnreadCount `json:"unreadcounts"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload.Max != 1 {
		t.Fatalf("expected max=1, got %d", payload.Max)
	}

	byID := map[string]greaderUnreadCount{}
	for _, entry := range payload.UnreadCounts {
		byID[entry.ID] = entry
	}
	if got := byID[greaderFeedStreamID(feed1.ID)]; got.Count != 1 || got.NewestItemTimestampUsec == "0" {
		t.Fatalf("unexpected feed1 count %+v", got)
	}
	if got := byID[greaderFeedStreamID(feed2.ID)]; got.Count != 0 {
		t.Fatalf("unexpected feed2 count %+v", got)
	}
	if got := byID["user/-/label/Tech"]; got.Count != 1 {
		t.Fatalf("unexpected label count %+v", got)
	}
	if got := byID[greaderReadingList]; got.Count != 1 {
		t.Fatalf("unexpected reading-list count %+v", got)
	}
}

func TestGReaderStreamItemIDsPagination(t *testing.T) {
	h, st := newGReaderTestHandler(t)
	router := h.SetupRouter()
	token := greaderClientLogin(t, router, "fusion", "secret")

	group := mustCreateGReaderGroup(t, st, "Tech")
	feed := mustCreateGReaderFeed(t, st, group.ID, "Feed", "https://example.com/rss.xml", "https://example.com")
	first := mustCreateGReaderItem(t, st, feed.ID, "One", 100)
	mustCreateGReaderItem(t, st, feed.ID, "Two", 200)
	third := mustCreateGReaderItem(t, st, feed.ID, "Three", 300)

	decodeRefs := func(t *testing.T, w *httptest.ResponseRecorder) greaderItemRefsPayload {
		t.Helper()
		var payload greaderItemRefsPayload
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatalf("unmarshal response: %v (body: %s)", err, w.Body.String())
		}
		return payload
	}

	w := performGReaderRequest(
		router, http.MethodGet,
		"/reader/api/0/stream/items/ids?s="+url.QueryEscape(greaderReadingList)+"&n=2",
		token, nil,
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	firstPage := decodeRefs(t, w)
	if len(firstPage.ItemRefs) != 2 {
		t.Fatalf("expected 2 refs, got %d", len(firstPage.ItemRefs))
	}
	if firstPage.Continuation == "" {
		t.Fatal("expected a continuation on the first page")
	}
	if firstPage.ItemRefs[0].TimestampUsec == "" || firstPage.ItemRefs[0].DirectStreamIDs == nil {
		t.Fatalf("unexpected ref payload %+v", firstPage.ItemRefs[0])
	}

	w = performGReaderRequest(
		router, http.MethodGet,
		"/reader/api/0/stream/items/ids?s="+url.QueryEscape(greaderReadingList)+"&n=2&c="+firstPage.Continuation,
		token, nil,
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	secondPage := decodeRefs(t, w)
	if len(secondPage.ItemRefs) != 1 {
		t.Fatalf("expected 1 ref, got %d", len(secondPage.ItemRefs))
	}
	if secondPage.Continuation != "" {
		t.Fatalf("expected no continuation on the last page, got %q", secondPage.Continuation)
	}
	if secondPage.ItemRefs[0].ID != strconv.FormatInt(first.ID, 10) {
		t.Fatalf("expected oldest item last, got %q", secondPage.ItemRefs[0].ID)
	}

	t.Run("oldest first with r=o", func(t *testing.T) {
		w := performGReaderRequest(
			router, http.MethodGet,
			"/reader/api/0/stream/items/ids?s="+url.QueryEscape(greaderReadingList)+"&r=o&n=1",
			token, nil,
		)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		page := decodeRefs(t, w)
		if len(page.ItemRefs) != 1 || page.ItemRefs[0].ID != strconv.FormatInt(first.ID, 10) {
			t.Fatalf("expected the oldest item first, got %+v", page.ItemRefs)
		}
	})

	t.Run("POST works for read endpoints", func(t *testing.T) {
		w := performGReaderRequest(
			router, http.MethodPost, "/reader/api/0/stream/items/ids", token,
			url.Values{"s": {greaderReadingList}},
		)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("nt=0 and ot=0 mean no bound", func(t *testing.T) {
		for _, param := range []string{"nt=0", "ot=0"} {
			w := performGReaderRequest(
				router, http.MethodGet,
				"/reader/api/0/stream/items/ids?s="+url.QueryEscape(greaderReadingList)+"&"+param,
				token, nil,
			)
			if w.Code != http.StatusOK {
				t.Fatalf("%s: expected status 200, got %d: %s", param, w.Code, w.Body.String())
			}
			if page := decodeRefs(t, w); len(page.ItemRefs) != 3 {
				t.Fatalf("%s: expected all 3 refs, got %d", param, len(page.ItemRefs))
			}
		}
	})

	t.Run("r=d means newest first", func(t *testing.T) {
		w := performGReaderRequest(
			router, http.MethodGet,
			"/reader/api/0/stream/items/ids?s="+url.QueryEscape(greaderReadingList)+"&r=d&n=1",
			token, nil,
		)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
		page := decodeRefs(t, w)
		if len(page.ItemRefs) != 1 || page.ItemRefs[0].ID != strconv.FormatInt(third.ID, 10) {
			t.Fatalf("expected the newest item first, got %+v", page.ItemRefs)
		}
	})
}

func TestGReaderStreamContents(t *testing.T) {
	h, st := newGReaderTestHandler(t)
	router := h.SetupRouter()
	token := greaderClientLogin(t, router, "fusion", "secret")

	group := mustCreateGReaderGroup(t, st, "Tech")
	feed := mustCreateGReaderFeed(t, st, group.ID, "Fusion Feed", "https://example.com/rss.xml", "https://example.com")
	readItem := mustCreateGReaderItem(t, st, feed.ID, "Read", 100)
	mustCreateGReaderItem(t, st, feed.ID, "Unread", 200)
	starredItem := mustCreateGReaderItem(t, st, feed.ID, "Starred", 300)

	if err := st.BatchUpdateItemsUnread([]int64{readItem.ID}, false); err != nil {
		t.Fatalf("mark read: %v", err)
	}
	if _, err := st.CreateBookmark(&starredItem.ID, &starredItem.FeedID, starredItem.Link, starredItem.Title, starredItem.Content, starredItem.PubDate, feed.Name); err != nil {
		t.Fatalf("bookmark item: %v", err)
	}

	fetch := func(t *testing.T, target string) greaderItemsPayload {
		t.Helper()
		w := performGReaderRequest(router, http.MethodGet, target, token, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
		var payload greaderItemsPayload
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		return payload
	}

	t.Run("reading list payload", func(t *testing.T) {
		payload := fetch(t, "/reader/api/0/stream/contents/"+greaderReadingList)
		if payload.ID != greaderReadingList || payload.Updated == 0 {
			t.Fatalf("unexpected payload header %+v", payload)
		}
		if len(payload.Items) != 3 {
			t.Fatalf("expected 3 items, got %d", len(payload.Items))
		}

		item := payload.Items[0]
		if item.ID != formatGReaderItemID(starredItem.ID) {
			t.Fatalf("expected long item id, got %q", item.ID)
		}
		if item.CrawlTimeMsec != strconv.FormatInt(starredItem.CreatedAt*1000, 10) {
			t.Fatalf("unexpected crawlTimeMsec %q", item.CrawlTimeMsec)
		}
		if item.TimestampUsec != strconv.FormatInt(starredItem.CreatedAt*1_000_000, 10) {
			t.Fatalf("unexpected timestampUsec %q", item.TimestampUsec)
		}
		if item.Published != 300 || item.Updated != 300 {
			t.Fatalf("expected pub_date for published/updated, got %+v", item)
		}
		if item.Author != "" || item.Summary.Direction != "ltr" || item.Summary.Content == "" {
			t.Fatalf("unexpected item body %+v", item)
		}
		if len(item.Canonical) != 1 || item.Canonical[0].Href != starredItem.Link {
			t.Fatalf("unexpected canonical links %+v", item.Canonical)
		}
		if len(item.Alternate) != 1 || item.Alternate[0].Type != "text/html" {
			t.Fatalf("unexpected alternate links %+v", item.Alternate)
		}
		if item.Origin.StreamID != greaderFeedStreamID(feed.ID) || item.Origin.Title != feed.Name || item.Origin.HTMLURL != feed.SiteURL {
			t.Fatalf("unexpected origin %+v", item.Origin)
		}

		categories := strings.Join(item.Categories, ",")
		for _, want := range []string{greaderReadingList, "user/-/label/Tech", greaderStarredState} {
			if !strings.Contains(categories, want) {
				t.Fatalf("expected category %q in %v", want, item.Categories)
			}
		}

		readPayload := fetch(t, "/reader/api/0/stream/contents/user/-/state/com.google/read")
		if len(readPayload.Items) != 1 || readPayload.Items[0].ID != formatGReaderItemID(readItem.ID) {
			t.Fatalf("unexpected read stream %+v", readPayload.Items)
		}
		if !slices.Contains(readPayload.Items[0].Categories, greaderReadState) {
			t.Fatalf("expected read category, got %v", readPayload.Items[0].Categories)
		}

		starredPayload := fetch(t, "/reader/api/0/stream/contents/user/-/state/com.google/starred")
		if len(starredPayload.Items) != 1 || starredPayload.Items[0].ID != formatGReaderItemID(starredItem.ID) {
			t.Fatalf("unexpected starred stream %+v", starredPayload.Items)
		}
		if slices.Contains(starredPayload.Items[0].Categories, greaderReadState) {
			t.Fatalf("expected unread item to lack the read category, got %v", starredPayload.Items[0].Categories)
		}
	})

	t.Run("xt=read excludes read items", func(t *testing.T) {
		payload := fetch(t, "/reader/api/0/stream/contents/"+greaderReadingList+"?xt="+url.QueryEscape(greaderReadState))
		if len(payload.Items) != 2 {
			t.Fatalf("expected 2 unread items, got %d", len(payload.Items))
		}
		for _, item := range payload.Items {
			if item.ID == formatGReaderItemID(readItem.ID) {
				t.Fatalf("read item should have been excluded")
			}
		}
	})

	t.Run("xt=starred excludes starred items", func(t *testing.T) {
		payload := fetch(t, "/reader/api/0/stream/contents/"+greaderReadingList+"?xt="+url.QueryEscape(greaderStarredState))
		if len(payload.Items) != 2 {
			t.Fatalf("expected 2 unstarred items, got %d", len(payload.Items))
		}
		for _, item := range payload.Items {
			if item.ID == formatGReaderItemID(starredItem.ID) {
				t.Fatal("starred item should have been excluded")
			}
		}
	})

	t.Run("contradictory it and xt filters return nothing", func(t *testing.T) {
		target := "/reader/api/0/stream/contents/user/-/state/com.google/read?xt=" + url.QueryEscape(greaderReadState)
		payload := fetch(t, target)
		if len(payload.Items) != 0 {
			t.Fatalf("expected an empty page, got %d items", len(payload.Items))
		}
	})

	t.Run("feed stream", func(t *testing.T) {
		payload := fetch(t, "/reader/api/0/stream/contents/"+greaderFeedStreamID(feed.ID))
		if len(payload.Items) != 3 {
			t.Fatalf("expected 3 items, got %d", len(payload.Items))
		}
	})

	t.Run("POST works for read endpoints", func(t *testing.T) {
		w := performGReaderRequest(
			router, http.MethodPost, "/reader/api/0/stream/contents/"+greaderReadingList, token,
			url.Values{"n": {"1"}},
		)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		var payload greaderItemsPayload
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if len(payload.Items) != 1 {
			t.Fatalf("expected one item, got %d", len(payload.Items))
		}
	})

	t.Run("label stream with encoded slashes", func(t *testing.T) {
		payload := fetch(t, "/reader/api/0/stream/contents/user%2F-%2Flabel%2FTech")
		if len(payload.Items) != 3 {
			t.Fatalf("expected 3 items, got %d", len(payload.Items))
		}
	})

	t.Run("label stream with plain slashes", func(t *testing.T) {
		payload := fetch(t, "/reader/api/0/stream/contents/user/-/label/Tech")
		if len(payload.Items) != 3 {
			t.Fatalf("expected 3 items, got %d", len(payload.Items))
		}
	})

	t.Run("ot filters on crawl time, not pub_date", func(t *testing.T) {
		future := mustCreateGReaderItem(t, st, feed.ID, "Future", 4102444800)
		cutoff := future.CreatedAt + 3600*1000 // far enough behind the future pub_date
		payload := fetch(t, "/reader/api/0/stream/contents/"+greaderReadingList+"?ot="+strconv.FormatInt(cutoff, 10))
		found := false
		for _, item := range payload.Items {
			if item.ID == formatGReaderItemID(future.ID) {
				found = true
			}
		}
		if !found {
			t.Fatal("expected an item with a future pub_date to pass an ot filter based on crawl time")
		}
	})

	t.Run("unknown streams", func(t *testing.T) {
		w := performGReaderRequest(router, http.MethodGet, "/reader/api/0/stream/contents/feed/9999", token, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for unknown feed, got %d", w.Code)
		}
		w = performGReaderRequest(router, http.MethodGet, "/reader/api/0/stream/contents/user/-/label/Missing", token, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for unknown label, got %d", w.Code)
		}
		w = performGReaderRequest(router, http.MethodGet, "/reader/api/0/stream/contents/not-a-stream", token, nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for a malformed stream, got %d", w.Code)
		}
	})
}

func TestGReaderStreamItemsContents(t *testing.T) {
	h, st := newGReaderTestHandler(t)
	router := h.SetupRouter()
	token := greaderClientLogin(t, router, "fusion", "secret")

	group := mustCreateGReaderGroup(t, st, "Tech")
	feed := mustCreateGReaderFeed(t, st, group.ID, "Feed", "https://example.com/rss.xml", "https://example.com")
	first := mustCreateGReaderItem(t, st, feed.ID, "One", 100)
	second := mustCreateGReaderItem(t, st, feed.ID, "Two", 200)

	decode := func(t *testing.T, w *httptest.ResponseRecorder) greaderItemsPayload {
		t.Helper()
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
		var payload greaderItemsPayload
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		return payload
	}

	t.Run("both id forms, request order, unknown ids skipped", func(t *testing.T) {
		payload := decode(t, performGReaderRequest(
			router, http.MethodPost, "/reader/api/0/stream/items/contents", token,
			url.Values{"i": {formatGReaderItemID(second.ID), "9999", strconv.FormatInt(first.ID, 10)}},
		))
		if len(payload.Items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(payload.Items))
		}
		if payload.Items[0].ID != formatGReaderItemID(second.ID) || payload.Items[1].ID != formatGReaderItemID(first.ID) {
			t.Fatalf("expected request order, got %v", payload.Items)
		}
	})

	t.Run("invalid item id", func(t *testing.T) {
		w := performGReaderRequest(
			router, http.MethodPost, "/reader/api/0/stream/items/contents", token,
			url.Values{"i": {"tag:google.com,2005:reader/item/zzzz"}},
		)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("missing item ids", func(t *testing.T) {
		for _, values := range []url.Values{{}, {"i": {""}}} {
			w := performGReaderRequest(router, http.MethodPost, "/reader/api/0/stream/items/contents", token, values)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for %v, got %d", values, w.Code)
			}
		}
	})
}

func TestGReaderEditTag(t *testing.T) {
	h, st := newGReaderTestHandler(t)
	router := h.SetupRouter()
	token := greaderClientLogin(t, router, "fusion", "secret")

	group := mustCreateGReaderGroup(t, st, "Tech")
	feed := mustCreateGReaderFeed(t, st, group.ID, "Feed", "https://example.com/rss.xml", "https://example.com")
	first := mustCreateGReaderItem(t, st, feed.ID, "One", 100)
	second := mustCreateGReaderItem(t, st, feed.ID, "Two", 200)

	edit := func(t *testing.T, values url.Values) *httptest.ResponseRecorder {
		t.Helper()
		w := performGReaderRequest(router, http.MethodPost, "/reader/api/0/edit-tag", token, values)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
		if got := w.Body.String(); got != "OK" {
			t.Fatalf("expected OK body, got %q", got)
		}
		return w
	}

	isUnread := func(t *testing.T, id int64) bool {
		t.Helper()
		item, err := st.GetItem(id)
		if err != nil {
			t.Fatalf("get item: %v", err)
		}
		return item.Unread
	}

	t.Run("add read uses the long id form", func(t *testing.T) {
		edit(t, url.Values{
			"i": {formatGReaderItemID(first.ID)},
			"a": {greaderReadState},
		})
		if isUnread(t, first.ID) {
			t.Fatal("expected the item to be read")
		}
	})

	t.Run("remove read marks unread", func(t *testing.T) {
		edit(t, url.Values{
			"i": {formatGReaderItemID(first.ID)},
			"r": {greaderReadState},
		})
		if !isUnread(t, first.ID) {
			t.Fatal("expected the item to be unread")
		}
	})

	t.Run("add kept-unread marks unread and accepts other user ids", func(t *testing.T) {
		edit(t, url.Values{
			"i": {strconv.FormatInt(first.ID, 10)},
			"a": {"user/42/state/com.google/read"},
		})
		if isUnread(t, first.ID) {
			t.Fatal("expected the item to be read")
		}
		edit(t, url.Values{
			"i": {strconv.FormatInt(first.ID, 10)},
			"a": {"user/42/state/com.google/kept-unread"},
		})
		if !isUnread(t, first.ID) {
			t.Fatal("expected the item to be unread")
		}
	})

	t.Run("star and unstar map to bookmarks", func(t *testing.T) {
		edit(t, url.Values{
			"i": {strconv.FormatInt(second.ID, 10)},
			"a": {greaderStarredState},
		})
		exists, err := st.BookmarkExists(second.Link)
		if err != nil || !exists {
			t.Fatalf("expected a bookmark after starring, exists=%v err=%v", exists, err)
		}

		edit(t, url.Values{
			"i": {strconv.FormatInt(second.ID, 10)},
			"r": {greaderStarredState},
		})
		exists, err = st.BookmarkExists(second.Link)
		if err != nil || exists {
			t.Fatalf("expected the bookmark to be removed, exists=%v err=%v", exists, err)
		}
	})

	t.Run("unknown tags are ignored", func(t *testing.T) {
		edit(t, url.Values{
			"i": {strconv.FormatInt(second.ID, 10)},
			"a": {"user/-/state/com.google/broadcast"},
		})
		if !isUnread(t, second.ID) {
			t.Fatal("expected the ignored tag to leave read state alone")
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		w := performGReaderRequest(
			router, http.MethodPost, "/reader/api/0/edit-tag", token,
			url.Values{"i": {"nope"}, "a": {greaderReadState}},
		)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("missing ids", func(t *testing.T) {
		w := performGReaderRequest(
			router, http.MethodPost, "/reader/api/0/edit-tag", token,
			url.Values{"a": {greaderReadState}},
		)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})
}

func TestGReaderMarkAllAsRead(t *testing.T) {
	h, st := newGReaderTestHandler(t)
	router := h.SetupRouter()
	token := greaderClientLogin(t, router, "fusion", "secret")

	tech := mustCreateGReaderGroup(t, st, "Tech")
	news := mustCreateGReaderGroup(t, st, "News")
	feed1 := mustCreateGReaderFeed(t, st, tech.ID, "Feed 1", "https://example.com/1.xml", "https://example.com")
	feed2 := mustCreateGReaderFeed(t, st, news.ID, "Feed 2", "https://example.com/2.xml", "https://example.com")

	futurePubDate := mustCreateGReaderItem(t, st, feed1.ID, "Future", 4102444800)
	otherItem := mustCreateGReaderItem(t, st, feed2.ID, "Other", 100)

	isUnread := func(t *testing.T, id int64) bool {
		t.Helper()
		item, err := st.GetItem(id)
		if err != nil {
			t.Fatalf("get item: %v", err)
		}
		return item.Unread
	}

	markAll := func(t *testing.T, values url.Values) *httptest.ResponseRecorder {
		t.Helper()
		w := performGReaderRequest(router, http.MethodPost, "/reader/api/0/mark-all-as-read", token, values)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
		return w
	}

	// ts is in microseconds and far enough in the future to cover "now".
	ts := strconv.FormatInt((time.Now().Unix()+86400)*1_000_000, 10)

	t.Run("feed scope uses crawl time", func(t *testing.T) {
		markAll(t, url.Values{"s": {greaderFeedStreamID(feed1.ID)}, "ts": {ts}})
		if isUnread(t, futurePubDate.ID) {
			t.Fatal("expected an item with a future pub_date to be marked read")
		}
		if !isUnread(t, otherItem.ID) {
			t.Fatal("expected other feeds to stay unread")
		}
	})

	t.Run("label scope", func(t *testing.T) {
		markAll(t, url.Values{"s": {greaderLabelState("News")}})
		if isUnread(t, otherItem.ID) {
			t.Fatal("expected the label scope to mark its feed read")
		}
	})

	t.Run("reading list with a past ts marks nothing", func(t *testing.T) {
		// Re-create an unread item so the assertion is meaningful.
		fresh := mustCreateGReaderItem(t, st, feed1.ID, "Fresh", 500)
		past := strconv.FormatInt(1_000_000, 10) // 1 second after the epoch
		markAll(t, url.Values{"s": {greaderReadingList}, "ts": {past}})
		if !isUnread(t, fresh.ID) {
			t.Fatal("expected items crawled after ts to stay unread")
		}
	})

	t.Run("reading list without ts", func(t *testing.T) {
		markAll(t, url.Values{"s": {greaderReadingList}})
		items, err := st.ListItems(store.ListItemsParams{})
		if err != nil {
			t.Fatalf("list items: %v", err)
		}
		for _, item := range items {
			if item.Unread {
				t.Fatalf("expected item %d to be read", item.ID)
			}
		}
	})

	t.Run("invalid input", func(t *testing.T) {
		for _, values := range []url.Values{
			{"s": {greaderReadingList}, "ts": {"abc"}},
			{"s": {"user/-/state/com.google/starred"}},
			{"s": {"feed/9999"}},
			{"s": {"user/-/label/Missing"}},
			{},
		} {
			w := performGReaderRequest(router, http.MethodPost, "/reader/api/0/mark-all-as-read", token, values)
			if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
				t.Fatalf("expected 400 or 404 for %v, got %d", values, w.Code)
			}
		}
	})
}

// newFeedServer serves one RSS document so subscription and quickadd tests
// stay offline. The feed is only served at /feed.xml, like a real feed URL.
func newFeedServer(t *testing.T, title string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/feed.xml" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel><title>%s</title><link>https://example.com</link>
<item><guid>g1</guid><title>Item</title><link>https://example.com/1</link></item>
</channel></rss>`, title)
	}))
	t.Cleanup(server.Close)

	return server
}

func newGReaderPrivateHandler(t *testing.T) (*Handler, *store.Store) {
	t.Helper()

	return newGReaderTestHandlerWithConfig(t, &config.Config{AllowPrivateFeeds: true})
}

func TestGReaderSubscriptionEdit(t *testing.T) {
	h, st := newGReaderPrivateHandler(t)
	router := h.SetupRouter()
	token := greaderClientLogin(t, router, "fusion", "secret")

	tech := mustCreateGReaderGroup(t, st, "Tech")
	server := newFeedServer(t, "Discovered Feed")
	feedURL := server.URL + "/feed.xml"

	edit := func(t *testing.T, values url.Values) *httptest.ResponseRecorder {
		t.Helper()
		w := performGReaderRequest(router, http.MethodPost, "/reader/api/0/subscription/edit", token, values)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
		return w
	}

	getFeed := func(t *testing.T, link string) *model.Feed {
		t.Helper()
		feed, err := st.GetFeedByLink(link)
		if err != nil {
			t.Fatalf("get feed by link: %v", err)
		}
		return feed
	}

	t.Run("subscribe with a new label", func(t *testing.T) {
		edit(t, url.Values{
			"ac": {"subscribe"},
			"s":  {"feed/" + feedURL},
			"t":  {"My Feed"},
			"a":  {greaderLabelState("Tech")},
		})

		feed := getFeed(t, feedURL)
		if feed.Name != "My Feed" {
			t.Fatalf("expected title to be used, got %q", feed.Name)
		}
		if feed.GroupID != tech.ID {
			t.Fatalf("expected the feed in Tech, got group %d", feed.GroupID)
		}

		// Subscribing again is idempotent and applies a label move.
		edit(t, url.Values{
			"ac": {"subscribe"},
			"s":  {"feed/" + feedURL},
			"a":  {greaderLabelState("Reading")},
		})
		feed = getFeed(t, feedURL)
		reading, err := st.ListGroups()
		if err != nil {
			t.Fatalf("list groups: %v", err)
		}
		var readingID int64
		for _, group := range reading {
			if group.Name == "Reading" {
				readingID = group.ID
			}
		}
		if readingID == 0 {
			t.Fatal("expected the label to be created")
		}
		if feed.GroupID != readingID {
			t.Fatalf("expected the feed to move to Reading, got group %d", feed.GroupID)
		}
	})

	t.Run("subscribe without a label uses Default", func(t *testing.T) {
		other := newFeedServer(t, "Default Feed")
		edit(t, url.Values{
			"ac": {"subscribe"},
			"s":  {"feed/" + other.URL + "/feed.xml"},
		})

		feed := getFeed(t, other.URL+"/feed.xml")
		if feed.GroupID != store.DefaultGroupID {
			t.Fatalf("expected the default group, got %d", feed.GroupID)
		}
		if feed.Name != other.URL+"/feed.xml" {
			t.Fatalf("expected the link as fallback title, got %q", feed.Name)
		}
	})

	t.Run("subscribe to an existing feed id renames it", func(t *testing.T) {
		feed := getFeed(t, feedURL)
		edit(t, url.Values{
			"ac": {"subscribe"},
			"s":  {greaderFeedStreamID(feed.ID)},
			"t":  {"Renamed"},
		})
		if got := getFeed(t, feedURL); got.Name != "Renamed" {
			t.Fatalf("expected renamed feed, got %q", got.Name)
		}
	})

	t.Run("edit renames and moves", func(t *testing.T) {
		feed := getFeed(t, feedURL)
		edit(t, url.Values{
			"ac": {"edit"},
			"s":  {greaderFeedStreamID(feed.ID)},
			"t":  {"Edited"},
			"a":  {greaderLabelState("News")},
		})
		feed = getFeed(t, feedURL)
		if feed.Name != "Edited" {
			t.Fatalf("expected edited name, got %q", feed.Name)
		}
		group, err := st.GetGroup(feed.GroupID)
		if err != nil {
			t.Fatalf("get group: %v", err)
		}
		if group.Name != "News" {
			t.Fatalf("expected the News group, got %q", group.Name)
		}

		// r on its own moves the feed back to Default.
		edit(t, url.Values{
			"ac": {"edit"},
			"s":  {greaderFeedStreamID(feed.ID)},
			"r":  {greaderLabelState("News")},
		})
		if got := getFeed(t, feedURL); got.GroupID != store.DefaultGroupID {
			t.Fatalf("expected the default group, got %d", got.GroupID)
		}
	})

	t.Run("unsubscribe by id and by url", func(t *testing.T) {
		feed := getFeed(t, feedURL)
		edit(t, url.Values{"ac": {"unsubscribe"}, "s": {greaderFeedStreamID(feed.ID)}})
		if _, err := st.GetFeed(feed.ID); !errors.Is(err, store.ErrNotFound) {
			t.Fatalf("expected the feed to be deleted, got %v", err)
		}

		other := newFeedServer(t, "Remove Me")
		link := other.URL + "/feed.xml"
		edit(t, url.Values{"ac": {"subscribe"}, "s": {"feed/" + link}})
		edit(t, url.Values{"ac": {"unsubscribe"}, "s": {"feed/" + link}})
		if _, err := st.GetFeedByLink(link); !errors.Is(err, store.ErrNotFound) {
			t.Fatalf("expected the feed to be deleted, got %v", err)
		}

		// Unsubscribing an unknown feed is a no-op.
		edit(t, url.Values{"ac": {"unsubscribe"}, "s": {"feed/9999"}})
	})

	t.Run("empty r on edit does not move the feed", func(t *testing.T) {
		moveServer := newFeedServer(t, "Move Test")
		moveLink := moveServer.URL + "/feed.xml"
		edit(t, url.Values{
			"ac": {"subscribe"},
			"s":  {"feed/" + moveLink},
			"a":  {greaderLabelState("Tech")},
		})
		moved := getFeed(t, moveLink)

		edit(t, url.Values{
			"ac": {"edit"},
			"s":  {greaderFeedStreamID(moved.ID)},
			"t":  {"Rename Only"},
			"a":  {""},
			"r":  {""},
		})
		after := getFeed(t, moveLink)
		if after.Name != "Rename Only" {
			t.Fatalf("expected the rename to apply, got %q", after.Name)
		}
		if after.GroupID != moved.GroupID {
			t.Fatalf("expected the feed to stay in group %d, got %d", moved.GroupID, after.GroupID)
		}
	})

	t.Run("failed edits do not create labels", func(t *testing.T) {
		orphanServer := newFeedServer(t, "Orphan Test")
		orphanLink := orphanServer.URL + "/feed.xml"
		edit(t, url.Values{"ac": {"subscribe"}, "s": {"feed/" + orphanLink}})
		orphan := getFeed(t, orphanLink)

		groupsBefore, err := st.ListGroups()
		if err != nil {
			t.Fatalf("list groups: %v", err)
		}

		for _, values := range []url.Values{
			// Rejected before the label is resolved.
			{"ac": {"subscribe"}, "s": {"feed/not a url"}, "a": {greaderLabelState("Ghost")}},
			{"ac": {"edit"}, "s": {"feed/9999"}, "a": {greaderLabelState("Ghost")}},
			// Malformed label on a valid feed.
			{"ac": {"edit"}, "s": {greaderFeedStreamID(orphan.ID)}, "a": {"not-a-label"}},
		} {
			w := performGReaderRequest(router, http.MethodPost, "/reader/api/0/subscription/edit", token, values)
			if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
				t.Fatalf("expected 400 or 404 for %v, got %d", values, w.Code)
			}
		}

		groupsAfter, err := st.ListGroups()
		if err != nil {
			t.Fatalf("list groups: %v", err)
		}
		if len(groupsAfter) != len(groupsBefore) {
			t.Fatalf("expected %d groups, got %d", len(groupsBefore), len(groupsAfter))
		}
	})

	t.Run("invalid input", func(t *testing.T) {
		for _, values := range []url.Values{
			{"ac": {"nope"}, "s": {greaderFeedStreamID(1)}},
			{"ac": {"subscribe"}},
			{"ac": {"subscribe"}, "s": {"not-a-feed"}},
			{"ac": {"subscribe"}, "s": {"feed/not a url"}},
			{"ac": {"edit"}, "s": {"feed/9999"}},
		} {
			w := performGReaderRequest(router, http.MethodPost, "/reader/api/0/subscription/edit", token, values)
			if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
				t.Fatalf("expected 400 or 404 for %v, got %d", values, w.Code)
			}
		}
	})
}

func TestGReaderQuickadd(t *testing.T) {
	h, st := newGReaderPrivateHandler(t)
	router := h.SetupRouter()
	token := greaderClientLogin(t, router, "fusion", "secret")

	server := newFeedServer(t, "Discovered Feed")
	feedURL := server.URL + "/feed.xml"

	decode := func(t *testing.T, w *httptest.ResponseRecorder) greaderQuickaddResponse {
		t.Helper()
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
		var payload greaderQuickaddResponse
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		return payload
	}

	t.Run("subscribes to the first discovered feed in Default", func(t *testing.T) {
		payload := decode(t, performGReaderRequest(
			router, http.MethodPost, "/reader/api/0/subscription/quickadd", token,
			url.Values{"quickadd": {feedURL}},
		))
		if payload.NumResults != 1 || payload.Query != feedURL {
			t.Fatalf("unexpected quickadd response %+v", payload)
		}

		feed, err := st.GetFeedByLink(feedURL)
		if err != nil {
			t.Fatalf("expected the feed to be created: %v", err)
		}
		if payload.StreamID != greaderFeedStreamID(feed.ID) || payload.StreamName != feed.Name {
			t.Fatalf("unexpected stream metadata %+v", payload)
		}
		if feed.GroupID != store.DefaultGroupID {
			t.Fatalf("expected the default group, got %d", feed.GroupID)
		}
	})

	t.Run("repeating quickadd returns the existing subscription", func(t *testing.T) {
		payload := decode(t, performGReaderRequest(
			router, http.MethodPost, "/reader/api/0/subscription/quickadd", token,
			url.Values{"quickadd": {feedURL}},
		))
		if payload.NumResults != 1 {
			t.Fatalf("expected the existing feed, got %+v", payload)
		}
	})

	t.Run("nothing found", func(t *testing.T) {
		empty := httptest.NewServer(http.NotFoundHandler())
		defer empty.Close()

		payload := decode(t, performGReaderRequest(
			router, http.MethodPost, "/reader/api/0/subscription/quickadd", token,
			url.Values{"quickadd": {empty.URL}},
		))
		if payload.NumResults != 0 || payload.StreamID != "" {
			t.Fatalf("unexpected quickadd response %+v", payload)
		}
		if _, err := st.GetFeedByLink(empty.URL); !errors.Is(err, store.ErrNotFound) {
			t.Fatalf("expected no feed to be created, got %v", err)
		}
	})

	t.Run("missing url", func(t *testing.T) {
		w := performGReaderRequest(router, http.MethodPost, "/reader/api/0/subscription/quickadd", token, url.Values{})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})
}

func TestGReaderRenameTag(t *testing.T) {
	h, st := newGReaderTestHandler(t)
	router := h.SetupRouter()
	token := greaderClientLogin(t, router, "fusion", "secret")

	tech := mustCreateGReaderGroup(t, st, "Tech")
	mustCreateGReaderGroup(t, st, "Archive")
	feed := mustCreateGReaderFeed(t, st, tech.ID, "Feed", "https://example.com/rss.xml", "https://example.com")

	rename := func(t *testing.T, values url.Values) *httptest.ResponseRecorder {
		t.Helper()
		return performGReaderRequest(router, http.MethodPost, "/reader/api/0/rename-tag", token, values)
	}

	if w := rename(t, url.Values{
		"s":    {greaderLabelState("Tech")},
		"dest": {greaderLabelState("Renamed")},
	}); w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	updated, err := st.GetGroup(tech.ID)
	if err != nil {
		t.Fatalf("get group: %v", err)
	}
	if updated.Name != "Renamed" {
		t.Fatalf("expected renamed group, got %q", updated.Name)
	}
	if got, err := st.GetFeed(feed.ID); err != nil || got.GroupID != tech.ID {
		t.Fatalf("expected the feed to stay in the renamed group, got %+v err=%v", got, err)
	}

	if w := rename(t, url.Values{
		"s":    {greaderLabelState("Renamed")},
		"dest": {greaderLabelState("Archive")},
	}); w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when the target label exists, got %d", w.Code)
	}
	if w := rename(t, url.Values{
		"s":    {greaderLabelState("Missing")},
		"dest": {greaderLabelState("Other")},
	}); w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for a missing source label, got %d", w.Code)
	}
	if w := rename(t, url.Values{
		"s":    {"not-a-label"},
		"dest": {greaderLabelState("Other")},
	}); w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for an invalid tag, got %d", w.Code)
	}
}

func TestGReaderDisableTag(t *testing.T) {
	h, st := newGReaderTestHandler(t)
	router := h.SetupRouter()
	token := greaderClientLogin(t, router, "fusion", "secret")

	tech := mustCreateGReaderGroup(t, st, "Tech")
	feed := mustCreateGReaderFeed(t, st, tech.ID, "Feed", "https://example.com/rss.xml", "https://example.com")

	disable := func(t *testing.T, values url.Values) *httptest.ResponseRecorder {
		t.Helper()
		return performGReaderRequest(router, http.MethodPost, "/reader/api/0/disable-tag", token, values)
	}

	if w := disable(t, url.Values{"s": {greaderLabelState("Default")}}); w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when deleting the Default label, got %d", w.Code)
	}

	if w := disable(t, url.Values{"s": {greaderLabelState("Tech")}}); w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	if _, err := st.GetGroup(tech.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected the group to be deleted, got %v", err)
	}
	moved, err := st.GetFeed(feed.ID)
	if err != nil {
		t.Fatalf("get feed: %v", err)
	}
	if moved.GroupID != store.DefaultGroupID {
		t.Fatalf("expected the feed to move to Default, got %d", moved.GroupID)
	}

	if w := disable(t, url.Values{"s": {greaderLabelState("Missing")}}); w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for a missing label, got %d", w.Code)
	}
	if w := disable(t, url.Values{"s": {"not-a-label"}}); w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for an invalid tag, got %d", w.Code)
	}
}

// TestGReaderEndToEndClientFlow walks the request sequence a GReader client
// performs: login, discovery of streams, sync, state changes, subscription
// management and label management.
func TestGReaderEndToEndClientFlow(t *testing.T) {
	h, st := newGReaderPrivateHandler(t)
	router := h.SetupRouter()

	server := newFeedServer(t, "Flow Feed")
	feedURL := server.URL + "/feed.xml"

	do := func(t *testing.T, method, target, token string, values url.Values) *httptest.ResponseRecorder {
		t.Helper()
		w := performGReaderRequest(router, method, target, token, values)
		if w.Code != http.StatusOK {
			t.Fatalf("%s %s: expected 200, got %d: %s", method, target, w.Code, w.Body.String())
		}
		return w
	}

	// 1. ClientLogin -> user-info -> token
	token := greaderClientLogin(t, router, "fusion", "secret")

	var userInfo greaderUserInfoResponse
	if err := json.Unmarshal(do(t, http.MethodGet, "/reader/api/0/user-info", token, nil).Body.Bytes(), &userInfo); err != nil {
		t.Fatalf("unmarshal user-info: %v", err)
	}
	if userInfo.UserID != "1" || userInfo.UserName != "fusion" {
		t.Fatalf("unexpected user-info %+v", userInfo)
	}
	if got := do(t, http.MethodGet, "/reader/api/0/token", token, nil).Body.String(); got != token {
		t.Fatalf("expected token %q, got %q", token, got)
	}

	// 2. Empty account: lists and counts still answer.
	do(t, http.MethodGet, "/reader/api/0/subscription/list", token, nil)
	do(t, http.MethodGet, "/reader/api/0/tag/list", token, nil)
	do(t, http.MethodGet, "/reader/api/0/unread-count", token, nil)

	// 3. Subscribe into a new label.
	do(t, http.MethodPost, "/reader/api/0/subscription/edit", token, url.Values{
		"ac": {"subscribe"},
		"s":  {"feed/" + feedURL},
		"t":  {"Flow Feed"},
		"a":  {greaderLabelState("Flow")},
	})
	feed, err := st.GetFeedByLink(feedURL)
	if err != nil {
		t.Fatalf("expected the subscription to create a feed: %v", err)
	}

	// Seed items directly: the flow starts once a feed has been pulled.
	mustCreateGReaderItem(t, st, feed.ID, "Alpha", 100)
	beta := mustCreateGReaderItem(t, st, feed.ID, "Beta", 200)
	gamma := mustCreateGReaderItem(t, st, feed.ID, "Gamma", 300)

	var subscriptions struct {
		Subscriptions []greaderSubscription `json:"subscriptions"`
	}
	if err := json.Unmarshal(do(t, http.MethodGet, "/reader/api/0/subscription/list", token, nil).Body.Bytes(), &subscriptions); err != nil {
		t.Fatalf("unmarshal subscriptions: %v", err)
	}
	if len(subscriptions.Subscriptions) != 1 || subscriptions.Subscriptions[0].ID != greaderFeedStreamID(feed.ID) {
		t.Fatalf("unexpected subscriptions %+v", subscriptions.Subscriptions)
	}

	var tags struct {
		Tags []greaderTag `json:"tags"`
	}
	if err := json.Unmarshal(do(t, http.MethodGet, "/reader/api/0/tag/list", token, nil).Body.Bytes(), &tags); err != nil {
		t.Fatalf("unmarshal tags: %v", err)
	}
	if !slices.ContainsFunc(tags.Tags, func(tag greaderTag) bool { return tag.ID == greaderLabelState("Flow") }) {
		t.Fatalf("expected the Flow label in %+v", tags.Tags)
	}

	// 4. Sync unread items: first page + continuation.
	var page greaderItemRefsPayload
	if err := json.Unmarshal(do(t, http.MethodGet,
		"/reader/api/0/stream/items/ids?s="+url.QueryEscape(greaderReadingList)+"&xt="+url.QueryEscape(greaderReadState)+"&n=2",
		token, nil).Body.Bytes(), &page); err != nil {
		t.Fatalf("unmarshal item refs: %v", err)
	}
	if len(page.ItemRefs) != 2 || page.Continuation == "" {
		t.Fatalf("expected a first page with continuation, got %+v", page)
	}
	var rest greaderItemRefsPayload
	if err := json.Unmarshal(do(t, http.MethodGet,
		"/reader/api/0/stream/items/ids?s="+url.QueryEscape(greaderReadingList)+"&c="+page.Continuation,
		token, nil).Body.Bytes(), &rest); err != nil {
		t.Fatalf("unmarshal item refs: %v", err)
	}
	if len(rest.ItemRefs) != 1 || rest.Continuation != "" {
		t.Fatalf("expected the last page, got %+v", rest)
	}

	// 5. Full contents by stream and by ids, both id forms.
	var contents greaderItemsPayload
	if err := json.Unmarshal(do(t, http.MethodGet,
		"/reader/api/0/stream/contents/"+greaderReadingList, token, nil).Body.Bytes(), &contents); err != nil {
		t.Fatalf("unmarshal contents: %v", err)
	}
	if len(contents.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(contents.Items))
	}
	var byIDs greaderItemsPayload
	if err := json.Unmarshal(do(t, http.MethodPost, "/reader/api/0/stream/items/contents", token, url.Values{
		"i": {formatGReaderItemID(beta.ID), strconv.FormatInt(gamma.ID, 10)},
	}).Body.Bytes(), &byIDs); err != nil {
		t.Fatalf("unmarshal items by ids: %v", err)
	}
	if len(byIDs.Items) != 2 || byIDs.Items[0].ID != formatGReaderItemID(beta.ID) {
		t.Fatalf("unexpected items by ids %+v", byIDs.Items)
	}

	// 6. Read/star state.
	do(t, http.MethodPost, "/reader/api/0/edit-tag", token, url.Values{
		"i": {formatGReaderItemID(beta.ID)},
		"a": {greaderReadState},
	})
	do(t, http.MethodPost, "/reader/api/0/edit-tag", token, url.Values{
		"i": {strconv.FormatInt(gamma.ID, 10)},
		"a": {greaderStarredState},
	})
	if item, err := st.GetItem(beta.ID); err != nil || item.Unread {
		t.Fatalf("expected beta to be read, unread=%v err=%v", item.Unread, err)
	}
	starred, err := st.ListSavedItemIDs()
	if err != nil {
		t.Fatalf("list saved item ids: %v", err)
	}
	if !slices.Equal(starred, []int64{gamma.ID}) {
		t.Fatalf("expected gamma to be starred, got %v", starred)
	}

	var unread struct {
		Max          int64                `json:"max"`
		UnreadCounts []greaderUnreadCount `json:"unreadcounts"`
	}
	if err := json.Unmarshal(do(t, http.MethodGet, "/reader/api/0/unread-count", token, nil).Body.Bytes(), &unread); err != nil {
		t.Fatalf("unmarshal unread count: %v", err)
	}
	if unread.Max != 2 {
		t.Fatalf("expected 2 unread items, got %d", unread.Max)
	}

	// Undo the state changes.
	do(t, http.MethodPost, "/reader/api/0/edit-tag", token, url.Values{
		"i": {formatGReaderItemID(beta.ID)},
		"r": {greaderReadState},
	})
	do(t, http.MethodPost, "/reader/api/0/edit-tag", token, url.Values{
		"i": {strconv.FormatInt(gamma.ID, 10)},
		"r": {greaderStarredState},
	})

	// 7. Mark all as read: feed, label, then the whole reading list.
	do(t, http.MethodPost, "/reader/api/0/mark-all-as-read", token, url.Values{
		"s": {greaderFeedStreamID(feed.ID)},
	})
	if items, err := st.ListItems(store.ListItemsParams{FeedID: &feed.ID}); err != nil || len(items) == 0 {
		t.Fatalf("list feed items: %v", err)
	} else {
		for _, item := range items {
			if item.Unread {
				t.Fatalf("expected the feed scope to mark item %d read", item.ID)
			}
		}
	}
	do(t, http.MethodPost, "/reader/api/0/mark-all-as-read", token, url.Values{
		"s": {greaderLabelState("Flow")},
	})
	do(t, http.MethodPost, "/reader/api/0/mark-all-as-read", token, url.Values{
		"s": {greaderReadingList},
	})

	// 8. Rename and move the subscription; unsubscribe it.
	do(t, http.MethodPost, "/reader/api/0/subscription/edit", token, url.Values{
		"ac": {"edit"},
		"s":  {greaderFeedStreamID(feed.ID)},
		"t":  {"Flow Feed Renamed"},
		"a":  {greaderLabelState("Flow")},
	})
	renamed, err := st.GetFeed(feed.ID)
	if err != nil || renamed.Name != "Flow Feed Renamed" {
		t.Fatalf("expected renamed feed, got %+v err=%v", renamed, err)
	}

	extra := newFeedServer(t, "Extra Feed")
	do(t, http.MethodPost, "/reader/api/0/subscription/edit", token, url.Values{
		"ac": {"subscribe"},
		"s":  {"feed/" + extra.URL + "/feed.xml"},
	})
	do(t, http.MethodPost, "/reader/api/0/subscription/edit", token, url.Values{
		"ac": {"unsubscribe"},
		"s":  {"feed/" + extra.URL + "/feed.xml"},
	})

	// 9. Quickadd, rename-tag and disable-tag.
	quickaddServer := newFeedServer(t, "Quickadd Feed")
	var quickadd greaderQuickaddResponse
	if err := json.Unmarshal(do(t, http.MethodPost, "/reader/api/0/subscription/quickadd", token, url.Values{
		"quickadd": {quickaddServer.URL + "/feed.xml"},
	}).Body.Bytes(), &quickadd); err != nil {
		t.Fatalf("unmarshal quickadd: %v", err)
	}
	if quickadd.NumResults != 1 || quickadd.StreamID == "" {
		t.Fatalf("unexpected quickadd %+v", quickadd)
	}

	do(t, http.MethodPost, "/reader/api/0/rename-tag", token, url.Values{
		"s":    {greaderLabelState("Flow")},
		"dest": {greaderLabelState("Flow Renamed")},
	})
	do(t, http.MethodPost, "/reader/api/0/disable-tag", token, url.Values{
		"s": {greaderLabelState("Flow Renamed")},
	})

	final, err := st.GetFeed(feed.ID)
	if err != nil {
		t.Fatalf("get feed: %v", err)
	}
	if final.GroupID != store.DefaultGroupID {
		t.Fatalf("expected the feed to fall back to Default, got %d", final.GroupID)
	}
}

func TestSetupRouterDoesNotServeFrontendForReaderPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{}
	router := h.SetupRouter()

	for _, target := range []string{
		"/reader/not-registered",
		"/accounts/not-registered",
	} {
		w := performRequest(router, http.MethodGet, target, nil, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("%s: expected 404, got %d", target, w.Code)
		}
		if contentType := w.Header().Get("Content-Type"); strings.Contains(contentType, "text/html") {
			t.Fatalf("%s: expected no SPA fallback, got %q", target, contentType)
		}
	}
}

// Quickadd validates the link discovered on the page, not just the page URL:
// an HTML page reachable at a public address must not be able to subscribe
// Fusion to an internal feed.
func TestGReaderQuickaddRejectsInternalDiscoveredFeed(t *testing.T) {
	h, st := newGReaderTestHandler(t) // AllowPrivateFeeds is false
	router := h.SetupRouter()
	token := greaderClientLogin(t, router, "fusion", "secret")

	internal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = fmt.Fprint(w, `<?xml version="1.0"?><rss version="2.0"><channel><title>Internal</title></channel></rss>`)
	}))
	defer internal.Close()

	page := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = fmt.Fprintf(w, `<!doctype html><html><head><link rel="alternate" type="application/rss+xml" href="%s/feed.xml"></head><body><a href="/about">About</a></body></html>`, internal.URL)
	}))
	defer page.Close()

	publicURL := routeFakePublicServer(t, page)

	w := performGReaderRequest(
		router, http.MethodPost, "/reader/api/0/subscription/quickadd", token,
		url.Values{"quickadd": {publicURL}},
	)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected the internal discovered feed to be rejected, got %d: %s", w.Code, w.Body.String())
	}

	feeds, err := st.ListFeeds()
	if err != nil {
		t.Fatalf("list feeds: %v", err)
	}
	if len(feeds) != 0 {
		t.Fatalf("expected no feed to be created, got %+v", feeds)
	}
}

func TestGReaderKeptUnreadStreamIsUnread(t *testing.T) {
	h, st := newGReaderTestHandler(t)
	router := h.SetupRouter()
	token := greaderClientLogin(t, router, "fusion", "secret")

	group := mustCreateGReaderGroup(t, st, "Tech")
	feed := mustCreateGReaderFeed(t, st, group.ID, "Feed", "https://example.com/rss.xml", "https://example.com")
	unread := mustCreateGReaderItem(t, st, feed.ID, "Unread", 100)
	read := mustCreateGReaderItem(t, st, feed.ID, "Read", 200)
	if err := st.BatchUpdateItemsUnread([]int64{read.ID}, false); err != nil {
		t.Fatalf("mark read: %v", err)
	}

	w := performGReaderRequest(
		router, http.MethodGet,
		"/reader/api/0/stream/items/ids?s="+url.QueryEscape("user/-/state/com.google/kept-unread"),
		token, nil,
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var payload greaderItemRefsPayload
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(payload.ItemRefs) != 1 || payload.ItemRefs[0].ID != strconv.FormatInt(unread.ID, 10) {
		t.Fatalf("expected only the unread item, got %+v", payload.ItemRefs)
	}
}

// Unimplemented reader endpoints answer an empty list instead of 404, because
// clients probe endpoints such as preference/list and give up on a 404.
func TestGReaderUnimplementedEndpointsReturnEmptyList(t *testing.T) {
	h, _ := newGReaderTestHandler(t)
	router := h.SetupRouter()
	token := greaderClientLogin(t, router, "fusion", "secret")

	for _, method := range []string{http.MethodGet, http.MethodPost} {
		w := performGReaderRequest(router, method, "/reader/api/0/preference/list", token, nil)
		if w.Code != http.StatusOK || strings.TrimSpace(w.Body.String()) != "[]" {
			t.Fatalf("%s: expected 200 [], got %d %q", method, w.Code, w.Body.String())
		}
	}

	if w := performGReaderRequest(router, http.MethodGet, "/reader/api/0/preference/list", "", nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without a token, got %d", w.Code)
	}
	if w := performGReaderRequest(router, http.MethodPut, "/reader/api/0/preference/list", token, nil); w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for PUT, got %d", w.Code)
	}
	if w := performGReaderRequest(router, http.MethodGet, "/reader/api/0/user-info", token, nil); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "userName") {
		t.Fatalf("registered routes must still win, got %d %q", w.Code, w.Body.String())
	}
}

// Some clients authenticate writes only with the T form value and no header.
func TestGReaderAcceptsTFormTokenOnPost(t *testing.T) {
	h, st := newGReaderTestHandler(t)
	router := h.SetupRouter()
	token := greaderClientLogin(t, router, "fusion", "secret")

	group := mustCreateGReaderGroup(t, st, "Tech")
	feed := mustCreateGReaderFeed(t, st, group.ID, "Feed", "https://example.com/rss.xml", "https://example.com")
	item := mustCreateGReaderItem(t, st, feed.ID, "One", 100)

	values := url.Values{"i": {formatGReaderItemID(item.ID)}, "a": {greaderReadState}, "T": {token}}
	w := performGReaderRequest(router, http.MethodPost, "/reader/api/0/edit-tag", "", values)
	if w.Code != http.StatusOK || w.Body.String() != "OK" {
		t.Fatalf("expected T to authenticate the POST, got %d %q", w.Code, w.Body.String())
	}
	got, err := st.GetItem(item.ID)
	if err != nil {
		t.Fatalf("get item: %v", err)
	}
	if got.Unread {
		t.Fatal("expected the item to be read")
	}

	values.Set("T", "nope")
	if w := performGReaderRequest(router, http.MethodPost, "/reader/api/0/edit-tag", "", values); w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for a wrong T, got %d", w.Code)
	}
	// A query-string T would end up in access logs, so GET does not accept it.
	if w := performGReaderRequest(router, http.MethodGet, "/reader/api/0/user-info?T="+token, "", nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for T on GET, got %d", w.Code)
	}
}

func TestParseGReaderItemIDForms(t *testing.T) {
	tests := []struct {
		raw  string
		want int64
	}{
		{raw: "tag:google.com,2005:reader/item/000000000000048c", want: 0x48c},
		{raw: "tag:google.com,2005:reader/item/2f2", want: 0x2f2},
		{raw: "000000000000048c", want: 0x48c},
		// 16 characters is the bare hex form even when every digit is decimal.
		{raw: "0000000000000010", want: 16},
		{raw: "12345", want: 12345},
	}
	for _, tt := range tests {
		got, err := parseGReaderItemID(tt.raw)
		if err != nil || got != tt.want {
			t.Fatalf("parseGReaderItemID(%q) = %d, %v; want %d", tt.raw, got, err, tt.want)
		}
	}

	for _, raw := range []string{"", "0", "-1", "000000000000000g", "0000000000000000", "abc"} {
		if _, err := parseGReaderItemID(raw); err == nil {
			t.Fatalf("parseGReaderItemID(%q): expected an error", raw)
		}
	}
}

// The GReader and Fever usernames are configured independently.
func TestGReaderUsernameIsSeparateFromFever(t *testing.T) {
	h, _ := newGReaderTestHandlerWithConfig(t, &config.Config{FeverUsername: "fever-user", GReaderUsername: "reader-user"})
	router := h.SetupRouter()

	login := func(email string) int {
		return performRequest(
			router, http.MethodPost, "/accounts/ClientLogin",
			strings.NewReader(url.Values{"Email": {email}, "Passwd": {"secret"}}.Encode()),
			map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		).Code
	}
	if code := login("fever-user"); code != http.StatusUnauthorized {
		t.Fatalf("expected the Fever username to be rejected, got %d", code)
	}
	if code := login("reader-user"); code != http.StatusOK {
		t.Fatalf("expected the GReader username to be accepted, got %d", code)
	}

	token := greaderClientLogin(t, router, "reader-user", "secret")
	w := performGReaderRequest(router, http.MethodGet, "/reader/api/0/user-info", token, nil)
	if !strings.Contains(w.Body.String(), `"userName":"reader-user"`) {
		t.Fatalf("expected user-info to report the GReader username, got %q", w.Body.String())
	}
}
