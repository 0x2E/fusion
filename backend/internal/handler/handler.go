package handler

import (
	"context"
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
	}
	sessions map[string]bool // sessionID -> valid, in-memory session store
	mu       sync.RWMutex    // protects sessions map
}

func New(store *store.Store, config *config.Config, puller interface {
	RefreshFeed(ctx context.Context, feedID int64) error
}) *Handler {
	// Hash password at startup for later verification
	passwordHash, err := auth.HashPassword(config.Password)
	if err != nil {
		panic("failed to hash password: " + err.Error())
	}

	return &Handler{
		store:        store,
		config:       config,
		passwordHash: passwordHash,
		puller:       puller,
		sessions:     make(map[string]bool),
	}
}

func (h *Handler) SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(h.corsMiddleware())

	api := r.Group("/api")
	{
		api.POST("/sessions", h.login)
		api.DELETE("/sessions", h.logout)

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
			auth.GET("/feeds/:id", h.getFeed)
			auth.PATCH("/feeds/:id", h.updateFeed)
			auth.DELETE("/feeds/:id", h.deleteFeed)
			auth.POST("/feeds/validate", h.validateFeed)
			auth.POST("/feeds/:id/refresh", h.refreshFeed)

			auth.GET("/items", h.listItems)
			auth.GET("/items/:id", h.getItem)
			auth.PATCH("/items/-/read", h.markItemsRead)
			auth.PATCH("/items/-/unread", h.markItemsUnread)

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
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
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
