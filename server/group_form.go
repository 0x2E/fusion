package server

type GroupForm struct {
	ID   uint    `json:"id"`
	Name *string `json:"name"`
}

type RespGroupAll struct {
	Groups []*GroupForm `json:"groups"`
}

type ReqGroupCreate struct {
	Name *string `json:"name" validate:"required"`
}

type ReqGroupUpdate struct {
	ID   uint    `param:"id" validate:"required"`
	Name *string `json:"name" validate:"required"`
}

type ReqGroupDelete struct {
	ID uint `param:"id" validate:"required"`
}
