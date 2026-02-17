package apperror

import (
	"errors"
	"fmt"
)

type AppError interface {
	error
	Type() ErrorType
	HTTPStatus() int
	Code() string
	Field() string
	Fields() map[string]string
	Unwrap() error
}

type ErrorType int

const (
	ErrorTypeClient ErrorType = iota
	ErrorTypeServer
	ErrorTypeValidation
)

type appError struct {
	errType    ErrorType
	httpStatus int
	code       string
	field      string
	fields     map[string]string
	message    string
	cause      error
}

func (e *appError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

func (e *appError) Type() ErrorType { return e.errType }

func (e *appError) HTTPStatus() int { return e.httpStatus }

func (e *appError) Code() string { return e.code }

func (e *appError) Field() string { return e.field }
func (e *appError) Unwrap() error { return e.cause }

func (e *appError) Fields() map[string]string { return e.fields }

func IsClientError(err error) bool {
	var appErr AppError
	if errors.As(err, &appErr) {
		return appErr.Type() == ErrorTypeClient || appErr.Type() == ErrorTypeValidation
	}
	return false
}

func IsServerError(err error) bool {
	var appErr AppError
	if errors.As(err, &appErr) {
		return appErr.Type() == ErrorTypeServer
	}
	return false
}
