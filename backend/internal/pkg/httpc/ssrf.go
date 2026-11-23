package httpc

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

func ValidateRequestURL(ctx context.Context, rawURL string, allowPrivate bool) error {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(rawURL))
	if err != nil {
		return err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("unsupported url scheme: %s", parsed.Scheme)
	}

	host := strings.TrimSpace(parsed.Hostname())
	if host == "" {
		return fmt.Errorf("url host is required")
	}

	if !allowPrivate {
		if err := validatePublicHost(ctx, host); err != nil {
			return err
		}
	}

	return nil
}

func validatePublicHost(ctx context.Context, host string) error {
	if strings.EqualFold(host, "localhost") {
		return fmt.Errorf("private host is not allowed")
	}

	if ip := net.ParseIP(host); ip != nil {
		if isPrivateOrLocalIP(ip) {
			return fmt.Errorf("private host is not allowed")
		}
		return nil
	}

	resolveCtx := ctx
	if resolveCtx == nil {
		resolveCtx = context.Background()
	}

	lookupCtx, cancel := context.WithTimeout(resolveCtx, 2*time.Second)
	defer cancel()

	addrs, err := net.DefaultResolver.LookupIPAddr(lookupCtx, host)
	if err != nil {
		return fmt.Errorf("resolve host: %w", err)
	}
	if len(addrs) == 0 {
		return fmt.Errorf("resolve host: no addresses")
	}

	for _, addr := range addrs {
		if isPrivateOrLocalIP(addr.IP) {
			return fmt.Errorf("private host is not allowed")
		}
	}

	return nil
}

// validateDialAddress checks the resolved IP a dialer is about to connect to.
// It runs from net.Dialer.ControlContext, where address is the actual ip:port
// being dialed, so a second DNS lookup cannot return a private address after
// ValidateRequestURL saw a public one.
func validateDialAddress(address string, allowPrivate bool) error {
	if allowPrivate {
		return nil
	}

	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("invalid dial address: %w", err)
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return fmt.Errorf("invalid dial address: %s", address)
	}
	if isPrivateOrLocalIP(ip) {
		return fmt.Errorf("private host is not allowed")
	}

	return nil
}

func isPrivateOrLocalIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() || ip.IsMulticast() || ip.IsInterfaceLocalMulticast() || ip.IsUnspecified()
}
