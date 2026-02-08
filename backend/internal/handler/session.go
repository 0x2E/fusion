package handler

import (
	"net/http"

	"github.com/0x2E/fusion/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func isSecureRequest(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return r.Header.Get("X-Forwarded-Proto") == "https"
}

type loginRequest struct {
	Password string `json:"password" binding:"required"`
}

func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestError(c, "invalid request")
		return
	}

	if err := auth.CheckPassword(h.passwordHash, req.Password); err != nil {
		unauthorizedError(c)
		return
	}

	h.createSession(c)
	dataResponse(c, gin.H{"message": "logged in"})
}

// createSession generates a new session ID, stores it, and sets the session cookie.
func (h *Handler) createSession(c *gin.Context) {
	sessionID := uuid.New().String()

	h.mu.Lock()
	h.sessions[sessionID] = true
	h.mu.Unlock()

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   3600 * 24 * 30,
		HttpOnly: true,
		Secure:   isSecureRequest(c.Request),
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) logout(c *gin.Context) {
	sessionID, err := c.Cookie("session")
	if err == nil {
		h.mu.Lock()
		delete(h.sessions, sessionID)
		h.mu.Unlock()
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isSecureRequest(c.Request),
		SameSite: http.SameSiteLaxMode,
	})

	dataResponse(c, gin.H{"message": "logged out"})
}
