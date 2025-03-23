package httpx

import (
	"net/http"
	"time"
)

type transportOptionFunc func(transport *http.Transport)

func newClient(options ...transportOptionFunc) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DisableKeepAlives = true
	transport.ForceAttemptHTTP2 = true

	for _, optionFunc := range options {
		optionFunc(transport)
	}

	return &http.Client{
		Transport: transport,
		Timeout:   1 * time.Minute, // fallback
	}
}
