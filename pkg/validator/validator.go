package validator

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	ErrUsernameTooShort     = "user.username.too_short"
	ErrUsernameTooLong      = "user.username.too_long"
	ErrUsernameInvalidChars = "user.username.invalid_chars"

	ErrEmailInvalid  = "user.email.invalid"
	ErrEmailRequired = "user.email.required"

	ErrPasswordTooShort = "user.password.too_short"
	ErrPasswordWeak     = "user.password.weak"

	ErrContentRequired = "content.required"
	ErrContentTooLong  = "content.too_long"
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

func ValidateRegisterUser(username, email, password string) []*ValidationError {
	var errs []*ValidationError

	if err := ValidateUsername(username); err != nil {
		errs = append(errs, err)
	}

	if err := ValidateEmail(email); err != nil {
		errs = append(errs, err)
	}

	if err := ValidatePassword(password); err != nil {
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

	if utf8.RuneCountInString(password) < 8 {
		return newError(ErrPasswordTooShort, "password")
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
