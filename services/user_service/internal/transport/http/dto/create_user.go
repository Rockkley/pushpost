package dto

import "errors"

type CreateUserRequestDTO struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"passwordHash"`
}

var (
	ErrUsernameRequired = errors.New("username is required")
	ErrEmailRequired    = errors.New("email is required")
	ErrPasswordRequired = errors.New("password is required")
)

func (dto *CreateUserRequestDTO) Validate() error {
	if dto.Username == "" {
		return ErrUsernameRequired
	}
	if dto.Email == "" {
		return ErrEmailRequired
	}
	if dto.PasswordHash == "" {
		return ErrPasswordRequired
	}
	return nil
}
