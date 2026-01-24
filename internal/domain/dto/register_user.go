package dto

import "github.com/rockkley/pushpost/pkg/validator"

type RegisterUserDto struct {
	Username string
	Email    string
	Password string
}

func (dto *RegisterUserDto) Validate() bool {
	err := validator.IsValidEmail(dto.Email)
	err = validator.IsValidUsername(dto.Username)
	err = validator.IsValidPassword(dto.Password)
	if err != false {
		return err
	}
	return true
}
