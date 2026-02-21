package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/patrickjmcd/reedme/internal/store"
)

type groupRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *Handler) listGroups(c *gin.Context) {
	groups, err := h.store.ListGroups()
	if err != nil {
		internalError(c, err, "list groups")
		return
	}

	listResponse(c, groups, len(groups))
}

func (h *Handler) getGroup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestError(c, "invalid id")
		return
	}

	group, err := h.store.GetGroup(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFoundError(c, "group")
			return
		}
		internalError(c, err, "get group")
		return
	}

	dataResponse(c, group)
}

func (h *Handler) createGroup(c *gin.Context) {
	var req groupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestError(c, "invalid request")
		return
	}

	group, err := h.store.CreateGroup(req.Name)
	if err != nil {
		internalError(c, err, "create group")
		return
	}

	dataResponse(c, group)
}

func (h *Handler) updateGroup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestError(c, "invalid id")
		return
	}

	var req groupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestError(c, "invalid request")
		return
	}

	if err := h.store.UpdateGroup(id, req.Name); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFoundError(c, "group")
			return
		}
		internalError(c, err, "update group")
		return
	}

	group, err := h.store.GetGroup(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFoundError(c, "group")
			return
		}
		internalError(c, err, "get group after update")
		return
	}

	dataResponse(c, group)
}

func (h *Handler) deleteGroup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestError(c, "invalid id")
		return
	}

	if err := h.store.DeleteGroup(id); err != nil {
		if errors.Is(err, store.ErrInvalid) {
			badRequestError(c, err.Error())
			return
		}
		if errors.Is(err, store.ErrNotFound) {
			notFoundError(c, "group")
			return
		}
		internalError(c, err, "delete group")
		return
	}

	c.Status(http.StatusNoContent)
}
