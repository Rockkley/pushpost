package dto

import (
	"errors"
	"github.com/google/uuid"
)

type RegisterUserDTO struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponseDTO struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}

func (dto *RegisterUserDTO) Validate() error {
	if dto.Username == "" {
		return errors.New("username is required")
	}

	if dto.Email == "" {
		return errors.New("email is required")
	}

	if dto.Password == "" {
		return errors.New("password is required")
	}

	return nil
}
