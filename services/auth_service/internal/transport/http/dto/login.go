package dto

import (
	"errors"
	"github.com/google/uuid"
)

type LoginUserDTO struct {
	Email    string    `json:"email"`
	Password string    `json:"password"`
	DeviceID uuid.UUID `json:"deviceID"`
}

func (dto *LoginUserDTO) Validate() error {
	if dto.Email == "" {
		return errors.New(ErrEmailRequired)
	}
	if dto.Password == "" {
		return errors.New("password is required")
	}
	if dto.DeviceID == uuid.Nil {
		return errors.New("device ID is required")
	}
	return nil
}
