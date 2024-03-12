package repo

import "errors"

var (
	ErrNotFound      = errors.New("resource not exists")
	ErrDuplicatedKey = errors.New("exists duplicated key(s)")
)
