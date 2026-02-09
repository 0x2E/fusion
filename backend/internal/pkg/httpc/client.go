package httpc

import (
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

type clientPool struct {
	mu      sync.RWMutex
	clients map[string]*http.Client
}

var defaultClientPool = &clientPool{clients: make(map[string]*http.Client)}

// NewClient creates HTTP client with specified timeout and optional proxy.
// Clients are reused by (timeout, proxy) to keep connections warm.
func NewClient(timeout time.Duration, proxyURL string) (*http.Client, error) {
	key := proxyURL + "|" + strconv.FormatInt(timeout.Milliseconds(), 10)

	defaultClientPool.mu.RLock()
	if client, ok := defaultClientPool.clients[key]; ok {
		defaultClientPool.mu.RUnlock()
		return client, nil
	}
	defaultClientPool.mu.RUnlock()

	transport := &http.Transport{
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          128,
		MaxIdleConnsPerHost:   16,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: timeout,
	}

	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxy)
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	defaultClientPool.mu.Lock()
	if existing, ok := defaultClientPool.clients[key]; ok {
		defaultClientPool.mu.Unlock()
		transport.CloseIdleConnections()
		return existing, nil
	}
	defaultClientPool.clients[key] = client
	defaultClientPool.mu.Unlock()

	return client, nil
}

// SetDefaultHeaders adds default headers required for feed fetching.
func SetDefaultHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "fusion/1.0")
}
