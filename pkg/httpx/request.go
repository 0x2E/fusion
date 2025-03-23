package httpx

import (
	"context"
	"net/http"
	"net/url"

	"github.com/0x2e/fusion/model"
)

const UserAgentString = "fusion/1.0"

var globalClient = newClient()

// SendHTTPRequestFn is a function type for sending HTTP requests, matching
// http.Client's Do method.
type SendHTTPRequestFn func(req *http.Request) (*http.Response, error)

// FusionRequest makes an HTTP request using the global client.
func FusionRequest(ctx context.Context, link string, options model.FeedRequestOptions) (*http.Response, error) {
	client := globalClient

	if options.ReqProxy != nil && *options.ReqProxy != "" {
		proxyURL, err := url.Parse(*options.ReqProxy)
		if err != nil {
			return nil, err
		}
		client = newClient(func(transport *http.Transport) {
			transport.Proxy = http.ProxyURL(proxyURL)
		})
	}

	return FusionRequestWithRequestSender(ctx, client.Do, link, options)
}

// FusionRequestWithRequestSender makes an HTTP request using the provided
// request sender function.
func FusionRequestWithRequestSender(ctx context.Context, sendRequest SendHTTPRequestFn, link string, options model.FeedRequestOptions) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", link, nil)
	if err != nil {
		return nil, err
	}
	req.Close = true
	req.Header.Add("User-Agent", UserAgentString)

	return sendRequest(req)
}
