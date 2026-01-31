package domain

import (
	"fmt"
	"net/http"
)

type DomainError interface {
	error
	HTTPStatus() int
	Code() string
	GetField() string
}

type Field string

const (
	FieldEmail    Field = "email"
	FieldUsername Field = "username"
	FieldPhone    Field = "phone"
)

type FieldValueAlreadyExistsError struct {
	DomainError
	Field Field
}

func (e *FieldValueAlreadyExistsError) Error() string {
	return fmt.Sprintf("%s is already exists in database", e.Field)
}
func (e *FieldValueAlreadyExistsError) HTTPStatus() int  { return http.StatusConflict }
func (e *FieldValueAlreadyExistsError) Code() string     { return "already_exists" }
func (e *FieldValueAlreadyExistsError) GetField() string { return string(e.Field) }

type InvalidCredentialsError struct {
	DomainError
}

func (e *InvalidCredentialsError) Error() string   { return "invalid credentials" }
func (e *InvalidCredentialsError) HTTPStatus() int { return http.StatusUnauthorized }
func (e *InvalidCredentialsError) Code() string    { return "invalid_credentials" }
func (e *InvalidCredentialsError) Field() string   { return "" }

type UnauthorizedError struct{}

func (e *UnauthorizedError) Error() string   { return "unauthorized" }
func (e *UnauthorizedError) HTTPStatus() int { return http.StatusUnauthorized }
func (e *UnauthorizedError) Code() string    { return "unauthorized" }
func (e *UnauthorizedError) Field() string   { return "" }
