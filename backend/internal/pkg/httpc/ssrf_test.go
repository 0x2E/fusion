package httpc

import (
	"context"
	"net/http"
	"testing"
)

func TestValidateRequestURL(t *testing.T) {
	tests := []struct {
		name         string
		rawURL       string
		allowPrivate bool
		wantErr      bool
	}{
		{name: "public ipv4", rawURL: "http://93.184.216.34/feed.xml", wantErr: false},
		{name: "public ipv6", rawURL: "https://[2001:4860:4860::8888]/feed.xml", wantErr: false},
		{name: "loopback blocked", rawURL: "http://127.0.0.1/feed.xml", wantErr: true},
		{name: "localhost blocked", rawURL: "http://localhost/feed.xml", wantErr: true},
		{name: "private blocked", rawURL: "http://10.0.0.8/feed.xml", wantErr: true},
		{name: "unsupported scheme", rawURL: "file:///tmp/a", wantErr: true},
		{name: "allow private", rawURL: "http://127.0.0.1/feed.xml", allowPrivate: true, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequestURL(context.Background(), tt.rawURL, tt.allowPrivate)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateRequestURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRedirectValidatorBlocksPrivateTargets(t *testing.T) {
	t.Run("block private target", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://127.0.0.1/feed.xml", nil)
		if err != nil {
			t.Fatalf("new request: %v", err)
		}

		if err := redirectValidator(false)(req, nil); err == nil {
			t.Fatal("expected private redirect target to be blocked")
		}
	})

	t.Run("allow private target", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://127.0.0.1/feed.xml", nil)
		if err != nil {
			t.Fatalf("new request: %v", err)
		}

		if err := redirectValidator(true)(req, nil); err != nil {
			t.Fatalf("expected private redirect target to be allowed, got %v", err)
		}
	})
}

func TestValidateDialAddressBlocksPrivateHost(t *testing.T) {
	tests := []struct {
		name         string
		address      string
		allowPrivate bool
		wantErr      bool
	}{
		{name: "ipv4 loopback", address: "127.0.0.1:8080", wantErr: true},
		{name: "ipv6 loopback", address: "[::1]:8080", wantErr: true},
		{name: "ipv4 private", address: "10.0.0.8:8080", wantErr: true},
		{name: "ipv4 mapped loopback", address: "[::ffff:127.0.0.1]:8080", wantErr: true},
		{name: "ipv4 mapped private", address: "[::ffff:10.0.0.8]:8080", wantErr: true},
		{name: "public ipv4", address: "93.184.216.34:443", wantErr: false},
		{name: "public ipv6", address: "[2606:4700::1]:443", wantErr: false},
		{name: "allow private ipv4 loopback", address: "127.0.0.1:8080", allowPrivate: true, wantErr: false},
		{name: "allow private ipv6 loopback", address: "[::1]:8080", allowPrivate: true, wantErr: false},
		{name: "allow private ipv4", address: "10.0.0.8:8080", allowPrivate: true, wantErr: false},
		{name: "allow private ipv4 mapped loopback", address: "[::ffff:127.0.0.1]:8080", allowPrivate: true, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDialAddress(tt.address, tt.allowPrivate)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateDialAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
