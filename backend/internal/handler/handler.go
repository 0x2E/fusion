package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/0x2E/fusion/internal/auth"
	"github.com/0x2E/fusion/internal/config"
	"github.com/0x2E/fusion/internal/store"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	store           *store.Store
	config          *config.Config
	passwordHash    string   // bcrypt hash computed at startup
	passwordEnabled bool     // true when a non-blank password is configured
	feverAPIKey     string   // md5(username:password) used by Fever API; empty disables Fever
	allowAnonAPI    bool     // true when both password and OIDC auth are disabled
	notices         []string // stable notice codes served by GET /api/app; copy lives in the frontend
	puller          interface {
		RefreshFeed(ctx context.Context, feedID int64) error
		RefreshAll(ctx context.Context) (int, error)
		Wake()
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
	Wake()
}) (*Handler, error) {
	passwordEnabled := strings.TrimSpace(config.Password) != ""
	allowAnonAPI := !passwordEnabled && strings.TrimSpace(config.OIDCIssuer) == ""

	var passwordHash string
	if passwordEnabled {
		hash, err := auth.HashPassword(config.Password)
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}
		passwordHash = hash
	}

	h := &Handler{
		store:           store,
		config:          config,
		passwordHash:    passwordHash,
		passwordEnabled: passwordEnabled,
		allowAnonAPI:    allowAnonAPI,
		notices:         []string{},
		puller:          puller,
		sessions:        make(map[string]int64),
		limiter:         newLoginLimiter(config.LoginRateLimit, config.LoginWindow, config.LoginBlock),
	}

	switch {
	case passwordEnabled:
		h.feverAPIKey = deriveFeverAPIKey(config.FeverUsername, config.Password)
	case allowAnonAPI:
		// The whole API is already open in anonymous mode, so Fever keeps
		// accepting the key derived from an empty password, as it always has.
		h.feverAPIKey = deriveFeverAPIKey(config.FeverUsername, "")
	default:
		// Leave the key empty: verifyFeverAPIKey rejects every request when the
		// expected key is empty, so no client can authenticate.
		slog.Warn("Fever API is disabled because no password is set")
	}

	if h.allowAnonAPI {
		slog.Warn("authentication is disabled because both password and OIDC are empty")
	}

	if config.OIDCIssuer != "" {
		oidcAuth, err := auth.NewOIDC(
			context.Background(),
			config.OIDCIssuer,
			config.OIDCClientID,
			config.OIDCClientSecret,
		)
		if err != nil {
			return nil, fmt.Errorf("initialize OIDC: %w", err)
		}
		if allowedUser := strings.TrimSpace(config.OIDCAllowedUser); allowedUser != "" {
			oidcAuth.SetAllowedUser(allowedUser)
		} else {
			// TODO(v1.5.0): turn this notice into a startup error instead, so a
			// deployment cannot silently accept every account of its IdP.
			h.notices = append(h.notices, noticeOIDCAllowedUserUnset)
			slog.Warn(
				"FUSION_OIDC_ALLOWED_USER is not set, so any account on the OIDC provider can sign in; a future version will require it",
				"issuer", config.OIDCIssuer,
			)
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
		// Password login only exists when a password is configured. OIDC-only
		// deployments still need session status and logout.
		if h.passwordEnabled {
			api.POST("/sessions", h.login)
		}
		api.GET("/sessions", h.sessionStatus)
		api.DELETE("/sessions", h.logout)

		// Public auth discovery and OIDC routes (no auth middleware).
		api.GET("/auth/methods", h.authMethods)
		if h.oidcAuth != nil {
			api.GET("/oidc/login", h.oidcLogin)
			api.GET("/oidc/callback", h.oidcCallback)
			// Compatibility route for deployments that configured redirect_uri without /api.
			r.GET("/oidc/callback", h.oidcCallback)
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

			auth.GET("/app", h.getAppInfo)

			auth.GET("/items", h.listItems)
			auth.GET("/items/:id", h.getItem)
			auth.PATCH("/items/-/read", h.markItemsRead)
			auth.PATCH("/items/-/unread", h.markItemsUnread)

			auth.GET("/search", h.search)

			auth.GET("/bookmarks", h.listBookmarks)
			auth.POST("/bookmarks", h.createBookmark)
			auth.GET("/bookmarks/:id", h.getBookmark)
			auth.DELETE("/bookmarks/:id", h.deleteBookmark)

			auth.GET("/settings", h.getSettings)
			auth.PATCH("/settings", h.updateSettings)
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
	allowedOrigins, crossOrigin := h.corsPolicy()

	return func(c *gin.Context) {
		// Cross-origin reads are only readable by a browser when the response
		// names the calling origin, so the allowlist is enforced by withholding
		// these headers rather than by rejecting the request.
		if origin := normalizeOrigin(c.Request.Header.Get("Origin")); origin != "" && allowedOrigins[origin] {
			header := c.Writer.Header()
			header.Set("Access-Control-Allow-Origin", origin)
			header.Set("Vary", "Origin")
			header.Set("Access-Control-Allow-Credentials", "true")
			header.Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			header.Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")
		}

		// Withholding CORS headers stops browsers from reading responses, but a
		// simple request (e.g. a text/plain POST, which needs no preflight) still
		// reaches the handler, so state-changing cross-origin requests are
		// rejected here.
		if err := crossOrigin.Check(c.Request); err != nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// corsPolicy builds both halves of the cross-origin policy from the configured
// allowlist: the origins allowed to read responses, and the guard for
// state-changing requests. Origins are normalized before use so configuration
// matching follows browser Origin formatting (lowercase, no trailing slash).
// Entries that are not valid origins (a leftover "*") are dropped with a
// warning: they cannot be trusted for credentialed access, and failing startup
// would leave deployments that copied the old advice unable to start at all.
func (h *Handler) corsPolicy() (map[string]bool, *http.CrossOriginProtection) {
	allowed := make(map[string]bool)
	crossOrigin := http.NewCrossOriginProtection()
	if h.config == nil {
		return allowed, crossOrigin
	}

	for _, origin := range h.config.CORSAllowedOrigins {
		normalized := normalizeOrigin(origin)
		if err := crossOrigin.AddTrustedOrigin(normalized); err != nil {
			slog.Warn("ignoring invalid CORS allowed origin", "origin", origin, "error", err)
			continue
		}
		allowed[normalized] = true
	}

	return allowed, crossOrigin
}

func normalizeOrigin(origin string) string {
	origin = strings.TrimSpace(origin)
	origin = strings.TrimSuffix(origin, "/")
	return strings.ToLower(origin)
}

func (h *Handler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.allowAnonAPI {
			c.Next()
			return
		}

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

// paginatedListResponse emits a list payload with a cursor-based next_cursor.
// nextCursor is non-nil only when more pages may exist; nil means "no more".
func paginatedListResponse(c *gin.Context, data any, total int, nextCursor *string) {
	c.JSON(200, gin.H{"data": data, "total": total, "next_cursor": nextCursor})
}

// parseCursor decodes a "<value>_<id>" cursor into its two int64 components.
// Shared by list endpoints that paginate on a composite (timestamp, id) key.
func parseCursor(cursor string) (first int64, second int64, err error) {
	parts := strings.SplitN(cursor, "_", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("malformed cursor")
	}
	first, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("malformed cursor")
	}
	second, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("malformed cursor")
	}
	return first, second, nil
}
