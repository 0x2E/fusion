package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/0x2E/fusion/internal/config"
	"github.com/0x2E/fusion/internal/pkg/httpc"
)

type validateFeedBody struct {
	Data struct {
		Feeds []struct {
			Title string `json:"title"`
			Link  string `json:"link"`
		} `json:"feeds"`
	} `json:"data"`
}

func postValidateFeed(t *testing.T, h *Handler, url string) (int, validateFeedBody) {
	t.Helper()

	r := newTestRouter()
	r.POST("/api/feeds/validate", h.validateFeed)
	w := performRequest(
		r,
		http.MethodPost,
		"/api/feeds/validate",
		strings.NewReader(fmt.Sprintf(`{"url":%q}`, url)),
		map[string]string{"Content-Type": "application/json"},
	)

	var body validateFeedBody
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil && w.Code == http.StatusOK {
		t.Fatalf("decode response: %v", err)
	}
	return w.Code, body
}

// The add-feed dialog calls validate before creating a feed and blocks the
// creation when no feed is found, so a malformed document must yield an empty
// feeds list rather than an error or a fabricated entry.
func TestValidateFeedMalformedXMLReturnsNoFeeds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = fmt.Fprint(w, `<?xml version="1.0"?><rss><channel><title>Broken`)
	}))
	defer server.Close()

	h := &Handler{config: &config.Config{AllowPrivateFeeds: true}}

	code, body := postValidateFeed(t, h, server.URL)
	if code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", code)
	}
	if len(body.Data.Feeds) != 0 {
		t.Fatalf("expected 0 feeds for malformed XML, got %d: %+v", len(body.Data.Feeds), body.Data.Feeds)
	}
}

func TestValidateFeedValidRSSReturnsFeed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Serve the feed only at its real path so the finder's common-path
		// guesses (index.xml, atom.xml, ...) miss like they would in production.
		if r.URL.Path != "/feed.xml" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel><title>Tech Daily</title><link>https://example.com</link>
<item><guid>g1</guid><title>Item</title><link>https://example.com/1</link></item>
</channel></rss>`)
	}))
	defer server.Close()

	h := &Handler{config: &config.Config{AllowPrivateFeeds: true}}

	code, body := postValidateFeed(t, h, server.URL+"/feed.xml")
	if code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", code)
	}
	if len(body.Data.Feeds) != 1 {
		t.Fatalf("expected 1 feed, got %d: %+v", len(body.Data.Feeds), body.Data.Feeds)
	}
	if body.Data.Feeds[0].Title != "Tech Daily" {
		t.Fatalf("expected title %q, got %q", "Tech Daily", body.Data.Feeds[0].Title)
	}
	if body.Data.Feeds[0].Link != server.URL+"/feed.xml" {
		t.Fatalf("expected link %q, got %q", server.URL+"/feed.xml", body.Data.Feeds[0].Link)
	}
}

func TestValidateFeedUnreachableURLReturnsNoFeeds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	url := server.URL
	server.Close()

	h := &Handler{config: &config.Config{AllowPrivateFeeds: true}}

	code, body := postValidateFeed(t, h, url)
	if code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", code)
	}
	if len(body.Data.Feeds) != 0 {
		t.Fatalf("expected 0 feeds for unreachable URL, got %d", len(body.Data.Feeds))
	}
}

// Discovery probes common paths on the host, so a feed elsewhere on the host
// makes /feeds/validate return a non-empty list even though the entered URL is
// not a feed itself. The frontend must therefore match the entered URL against
// the returned links instead of treating a non-empty list as approval.
func TestValidateFeedReturnsCommonPathFeedForInvalidPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/feed.xml":
			w.Header().Set("Content-Type", "application/rss+xml")
			_, _ = fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel><title>Tech Daily</title><link>https://example.com</link>
<item><guid>g1</guid><title>Item</title><link>https://example.com/1</link></item>
</channel></rss>`)
		default:
			w.Header().Set("Content-Type", "application/xml")
			_, _ = fmt.Fprint(w, `<?xml version="1.0"?><rss><channel><title>Broken`)
		}
	}))
	defer server.Close()

	h := &Handler{config: &config.Config{AllowPrivateFeeds: true}}

	code, body := postValidateFeed(t, h, server.URL+"/broken.xml")
	if code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", code)
	}
	found := false
	for _, feed := range body.Data.Feeds {
		if feed.Link == server.URL+"/feed.xml" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected discovery to surface %q for a broken entered path, got %+v", server.URL+"/feed.xml", body.Data.Feeds)
	}
}

// testPublicIP is in TEST-NET-1, so ValidateRequestURL treats it as public
// while the test dialer can route it to a local httptest server.
const testPublicIP = "192.0.2.10"

// routeFakePublicServer makes server reachable at testPublicIP. Only that
// address is rerouted; every other address still goes through httpc's SSRF
// checks. feedfinder's default client gets the same route so a regression that
// drops HTTPClient injection reaches the server instead of a public address.
func routeFakePublicServer(t *testing.T, server *httptest.Server) string {
	t.Helper()

	parsed, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse test server url: %v", err)
	}
	fakeAddr := net.JoinHostPort(testPublicIP, parsed.Port())
	targetAddr := parsed.Host

	withRoute := func(transport *http.Transport) *http.Transport {
		transport = transport.Clone()
		dial := transport.DialContext
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			if addr == fakeAddr {
				return (&net.Dialer{}).DialContext(ctx, network, targetAddr)
			}
			return dial(ctx, network, addr)
		}
		return transport
	}

	originalFactory := newFeedDiscoveryClient
	newFeedDiscoveryClient = func(timeout time.Duration, proxyURL string, allowPrivate bool) (*http.Client, error) {
		client, err := httpc.NewClient(timeout, proxyURL, allowPrivate)
		if err != nil {
			return nil, err
		}
		return &http.Client{
			Timeout:       client.Timeout,
			Transport:     withRoute(client.Transport.(*http.Transport)),
			CheckRedirect: client.CheckRedirect,
		}, nil
	}
	t.Cleanup(func() { newFeedDiscoveryClient = originalFactory })

	originalDefaultTransport := http.DefaultTransport
	http.DefaultTransport = withRoute(originalDefaultTransport.(*http.Transport))
	t.Cleanup(func() { http.DefaultTransport = originalDefaultTransport })

	return "http://" + fakeAddr
}

func TestValidateFeedBlocksRedirectToInternalAddress(t *testing.T) {
	var internalHits atomic.Int32
	internal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		internalHits.Add(1)
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = fmt.Fprint(w, `<?xml version="1.0"?><rss version="2.0"><channel><title>Internal</title></channel></rss>`)
	}))
	defer internal.Close()

	start := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, internal.URL+"/feed.xml", http.StatusFound)
	}))
	defer start.Close()

	publicURL := routeFakePublicServer(t, start)
	h := &Handler{config: &config.Config{AllowPrivateFeeds: false}}

	code, body := postValidateFeed(t, h, publicURL+"/page")
	if code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", code)
	}
	if got := internalHits.Load(); got != 0 {
		t.Fatalf("expected redirect to internal address not to be followed, got %d requests", got)
	}
	for _, feed := range body.Data.Feeds {
		if strings.Contains(feed.Link, internal.URL) {
			t.Fatalf("internal redirect target leaked into feeds: %+v", body.Data.Feeds)
		}
	}
}

func TestValidateFeedDoesNotProbeInternalLinks(t *testing.T) {
	var internalHits atomic.Int32
	internal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		internalHits.Add(1)
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = fmt.Fprint(w, `<?xml version="1.0"?><rss version="2.0"><channel><title>Internal</title></channel></rss>`)
	}))
	defer internal.Close()

	start := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/page" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		_, _ = fmt.Fprintf(w, `<html><head><title>Example</title></head><body><a href="%s/feed.xml">rss feed</a></body></html>`, internal.URL)
	}))
	defer start.Close()

	publicURL := routeFakePublicServer(t, start)
	h := &Handler{config: &config.Config{AllowPrivateFeeds: false}}

	code, body := postValidateFeed(t, h, publicURL+"/page")
	if code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", code)
	}
	if got := internalHits.Load(); got != 0 {
		t.Fatalf("expected internal link not to be probed, got %d requests", got)
	}
	for _, feed := range body.Data.Feeds {
		if strings.Contains(feed.Link, internal.URL) {
			t.Fatalf("internal link leaked into feeds: %+v", body.Data.Feeds)
		}
	}
}

func TestValidateFeedDiscoversPrivateFeedWhenAllowed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/feed.xml" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = fmt.Fprint(w, `<?xml version="1.0"?><rss version="2.0"><channel><title>Private Feed</title></channel></rss>`)
	}))
	defer server.Close()

	target := server.URL + "/feed.xml"

	h := &Handler{config: &config.Config{AllowPrivateFeeds: true}}

	code, body := postValidateFeed(t, h, target)
	if code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", code)
	}
	if len(body.Data.Feeds) != 1 {
		t.Fatalf("expected 1 feed, got %d: %+v", len(body.Data.Feeds), body.Data.Feeds)
	}
	if body.Data.Feeds[0].Title != "Private Feed" {
		t.Fatalf("expected title %q, got %q", "Private Feed", body.Data.Feeds[0].Title)
	}
}
