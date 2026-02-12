package user_api

import "errors"

var (
	ErrNotFound      = errors.New("user not found")
	ErrBadRequest    = errors.New("bad request")
	ErrInternalError = errors.New("internal error")
)
