package server

// BizError is the error allowed to show on frontend user side
type BizError struct {
	HTTPCode  uint
	FEMessage string
	Raw       error
}

func (e BizError) Error() string {
	return e.Raw.Error()
}

func NewBizError(raw error, code uint, message string) BizError {
	return BizError{
		HTTPCode:  code,
		FEMessage: message,
		Raw:       raw,
	}
}
