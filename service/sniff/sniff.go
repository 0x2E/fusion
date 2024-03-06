package sniff

import (
	"context"
	"log"
	"net/url"
	"sync"
)

type FeedLink struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

func Sniff(ctx context.Context, target *url.URL) ([]FeedLink, error) {
	// find in third-party service
	fromService, err := tryService(ctx, target)
	if err != nil {
		log.Printf("%s: %s\n", "parse service", err)
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
		data, err := tryPageSource(ctx, target.String())
		if err != nil {
			log.Printf("%s: %s\n", "parse page", err)
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
		data, err := tryWellKnown(ctx, target.Scheme+"://"+target.Host+target.Path) // https://go.dev/play/p/dVt-47_XWjU
		if err != nil {
			log.Printf("%s: %s\n", "parse wellknown", err)
		}
		if len(data) == 0 {
			// sniff well-knowns under url root
			data, err = tryWellKnown(ctx, target.Scheme+"://"+target.Host)
			if err != nil {
				log.Printf("%s: %s\n", "parse wellknown under root", err)
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
	return res, err
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
