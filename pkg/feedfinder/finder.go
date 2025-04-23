package feedfinder

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
)

type FeedLink struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

type Finder struct {
	target     *url.URL
	httpClient *http.Client
}

type Options struct {
	ReqProxy *string
}

func Find(ctx context.Context, target *url.URL, options Options) ([]FeedLink, error) {
	clientTransportOps := []transportOptionFunc{}
	if options.ReqProxy != nil && *options.ReqProxy != "" {
		proxyURL, err := url.Parse(*options.ReqProxy)
		if err != nil {
			return nil, err
		}
		clientTransportOps = append(clientTransportOps, func(transport *http.Transport) {
			transport.Proxy = http.ProxyURL(proxyURL)
		})
	}

	finder := Finder{
		target:     target,
		httpClient: newClient(clientTransportOps...),
	}
	return finder.Run(context.Background())
}

func (f *Finder) Run(ctx context.Context) ([]FeedLink, error) {
	// find in third-party service
	logger := slog.With("step", "third-party service")
	fromService, err := f.tryService(ctx)
	if err != nil {
		logger.Error(err.Error())
	}
	if len(fromService) != 0 {
		return fromService, nil
	}

	feedMap := make(map[string]FeedLink)
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		// sniff in HTML
		logger := slog.With("step", "page")
		data, err := f.tryPageSource(ctx)
		if err != nil {
			logger.Error(err.Error())
		}

		mu.Lock()
		for _, f := range data {
			feedMap[f.Link] = f
		}
		mu.Unlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		// sniff well-knowns under this url
		logger := logger.With("step", "well-knowns")
		data, err := f.tryWellKnown(ctx, fmt.Sprintf("%s://%s%s", f.target.Scheme, f.target.Host, f.target.Path))
		if err != nil {
			logger.Error(err.Error())
		}
		if len(data) == 0 {
			// sniff well-knowns under root path
			data, err = f.tryWellKnown(ctx, fmt.Sprintf("%s://%s", f.target.Scheme, f.target.Host))
			if err != nil {
				logger.Error(err.Error())
			}
		}

		mu.Lock()
		for _, f := range data {
			feedMap[f.Link] = f
		}
		mu.Unlock()
	}()

	wg.Wait()
	res := make([]FeedLink, 0, len(feedMap))
	for _, f := range feedMap {
		res = append(res, f)
	}
	return res, nil
}

func isEmptyFeedLink(feed FeedLink) bool {
	return feed == FeedLink{}
}

func formatLinkToAbs(base, link string) string {
	if link == "" {
		return base
	}
	linkURL, err := url.Parse(link)
	if err != nil {
		return link
	}
	if linkURL.IsAbs() {
		return link
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return link
	}
	return baseURL.ResolveReference(linkURL).String()
}
