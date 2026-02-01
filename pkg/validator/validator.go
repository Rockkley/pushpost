package validator

import (
	"github.com/rockkley/pushpost/internal/handler/http/dto"
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
	ErrUsernameTooShort     = "username is too short"
	ErrUsernameTooLong      = "username is too long"
	ErrUsernameInvalidChars = "username contains invalid characters"

	ErrEmailInvalid  = "email is invalid"
	ErrEmailRequired = "email is required"

	ErrPasswordTooLong  = "password is too long"
	ErrPasswordTooShort = "password is too short"
	ErrPasswordWeak     = "password is too weak"

	ErrContentRequired = "content is required"
	ErrContentTooLong  = "content is too long"
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

func ValidateRegisterUser(dto dto.RegisterUserDto) []*ValidationError {
	var errs []*ValidationError

	if err := ValidateUsername(dto.Username); err != nil {
		errs = append(errs, err)
	}

	if err := ValidateEmail(dto.Email); err != nil {
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

func ValidateEmail(email string) *ValidationError {
	email = strings.TrimSpace(email)

	if email == "" {

		return newError(ErrEmailRequired, "email")
	}

	email = strings.ToLower(email)

	if !emailRegex.MatchString(email) {

		return newError(ErrEmailInvalid, "email")
	}

	return nil
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
