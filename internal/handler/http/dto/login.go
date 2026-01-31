package dto

type LoginUserDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	DeviceID string `json:"deviceID"`
}
