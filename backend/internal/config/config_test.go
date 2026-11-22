package config

import (
	"strings"
	"testing"
)

func TestLoadParsesCORSAndPrivateFeedSettings(t *testing.T) {
	t.Setenv("FUSION_PASSWORD", "secret")
	t.Setenv("FUSION_CORS_ALLOWED_ORIGINS", " https://app.example.com , , https://admin.example.com/ ")
	t.Setenv("FUSION_TRUSTED_PROXIES", " 10.0.0.1 , 192.168.1.0/24 ")
	t.Setenv("FUSION_ALLOW_PRIVATE_FEEDS", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if len(cfg.CORSAllowedOrigins) != 2 {
		t.Fatalf("expected 2 allowed origins, got %d", len(cfg.CORSAllowedOrigins))
	}
	if cfg.CORSAllowedOrigins[0] != "https://app.example.com" {
		t.Fatalf("unexpected first origin: %q", cfg.CORSAllowedOrigins[0])
	}
	if cfg.CORSAllowedOrigins[1] != "https://admin.example.com/" {
		t.Fatalf("unexpected second origin: %q", cfg.CORSAllowedOrigins[1])
	}
	if !cfg.AllowPrivateFeeds {
		t.Fatal("expected AllowPrivateFeeds to be true")
	}
	if len(cfg.TrustedProxies) != 2 {
		t.Fatalf("expected 2 trusted proxies, got %d", len(cfg.TrustedProxies))
	}
	if cfg.TrustedProxies[0] != "10.0.0.1" {
		t.Fatalf("unexpected first trusted proxy: %q", cfg.TrustedProxies[0])
	}
	if cfg.TrustedProxies[1] != "192.168.1.0/24" {
		t.Fatalf("unexpected second trusted proxy: %q", cfg.TrustedProxies[1])
	}
}

func TestLoadUsesDefaultPullMaxBackoff(t *testing.T) {
	t.Setenv("FUSION_PASSWORD", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.PullMaxBackoff != 172800 {
		t.Fatalf("expected default PullMaxBackoff to be 172800, got %d", cfg.PullMaxBackoff)
	}
}

func TestLoadFeverUsername(t *testing.T) {
	t.Run("uses default username", func(t *testing.T) {
		t.Setenv("FUSION_PASSWORD", "secret")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if cfg.FeverUsername != "fusion" {
			t.Fatalf("expected default FeverUsername to be %q, got %q", "fusion", cfg.FeverUsername)
		}
	})

	t.Run("uses explicit username", func(t *testing.T) {
		t.Setenv("FUSION_PASSWORD", "secret")
		t.Setenv("FUSION_FEVER_USERNAME", " reader ")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if cfg.FeverUsername != "reader" {
			t.Fatalf("expected FeverUsername to be %q, got %q", "reader", cfg.FeverUsername)
		}
	})
}

func TestLoadParsesKubernetesStyleFusionPort(t *testing.T) {
	t.Setenv("FUSION_PASSWORD", "secret")
	t.Setenv("FUSION_PORT", "tcp://10.43.157.55:8080")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Port != 8080 {
		t.Fatalf("expected Port to be 8080, got %d", cfg.Port)
	}
}

func TestLoadRejectsInvalidFusionPort(t *testing.T) {
	t.Setenv("FUSION_PASSWORD", "secret")
	t.Setenv("FUSION_PORT", "tcp://10.43.157.55")

	_, err := Load()
	if err == nil {
		t.Fatal("expected Load() to fail for invalid FUSION_PORT")
	}
	if !strings.Contains(err.Error(), "invalid FUSION_PORT") {
		t.Fatalf("expected error to mention invalid FUSION_PORT, got %v", err)
	}
}

func TestLoadAuthConfiguration(t *testing.T) {
	tests := []struct {
		name       string
		password   string
		oidcIssuer string
		allowEmpty string
		wantErr    bool
	}{
		{name: "OIDC issuer without password", oidcIssuer: "https://id.example.com"},
		{name: "whitespace-only OIDC issuer", oidcIssuer: "   ", wantErr: true},
		{name: "neither password nor OIDC", wantErr: true},
		{name: "anonymous mode allowed", allowEmpty: "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("FUSION_PASSWORD", tt.password)
			t.Setenv("FUSION_OIDC_ISSUER", tt.oidcIssuer)
			t.Setenv("FUSION_ALLOW_EMPTY_PASSWORD", tt.allowEmpty)

			cfg, err := Load()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected Load() to fail")
				}
				if !strings.Contains(err.Error(), "FUSION_PASSWORD") || !strings.Contains(err.Error(), "FUSION_OIDC_ISSUER") {
					t.Fatalf("expected error to name both auth options, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}
			if cfg.OIDCIssuer != tt.oidcIssuer {
				t.Fatalf("expected OIDCIssuer %q, got %q", tt.oidcIssuer, cfg.OIDCIssuer)
			}
		})
	}
}
