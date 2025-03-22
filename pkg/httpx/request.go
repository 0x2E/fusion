package httpx

import (
	"context"
	"net/http"
	"net/url"

	"github.com/0x2e/fusion/model"
)

var globalClient = NewClient()

func FusionRequest(ctx context.Context, link string, options model.FeedRequestOptions) (*http.Response, error) {
	client := globalClient
	req, err := http.NewRequestWithContext(ctx, "GET", link, nil)
	if err != nil {
		return nil, err
	}
	req.Close = true
	req.Header.Add("User-Agent", "fusion/1.0")

	if options.ReqProxy != nil && *options.ReqProxy != "" {
		proxyURL, err := url.Parse(*options.ReqProxy)
		if err != nil {
			return nil, err
		}
		client = NewClient(func(transport *http.Transport) {
			transport.Proxy = http.ProxyURL(proxyURL)
		})
	}

	return client.Do(req)
}
