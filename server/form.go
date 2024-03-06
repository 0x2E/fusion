package server

type Paginate struct {
	Count  *int `query:"count" validate:"omitnil,min=0"`
	Offset *int `query:"offset" validate:"omitnil,min=0"`
}
