package handler

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func requestLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestPath := c.Request.URL.Path
		if rawQuery := strings.TrimSpace(c.Request.URL.RawQuery); rawQuery != "" {
			requestPath += "?" + rawQuery
		}

		c.Next()

		status := c.Writer.Status()
		if status == 0 {
			status = http.StatusOK
		}

		attrs := []any{
			"method", c.Request.Method,
			"path", requestPath,
			"status", status,
			"latency", time.Since(start),
			"client_ip", c.ClientIP(),
		}

		if route := c.FullPath(); route != "" {
			attrs = append(attrs, "route", route)
		}
		if size := c.Writer.Size(); size >= 0 {
			attrs = append(attrs, "bytes", size)
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, "errors", strings.TrimSpace(c.Errors.String()))
		}

		slog.Log(c.Request.Context(), requestLogLevel(status), "http request", attrs...)
	}
}

func recoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.ErrorContext(
					c.Request.Context(),
					"panic recovered",
					"panic", recovered,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"client_ip", c.ClientIP(),
					"stack", string(debug.Stack()),
				)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()

		c.Next()
	}
}

func requestLogLevel(status int) slog.Level {
	if status >= http.StatusInternalServerError {
		return slog.LevelError
	}
	if status >= http.StatusBadRequest {
		return slog.LevelWarn
	}

	return slog.LevelInfo
}
