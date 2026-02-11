package store

import "errors"

var (
	ErrNotFound = errors.New("not found")
	ErrInvalid  = errors.New("invalid")
)
