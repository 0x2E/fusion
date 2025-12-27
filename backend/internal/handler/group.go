package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type groupRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *Handler) listGroups(c *gin.Context) {
	groups, err := h.store.ListGroups()
	if err != nil {
		errorResponse(c, 500, err.Error())
		return
	}

	listResponse(c, groups, len(groups))
}

func (h *Handler) getGroup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errorResponse(c, 400, "invalid id")
		return
	}

	group, err := h.store.GetGroup(id)
	if err != nil {
		errorResponse(c, 404, "group not found")
		return
	}

	dataResponse(c, group)
}

func (h *Handler) createGroup(c *gin.Context) {
	var req groupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, 400, "invalid request")
		return
	}

	group, err := h.store.CreateGroup(req.Name)
	if err != nil {
		errorResponse(c, 500, err.Error())
		return
	}

	dataResponse(c, group)
}

func (h *Handler) updateGroup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errorResponse(c, 400, "invalid id")
		return
	}

	var req groupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, 400, "invalid request")
		return
	}

	if err := h.store.UpdateGroup(id, req.Name); err != nil {
		errorResponse(c, 500, err.Error())
		return
	}

	group, err := h.store.GetGroup(id)
	if err != nil {
		errorResponse(c, 500, err.Error())
		return
	}

	dataResponse(c, group)
}

func (h *Handler) deleteGroup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errorResponse(c, 400, "invalid id")
		return
	}

	if err := h.store.DeleteGroup(id); err != nil {
		errorResponse(c, 500, err.Error())
		return
	}

	dataResponse(c, gin.H{"message": "group deleted"})
}
