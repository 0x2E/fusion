package api

import (
	"errors"
	"net/http"

	"github.com/0x2e/fusion/auth"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type Session struct {
	PasswordHash    auth.HashedPassword
	UseSecureCookie bool
}

// sessionKeyName is the name of the key in the session store, and it's also the
// client-visible name of the HTTP cookie for the session.
const sessionKeyName = "session-token"

func (s Session) Create(c echo.Context) error {
	var req struct {
		Password string `json:"password" validate:"required"`
	}

	if err := bindAndValidate(&req, c); err != nil {
		return err
	}

	attemptedPasswordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid password")
	}

	if correctPasswordHash := s.PasswordHash; !attemptedPasswordHash.Equals(correctPasswordHash) {
		return echo.NewHTTPError(http.StatusUnauthorized, "Wrong password")
	}

	sess, err := session.Get(sessionKeyName, c)
	if err != nil {
		return err
	}

	if !s.UseSecureCookie {
		sess.Options.Secure = false
		sess.Options.SameSite = http.SameSiteDefaultMode
	}

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusCreated)
}

func (s Session) Check(c echo.Context) error {
	sess, err := session.Get(sessionKeyName, c)
	if err != nil {
		// If the session token is invalid, advise the client browser to delete the
		// session token cookie.
		sess.Options.MaxAge = -1
		// Deliberately swallow the error because we're already returning a more
		// important error.
		sess.Save(c.Request(), c.Response())
		return err
	}

	// If IsNew is true, it means that Get created a new session on-demand rather
	// than retrieving a previously authenticated session.
	if sess.IsNew {
		return errors.New("invalid session")
	}

	return nil
}

func (s Session) Delete(c echo.Context) error {
	sess, err := session.Get(sessionKeyName, c)
	if err != nil {
		return err
	}
	sess.Options.MaxAge = -1
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}
