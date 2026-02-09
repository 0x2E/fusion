package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/0x2E/fusion/internal/auth"
	"github.com/0x2E/fusion/internal/config"
	"github.com/0x2E/fusion/internal/store"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	store        *store.Store
	config       *config.Config
	passwordHash string // bcrypt hash computed at startup
	puller       interface {
		RefreshFeed(ctx context.Context, feedID int64) error
		RefreshAll(ctx context.Context) (int, error)
	}
	sessions map[string]bool         // sessionID -> valid, in-memory session store
	mu       sync.RWMutex            // protects sessions map
	oidcAuth *auth.OIDCAuthenticator // nil when OIDC is disabled
	limiter  *loginLimiter
}

func New(store *store.Store, config *config.Config, puller interface {
	RefreshFeed(ctx context.Context, feedID int64) error
	RefreshAll(ctx context.Context) (int, error)
}) (*Handler, error) {
	// Hash password at startup for later verification
	passwordHash, err := auth.HashPassword(config.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	h := &Handler{
		store:        store,
		config:       config,
		passwordHash: passwordHash,
		puller:       puller,
		sessions:     make(map[string]bool),
		limiter:      newLoginLimiter(config.LoginRateLimit, config.LoginWindow, config.LoginBlock),
	}

	if config.OIDCIssuer != "" {
		if strings.TrimSpace(config.OIDCRedirectURI) == "" {
			return nil, fmt.Errorf("FUSION_OIDC_REDIRECT_URI is required when OIDC is enabled")
		}

		oidcAuth, err := auth.NewOIDC(
			context.Background(),
			config.OIDCIssuer,
			config.OIDCClientID,
			config.OIDCClientSecret,
			config.OIDCRedirectURI,
		)
		if err != nil {
			return nil, fmt.Errorf("initialize OIDC: %w", err)
		}
		if config.OIDCAllowedUser != "" {
			oidcAuth.SetAllowedUser(config.OIDCAllowedUser)
		}
		h.oidcAuth = oidcAuth
		slog.Info("OIDC authentication enabled", "issuer", config.OIDCIssuer)
	}

	return h, nil
}

func (h *Handler) SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(h.corsMiddleware())

	api := r.Group("/api")
	{
		api.POST("/sessions", h.login)
		api.DELETE("/sessions", h.logout)

		// OIDC routes (public, no auth middleware)
		api.GET("/oidc/enabled", h.oidcEnabled)
		if h.oidcAuth != nil {
			api.GET("/oidc/login", h.oidcLogin)
			api.GET("/oidc/callback", h.oidcCallback)
		}

		auth := api.Group("")
		auth.Use(h.authMiddleware())
		{
			auth.GET("/groups", h.listGroups)
			auth.POST("/groups", h.createGroup)
			auth.GET("/groups/:id", h.getGroup)
			auth.PATCH("/groups/:id", h.updateGroup)
			auth.DELETE("/groups/:id", h.deleteGroup)

			auth.GET("/feeds", h.listFeeds)
			auth.POST("/feeds", h.createFeed)
			auth.POST("/feeds/batch", h.batchCreateFeeds)
			auth.POST("/feeds/refresh", h.refreshAllFeeds)
			auth.GET("/feeds/:id", h.getFeed)
			auth.PATCH("/feeds/:id", h.updateFeed)
			auth.DELETE("/feeds/:id", h.deleteFeed)
			auth.POST("/feeds/validate", h.validateFeed)
			auth.POST("/feeds/:id/refresh", h.refreshFeed)

			auth.GET("/items", h.listItems)
			auth.GET("/items/:id", h.getItem)
			auth.PATCH("/items/-/read", h.markItemsRead)
			auth.PATCH("/items/-/unread", h.markItemsUnread)

			auth.GET("/search", h.search)

			auth.GET("/bookmarks", h.listBookmarks)
			auth.POST("/bookmarks", h.createBookmark)
			auth.GET("/bookmarks/:id", h.getBookmark)
			auth.DELETE("/bookmarks/:id", h.deleteBookmark)
		}
	}

	return r
}

func (h *Handler) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			// Cookie-based auth needs a concrete origin ("*" + credentials is rejected by browsers).
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func (h *Handler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session")
		if err != nil {
			unauthorizedError(c)
			c.Abort()
			return
		}

		h.mu.RLock()
		valid := h.sessions[sessionID]
		h.mu.RUnlock()

		if !valid {
			unauthorizedError(c)
			c.Abort()
			return
		}

		c.Next()
	}
}

func dataResponse(c *gin.Context, data interface{}) {
	c.JSON(200, gin.H{"data": data})
}

func listResponse(c *gin.Context, data interface{}, total int) {
	c.JSON(200, gin.H{"data": data, "total": total})
}
