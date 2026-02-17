package dto

import "errors"

var (
	ErrInvalidUsername = errors.New("username is required")
	ErrInvalidEmail    = errors.New("email is required")
	ErrInvalidPassword = errors.New("password is required")
)

type CreateUserDTO struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"passwordHash"`
}

func (dto *CreateUserDTO) Validate() error {
	if dto.Username == "" {
		return ErrInvalidUsername
	}
	if dto.Email == "" {
		return ErrInvalidEmail
	}
	if dto.PasswordHash == "" {
		return ErrInvalidPassword
	}
	return nil
}
