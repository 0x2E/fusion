package sniff

import (
	"context"
	"net/url"
	"sync"

	"github.com/0x2e/fusion/pkg/logx"
)

var sniffLogger = logx.Logger.With("module", "sniffer")

type FeedLink struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

func Sniff(ctx context.Context, target *url.URL) ([]FeedLink, error) {
	logger := sniffLogger.With("url", target.String())
	ctx = logx.ContextWithLogger(ctx, logger)

	// find in third-party service
	sLogger := logger.With("step", "third-party service")
	fromService, err := tryService(logx.ContextWithLogger(ctx, sLogger), target)
	if err != nil {
		sLogger.Errorln(err)
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
		pLogger := logger.With("step", "page")
		data, err := tryPageSource(
			logx.ContextWithLogger(ctx, pLogger),
			target.String(),
		)
		if err != nil {
			pLogger.Errorln(err)
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
		wLogger := logger.With("step", "well-knowns")
		data, err := tryWellKnown(
			logx.ContextWithLogger(ctx, wLogger),
			target.Scheme+"://"+target.Host+target.Path,
		) // https://go.dev/play/p/dVt-47_XWjU
		if err != nil {
			wLogger.Errorln(err)
		}
		if len(data) == 0 {
			// sniff well-knowns under url root
			data, err = tryWellKnown(ctx, target.Scheme+"://"+target.Host)
			if err != nil {
				wLogger.Errorln(err)
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
