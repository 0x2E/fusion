package api

import (
	"net/http"

	"github.com/0x2e/fusion/conf"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type Session struct{}

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

	if req.Password != conf.Conf.Password {
		return echo.NewHTTPError(http.StatusUnauthorized, "Wrong password")
	}

	sess, _ := session.Get(sessionKeyName, c)

	if !conf.Conf.SecureCookie {
		sess.Options.Secure = false
		sess.Options.SameSite = http.SameSiteDefaultMode
	}

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusCreated)
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
