package dto

import "errors"

type AuthenticateUserRequestDTO struct {
	Email        string `json:"email"`
	PasswordHash string `json:"passwordHash"`
}

func (dto *AuthenticateUserRequestDTO) Validate() error {
	if dto.Email == "" {
		return errors.New("email is required")
	}
	if dto.PasswordHash == "" {
		return errors.New("passwordHash is required")
	}
	return nil
}
