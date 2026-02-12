package dto

import "github.com/google/uuid"

type RegisterUserDto struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponseDto struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}

func (dto *RegisterUserDto) Validate() error {
	if dto.Username == "" {
		return ErrUsernameRequired
	}
	if dto.Email == "" {
		return ErrEmailRequired
	}
	if dto.Password == "" {
		return ErrPasswordRequired
	}
	return nil
}
