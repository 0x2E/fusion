package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

// internalError logs the actual error and returns a generic message to client.
func internalError(c *gin.Context, err error, context string) {
	slog.Error("internal error",
		"context", context,
		"error", err,
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
	)
	c.JSON(500, gin.H{"error": "internal server error"})
}

// notFoundError returns 404 with a consistent message.
func notFoundError(c *gin.Context, resource string) {
	c.JSON(404, gin.H{"error": resource + " not found"})
}

// badRequestError returns 400 with the given message.
func badRequestError(c *gin.Context, message string) {
	c.JSON(400, gin.H{"error": message})
}

// unauthorizedError returns 401.
func unauthorizedError(c *gin.Context) {
	c.JSON(401, gin.H{"error": "unauthorized"})
}
