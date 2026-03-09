package apperror

import (
	"fmt"
)

type AppError interface {
	error
	HTTPStatus() int
	Code() string
	Field() string
	Fields() map[string]string
	Unwrap() error
}

type appError struct {
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

func (e *appError) HTTPStatus() int           { return e.httpStatus }
func (e *appError) Code() string              { return e.code }
func (e *appError) Field() string             { return e.field }
func (e *appError) Unwrap() error             { return e.cause }
func (e *appError) Fields() map[string]string { return e.fields }
