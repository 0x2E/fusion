package httpx

import (
	"fmt"
	"net"
	"net/http"
	"syscall"
	"time"
)

func NewSafeClient() *http.Client {
	// avoid ssrf
	// https://www.agwa.name/blog/post/preventing_server_side_request_forgery_in_golang
	socketControl := func(network, address string, c syscall.RawConn) error {
		if !(network == "tcp4" || network == "tcp6") {
			return fmt.Errorf("banned network type: %s", network)
		}

		host, _, err := net.SplitHostPort(address)
		if err != nil {
			return fmt.Errorf("failed to split host:port: %s", err)
		}

		ipaddress := net.ParseIP(host)
		if ipaddress == nil {
			return fmt.Errorf("invalid ip: %s", host)
		}
		if ipaddress.IsLoopback() || ipaddress.IsPrivate() || ipaddress.IsUnspecified() {
			return fmt.Errorf("banned ip range: %s", ipaddress)
		}

		return nil
	}

	safeDialer := &net.Dialer{
		Timeout: 3 * time.Second,
		Control: socketControl,
	}
	safeTransport := &http.Transport{
		DialContext:       safeDialer.DialContext,
		ForceAttemptHTTP2: true,
		// DisableKeepAlives: true,
	}
	return &http.Client{
		Transport: safeTransport,
		Timeout:   5 * time.Second,
	}
}
