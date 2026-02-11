package config

import "testing"

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
