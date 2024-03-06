package sniff

import (
	"context"
	"net/http"

	"github.com/0x2e/fusion/pkg/httpx"
)

var globalClient *http.Client

func init() {
	globalClient = httpx.NewSafeClient()
}

func request(ctx context.Context, link string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", link, nil)
	if err != nil {
		return nil, err
	}

	ua := "fusion/1.0"
	req.Header.Add("User-Agent", ua)

	return globalClient.Do(req)
}
