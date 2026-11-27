package handler

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/0x2E/fusion/internal/web"
	"github.com/gin-gonic/gin"
)

// setupFrontendRoutes mounts the static assets and returns the SPA handler for
// unmatched routes. SetupRouter owns NoRoute because the Google Reader API
// also needs a catch-all there.
func (h *Handler) setupFrontendRoutes(r *gin.Engine) (gin.HandlerFunc, error) {
	frontendFS, _, err := web.FrontendFS()
	if err != nil {
		return nil, err
	}

	r.StaticFS("/assets", http.FS(frontendFS))

	fileServer := http.FileServer(http.FS(frontendFS))
	return func(c *gin.Context) {
		serveFrontendRoute(c, frontendFS, fileServer)
	}, nil
}

func serveFrontendRoute(c *gin.Context, frontendFS fs.FS, fileServer http.Handler) {
	if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
		c.Status(http.StatusNotFound)
		return
	}

	setFrontendSecurityHeaders(c)

	cleanedPath := path.Clean(c.Request.URL.Path)
	if cleanedPath == "." {
		cleanedPath = "/"
	}

	if isBackendPath(cleanedPath) {
		c.Status(http.StatusNotFound)
		return
	}

	if cleanedPath == "/" {
		serveFrontendIndex(c, fileServer)
		return
	}

	assetPath := strings.TrimPrefix(cleanedPath, "/")
	if assetPath == "" {
		assetPath = "index.html"
	}

	if frontendFileExists(frontendFS, assetPath) {
		serveFrontendRequestPath(c, fileServer, "/"+assetPath)
		return
	}

	if looksLikeAssetPath(assetPath) {
		c.Status(http.StatusNotFound)
		return
	}

	serveFrontendIndex(c, fileServer)
}

func serveFrontendIndex(c *gin.Context, fileServer http.Handler) {
	serveFrontendRequestPath(c, fileServer, "/")
}

func setFrontendSecurityHeaders(c *gin.Context) {
	c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' http: https: data:; font-src 'self' data:; connect-src 'self' http: https:; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Frame-Options", "DENY")
	c.Header("Referrer-Policy", "no-referrer")
}

func serveFrontendRequestPath(c *gin.Context, fileServer http.Handler, requestPath string) {
	originalPath := c.Request.URL.Path
	c.Request.URL.Path = requestPath
	fileServer.ServeHTTP(c.Writer, c.Request)
	c.Request.URL.Path = originalPath
}

func frontendFileExists(frontendFS fs.FS, filePath string) bool {
	info, err := fs.Stat(frontendFS, filePath)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

// isBackendPath reports paths owned by the backend. A GET under one of them
// must 404 instead of falling through to the SPA, including the compatibility
// APIs registered outside /api.
func isBackendPath(requestPath string) bool {
	cleanedPath := path.Clean(requestPath)
	if cleanedPath == "." {
		cleanedPath = "/"
	}

	for _, prefix := range []string{"/api", "/reader", "/accounts"} {
		if cleanedPath == prefix || strings.HasPrefix(cleanedPath, prefix+"/") {
			return true
		}
	}
	return false
}

// isGReaderAPIPath reports paths under the Google Reader API prefix, whose
// unregistered endpoints get the empty-list fallback instead of a 404.
func isGReaderAPIPath(requestPath string) bool {
	cleanedPath := path.Clean(requestPath)
	return cleanedPath == "/reader/api/0" || strings.HasPrefix(cleanedPath, "/reader/api/0/")
}

func looksLikeAssetPath(filePath string) bool {
	base := path.Base(filePath)
	return strings.Contains(base, ".")
}
