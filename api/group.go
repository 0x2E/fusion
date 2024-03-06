package api

import (
	"net/http"

	"github.com/0x2e/fusion/server"

	"github.com/labstack/echo/v4"
)

type groupAPI struct {
	srv *server.Group
}

func newGroupAPI(srv *server.Group) *groupAPI {
	return &groupAPI{
		srv: srv,
	}
}

func (f groupAPI) All(c echo.Context) error {
	resp, err := f.srv.All()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

func (f groupAPI) Create(c echo.Context) error {
	var req server.ReqGroupCreate
	if err := bindAndValidate(&req, c); err != nil {
		return err
	}

	if err := f.srv.Create(&req); err != nil {
		return err
	}

	return c.NoContent(http.StatusCreated)
}

func (f groupAPI) Update(c echo.Context) error {
	var req server.ReqGroupUpdate
	if err := bindAndValidate(&req, c); err != nil {
		return err
	}

	err := f.srv.Update(&req)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (f groupAPI) Delete(c echo.Context) error {
	var req server.ReqGroupDelete
	if err := bindAndValidate(&req, c); err != nil {
		return err
	}

	if err := f.srv.Delete(&req); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
