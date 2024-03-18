package api

import (
	"net/http"

	"github.com/0x2e/fusion/server"

	"github.com/labstack/echo/v4"
)

type feedAPI struct {
	srv *server.Feed
}

func newFeedAPI(srv *server.Feed) *feedAPI {
	return &feedAPI{
		srv: srv,
	}
}

func (f feedAPI) All(c echo.Context) error {
	resp, err := f.srv.All(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

func (f feedAPI) Get(c echo.Context) error {
	var req server.ReqFeedGet
	if err := bindAndValidate(&req, c); err != nil {
		return err
	}

	resp, err := f.srv.Get(c.Request().Context(), &req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

func (f feedAPI) Create(c echo.Context) error {
	var req server.ReqFeedCreate
	if err := bindAndValidate(&req, c); err != nil {
		return err
	}

	if err := f.srv.Create(c.Request().Context(), &req); err != nil {
		return err
	}

	return c.NoContent(http.StatusCreated)
}

func (f feedAPI) CheckValidity(c echo.Context) error {
	var req server.ReqFeedCheckValidity
	if err := bindAndValidate(&req, c); err != nil {
		return err
	}

	resp, err := f.srv.CheckValidity(c.Request().Context(), &req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, resp)
}

func (f feedAPI) Update(c echo.Context) error {
	var req server.ReqFeedUpdate
	if err := bindAndValidate(&req, c); err != nil {
		return err
	}

	err := f.srv.Update(c.Request().Context(), &req)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (f feedAPI) Delete(c echo.Context) error {
	var req server.ReqFeedDelete
	if err := bindAndValidate(&req, c); err != nil {
		return err
	}

	if err := f.srv.Delete(c.Request().Context(), &req); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (f feedAPI) Refresh(c echo.Context) error {
	var req server.ReqFeedRefresh
	if err := bindAndValidate(&req, c); err != nil {
		return err
	}

	if err := f.srv.Refresh(c.Request().Context(), &req); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
