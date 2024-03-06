package api

import (
	"net/http"

	"github.com/0x2e/fusion/server"

	"github.com/labstack/echo/v4"
)

type itemAPI struct {
	srv *server.Item
}

func newItemAPI(srv *server.Item) *itemAPI {
	return &itemAPI{
		srv: srv,
	}
}

func (i itemAPI) List(c echo.Context) error {
	var req server.ReqItemList
	if err := bindAndValidate(&req, c); err != nil {
		return err
	}

	resp, err := i.srv.List(&req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

func (i itemAPI) Get(c echo.Context) error {
	var req server.ReqItemGet
	if err := bindAndValidate(&req, c); err != nil {
		return err
	}

	resp, err := i.srv.Get(&req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

func (i itemAPI) Update(c echo.Context) error {
	var req server.ReqItemUpdate
	if err := bindAndValidate(&req, c); err != nil {
		return err
	}

	if err := i.srv.Update(&req); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (i itemAPI) Delete(c echo.Context) error {
	var req server.ReqItemDelete
	if err := bindAndValidate(&req, c); err != nil {
		return err
	}

	if err := i.srv.Delete(&req); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
