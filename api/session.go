package api

import (
	"net/http"

	"github.com/0x2e/fusion/conf"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type Session struct{}

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

	sess, _ := session.Get("login", c)

	if !conf.Conf.SecureCookie {
		sess.Options.Secure = false
		sess.Options.SameSite = http.SameSiteDefaultMode
	}

	sess.Values["password"] = conf.Conf.Password
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusCreated)
}

func (s Session) Check(c echo.Context) (bool, error) {
	sess, err := session.Get("login", c)
	if err != nil {
		return false, err
	}
	v, ok := sess.Values["password"]
	if !ok {
		return false, nil
	}
	return v == conf.Conf.Password, nil
}
