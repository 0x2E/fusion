package handler

import (
	"testing"

	"github.com/0x2E/fusion/internal/config"
	"github.com/gin-gonic/gin"
)

func TestConfigureTrustedProxies(t *testing.T) {
	t.Run("default disables proxy trust", func(t *testing.T) {
		h := &Handler{config: &config.Config{}}
		r := gin.New()

		if err := h.configureTrustedProxies(r); err != nil {
			t.Fatalf("configureTrustedProxies() failed: %v", err)
		}
	})

	t.Run("invalid trusted proxy config returns error", func(t *testing.T) {
		h := &Handler{config: &config.Config{TrustedProxies: []string{"not-an-ip"}}}
		r := gin.New()

		if err := h.configureTrustedProxies(r); err == nil {
			t.Fatal("expected configureTrustedProxies() to fail")
		}
	})
}
