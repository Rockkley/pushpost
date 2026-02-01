package dto

import "github.com/google/uuid"

type LoginUserDTO struct {
	Email    string    `json:"email"`
	Password string    `json:"password"`
	DeviceID uuid.UUID `json:"deviceID"`
}
