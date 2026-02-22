package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
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
	feverAPIKey  string // md5(username:password) used by Fever API
	puller       interface {
		RefreshFeed(ctx context.Context, feedID int64) error
		RefreshAll(ctx context.Context) (int, error)
	}
	sessions  map[string]int64        // sessionID -> unix expiry seconds
	mu        sync.RWMutex            // protects sessions state
	oidcAuth  *auth.OIDCAuthenticator // nil when OIDC is disabled
	limiter   *loginLimiter
	lastSweep int64

	refreshAllMu      sync.Mutex
	refreshAllRunning bool
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
		feverAPIKey:  deriveFeverAPIKey(config.FeverUsername, config.Password),
		puller:       puller,
		sessions:     make(map[string]int64),
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
	r := gin.New()
	r.Use(requestLogMiddleware(), recoveryMiddleware())

	if err := h.configureTrustedProxies(r); err != nil {
		slog.Warn("failed to configure trusted proxies", "error", err)
	}

	r.Use(h.corsMiddleware())
	r.POST("/fever", h.fever)
	r.POST("/fever/", h.fever)
	r.POST("/fever.php", h.fever)

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

	if err := h.setupFrontendRoutes(r); err != nil {
		slog.Warn("failed to configure frontend routes", "error", err)
	}

	return r
}

func (h *Handler) configureTrustedProxies(r *gin.Engine) error {
	if h.config == nil || len(h.config.TrustedProxies) == 0 {
		return r.SetTrustedProxies(nil)
	}

	return r.SetTrustedProxies(h.config.TrustedProxies)
}

func (h *Handler) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.Request.Header.Get("Origin"))
		if origin != "" {
			if !h.isOriginAllowed(origin) {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
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

func (h *Handler) isOriginAllowed(origin string) bool {
	if h.config == nil {
		return true
	}

	if len(h.config.CORSAllowedOrigins) == 0 {
		return true
	}

	normalizedOrigin := normalizeOrigin(origin)
	for _, allowed := range h.config.CORSAllowedOrigins {
		normalizedAllowed := normalizeOrigin(allowed)
		if normalizedAllowed == "*" || normalizedAllowed == normalizedOrigin {
			return true
		}
	}

	return false
}

func normalizeOrigin(origin string) string {
	origin = strings.TrimSpace(origin)
	origin = strings.TrimSuffix(origin, "/")
	return strings.ToLower(origin)
}

func (h *Handler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session")
		if err != nil {
			unauthorizedError(c)
			c.Abort()
			return
		}

		if !h.isSessionValid(sessionID) {
			unauthorizedError(c)
			c.Abort()
			return
		}

		c.Next()
	}
}

func dataResponse(c *gin.Context, data any) {
	c.JSON(200, gin.H{"data": data})
}

func listResponse(c *gin.Context, data any, total int) {
	c.JSON(200, gin.H{"data": data, "total": total})
}
