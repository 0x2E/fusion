package httpc

import (
	"net/http"
	"net/url"
	"time"
)

// NewClient creates HTTP client with specified timeout and optional proxy.
// Returns client configured for HTTP/2 with keep-alives disabled.
func NewClient(timeout time.Duration, proxyURL string) (*http.Client, error) {
	transport := &http.Transport{
		DisableKeepAlives: true,
		ForceAttemptHTTP2: true,
	}

	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxy)
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}, nil
}

// SetDefaultHeaders adds default headers required for feed fetching.
func SetDefaultHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "fusion/1.0")
	req.Header.Set("Connection", "close")
}
