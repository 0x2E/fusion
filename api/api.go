package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/0x2e/fusion/conf"
	"github.com/0x2e/fusion/frontend"
	"github.com/0x2e/fusion/pkg/errorx"
	"github.com/0x2e/fusion/repo"
	"github.com/0x2e/fusion/server"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Run() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	r := echo.New()

	if conf.Debug {
		r.Debug = true
		r.Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
			if len(resBody) > 500 {
				resBody = append(resBody[:500], []byte("...")...)
			}
			r.Logger.Debugf("req: %s\nresp: %s\n", reqBody, resBody)
		}))
	}

	r.HideBanner = true
	r.HTTPErrorHandler = errorHandler
	r.Validator = newCustomValidator()
	r.Use(middleware.Recover())
	r.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if !strings.HasPrefix(v.URI, "/api") {
				return nil
			}
			if v.Error == nil {
				logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
				)
			} else {
				logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
				)
			}
			return nil
		},
	}))
	r.Use(session.Middleware(sessions.NewCookieStore([]byte("fusion"))))
	r.Pre(middleware.RemoveTrailingSlash())
	r.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		HTML5:      true,
		Index:      "index.html",
		Filesystem: http.FS(frontend.Content),
		Browse:     false,
	}))

	loginAPI := Session{}
	r.POST("/api/sessions", loginAPI.Create)

	authed := r.Group("/api", func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ok, err := loginAPI.Check(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}
			return next(c)
		}
	})

	feeds := authed.Group("/feeds")
	feedAPIHandler := newFeedAPI(server.NewFeed(repo.NewFeed(repo.DB), repo.NewItem(repo.DB)))
	feeds.GET("", feedAPIHandler.All)
	feeds.GET("/:id", feedAPIHandler.Get)
	feeds.POST("", feedAPIHandler.Create)
	feeds.POST("/validation", feedAPIHandler.CheckValidity)
	feeds.PATCH("/:id", feedAPIHandler.Update)
	feeds.DELETE("/:id", feedAPIHandler.Delete)
	feeds.POST("/refresh", feedAPIHandler.Refresh)

	groups := authed.Group("/groups")
	groupAPIHandler := newGroupAPI(server.NewGroup(repo.NewGroup(repo.DB), repo.NewFeed(repo.DB)))
	groups.GET("", groupAPIHandler.All)
	groups.POST("", groupAPIHandler.Create)
	groups.PATCH("/:id", groupAPIHandler.Update)
	groups.DELETE("/:id", groupAPIHandler.Delete)

	items := authed.Group("/items")
	itemAPIHandler := newItemAPI(server.NewItem(repo.NewItem(repo.DB)))
	items.GET("", itemAPIHandler.List)
	items.GET("/:id", itemAPIHandler.Get)
	items.PATCH("/:id", itemAPIHandler.Update)
	items.DELETE("/:id", itemAPIHandler.Delete)

	r.Logger.Fatal(r.Start(fmt.Sprintf("%s:%d", conf.Conf.Host, conf.Conf.Port)))
}

func errorHandler(err error, c echo.Context) {
	if err == errorx.ErrNotFound {
		err = echo.NewHTTPError(http.StatusNotFound, "Resource does not exists")
	}

	c.Echo().DefaultHTTPErrorHandler(err, c)
}

type CustomValidator struct {
	handler *validator.Validate
	trans   ut.Translator
}

func newCustomValidator() *CustomValidator {
	en := en.New()
	uni := ut.New(en, en)
	trans, _ := uni.GetTranslator("en")
	validate := validator.New()
	en_translations.RegisterDefaultTranslations(validate, trans)
	return &CustomValidator{
		handler: validate,
		trans:   trans,
	}
}

func (v *CustomValidator) Validate(i interface{}) error {
	err := v.handler.Struct(i)
	if err != nil {
		errs := err.(validator.ValidationErrors)
		msg := strings.Builder{}
		for _, content := range errs.Translate(v.trans) {
			msg.WriteString(content)
			msg.WriteString(".")
		}
		err = echo.NewHTTPError(http.StatusBadRequest, msg.String())
	}
	return err
}

func bindAndValidate(i interface{}, c echo.Context) error {
	if err := c.Bind(i); err != nil {
		return err
	}
	return c.Validate(i)
}
