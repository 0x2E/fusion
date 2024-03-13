package server

type Paginate struct {
	PageSize int `query:"page_size" validate:"omitnil,min=0"`
	Page     int `query:"page" validate:"omitnil,min=0"`
}
