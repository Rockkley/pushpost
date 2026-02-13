package dto

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	MinPasswordLength = 8
	MaxPasswordLength = 128
)
const (
	ErrUsernameTooShort     = "field_too_short"
	ErrUsernameTooLong      = "field_too_long"
	ErrUsernameInvalidChars = "field_invalid"

	ErrEmailInvalid  = "field_invalid"
	ErrEmailRequired = "field_required"

	ErrPasswordTooLong  = "field_too_long"
	ErrPasswordTooShort = "field_too_short"
	ErrPasswordWeak     = "field_weak"

	ErrContentRequired = "field_required"
	ErrContentTooLong  = "field_too_long"
)

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)
)

type ValidationError struct {
	Code  string
	Field string
}

func (e ValidationError) Error() string { return e.Code }

func newError(code, field string) *ValidationError { return &ValidationError{Code: code, Field: field} }

func ValidateRegisterUser(dto RegisterUserDto) []*ValidationError {
	var errs []*ValidationError

	if err := ValidateUsername(dto.Username); err != nil {
		errs = append(errs, err)
	}

	if _, err := ValidateEmail(dto.Email); err != nil {
		errs = append(errs, err)
	}

	if err := ValidatePassword(dto.Password); err != nil {
		errs = append(errs, err)
	}

	return errs

}

func ValidateUsername(username string) *ValidationError {
	length := utf8.RuneCountInString(username)

	if length < 3 {
		return newError(ErrUsernameTooShort, "username")
	}

	if length > 30 {
		return newError(ErrUsernameTooLong, "username")
	}

	if !usernameRegex.MatchString(username) {
		return newError(ErrUsernameInvalidChars, "username")
	}

	return nil
}

func ValidateEmail(email string) (bool, *ValidationError) {
	email = strings.TrimSpace(email)

	if email == "" {

		return false, newError(ErrEmailRequired, "email")
	}

	email = strings.ToLower(email)

	if !emailRegex.MatchString(email) {

		return false, newError(ErrEmailInvalid, "email")
	}

	return true, nil
}

func ValidatePassword(password string) *ValidationError {

	length := utf8.RuneCountInString(password)
	if length < MinPasswordLength {
		return newError(ErrPasswordTooShort, "password")
	}

	if length > MaxPasswordLength {
		return newError(ErrPasswordTooLong, "password")
	}

	var hasLetters, hasDigits bool

	for _, c := range password {
		switch {
		case unicode.IsDigit(c):
			hasDigits = true
		case unicode.IsLetter(c):
			hasLetters = true
		}

		if hasLetters && hasDigits {

			return nil
		}
	}

	if !hasLetters || !hasDigits {

		return newError(ErrPasswordWeak, "password")
	}

	return nil
}
