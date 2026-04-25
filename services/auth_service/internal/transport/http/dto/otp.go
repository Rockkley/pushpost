package dto

import (
	"errors"
	"strings"
)

type VerifyEmailRequestDTO struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func (d *VerifyEmailRequestDTO) Validate() error {
	d.Email = strings.ToLower(strings.TrimSpace(d.Email))
	d.Code = strings.TrimSpace(d.Code)

	if d.Email == "" {
		return errors.New("email is required")
	}

	if d.Code == "" {
		return errors.New("code is required")
	}

	return nil
}

type ResendOTPRequestDTO struct {
	Email string `json:"email"`
}

func (d *ResendOTPRequestDTO) Validate() error {
	d.Email = strings.ToLower(strings.TrimSpace(d.Email))

	if d.Email == "" {
		return errors.New("email is required")
	}

	return nil
}
