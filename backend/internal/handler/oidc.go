package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) oidcEnabled(c *gin.Context) {
	dataResponse(c, gin.H{"enabled": h.oidcAuth != nil})
}

func (h *Handler) oidcLogin(c *gin.Context) {
	if h.oidcAuth == nil {
		badRequestError(c, "OIDC is not configured")
		return
	}

	// Auto-detect redirect URI from Host header if not explicitly configured
	if h.oidcAuth.RedirectURI() == "" {
		scheme := "https"
		if !isSecureRequest(c.Request) {
			scheme = "http"
		}
		h.oidcAuth.SetRedirectURI(fmt.Sprintf("%s://%s/api/oidc/callback", scheme, c.Request.Host))
	}

	authURL, err := h.oidcAuth.AuthURL()
	if err != nil {
		internalError(c, err, "oidc auth url")
		return
	}

	dataResponse(c, gin.H{"auth_url": authURL})
}

func (h *Handler) oidcCallback(c *gin.Context) {
	if h.oidcAuth == nil {
		badRequestError(c, "OIDC is not configured")
		return
	}

	state := c.Query("state")
	code := c.Query("code")
	if state == "" || code == "" {
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=oidc_failed")
		return
	}

	userID, err := h.oidcAuth.Callback(c.Request.Context(), state, code)
	if err != nil {
		slog.Error("OIDC callback failed", "error", err)
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=oidc_failed")
		return
	}

	slog.Info("OIDC login successful", "user", userID)
	h.createSession(c)
	c.Redirect(http.StatusTemporaryRedirect, "/")
}
