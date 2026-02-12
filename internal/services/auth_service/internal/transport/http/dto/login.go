package dto

import "github.com/google/uuid"

type LoginUserDTO struct {
	Email    string    `json:"email"`
	Password string    `json:"password"`
	DeviceID uuid.UUID `json:"deviceID"`
}

func (dto *LoginUserDTO) Validate() error {
	if dto.Email == "" {
		return ErrEmailRequired
	}
	if dto.Password == "" {
		return ErrPasswordRequired
	}
	if dto.DeviceID == uuid.Nil {
		return ErrDeviceIDRequired
	}
	return nil
}
