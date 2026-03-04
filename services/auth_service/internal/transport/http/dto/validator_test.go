package dto

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// ── ValidateUsername ──────────────────────────────────────────────────────────

func TestValidateUsername_Valid(t *testing.T) {
	valid := []string{
		"abc",
		"ABC",
		"username",
		"user_name",
		"user123",
		"U1_",
		strings.Repeat("a", 30), // exactly max length
		"aaa",                   // exactly min length
	}

	for _, u := range valid {
		t.Run(u, func(t *testing.T) {
			require.Nil(t, ValidateUsername(u))
		})
	}
}

func TestValidateUsername_TooShort(t *testing.T) {
	cases := []string{"", "a", "ab"}
	for _, u := range cases {
		t.Run(fmt.Sprintf("len=%d", len(u)), func(t *testing.T) {
			err := ValidateUsername(u)
			require.NotNil(t, err)
			require.Equal(t, "field_too_short", err.Code)
			require.Equal(t, "username", err.Field)
		})
	}
}

func TestValidateUsername_TooLong(t *testing.T) {
	err := ValidateUsername(strings.Repeat("a", 31))
	require.NotNil(t, err)
	require.Equal(t, "field_too_long", err.Code)
	require.Equal(t, "username", err.Field)
}

func TestValidateUsername_InvalidCharacters(t *testing.T) {
	invalid := []string{
		"user name",    // space
		"user@name",    // @
		"user-name",    // hyphen
		"user.name",    // dot
		"user!",        // exclamation
		"пользователь", // cyrillic
		"用户名",          // Chinese
		"user😀",        // emoji
		"user\tname",   // tab
		"user\nname",   // newline
	}

	for _, u := range invalid {
		t.Run(u, func(t *testing.T) {
			err := ValidateUsername(u)
			require.NotNil(t, err)
			require.Equal(t, "field_invalid", err.Code)
			require.Equal(t, "username", err.Field)
		})
	}
}

func TestValidateUsername_BoundaryLengths(t *testing.T) {
	// Exactly 3 chars → valid.
	require.Nil(t, ValidateUsername("abc"))
	// Exactly 30 chars → valid.
	require.Nil(t, ValidateUsername(strings.Repeat("a", 30)))
	// 31 chars → too long.
	require.NotNil(t, ValidateUsername(strings.Repeat("a", 31)))
}

// ── ValidateEmail ─────────────────────────────────────────────────────────────

func TestValidateEmail_Valid(t *testing.T) {
	valid := []string{
		"user@example.com",
		"USER@EXAMPLE.COM",
		"user+tag@example.co.uk",
		"user.name@sub.domain.org",
		"u@e.io",
		"  user@example.com  ", // trimmed by validator
	}

	for _, e := range valid {
		t.Run(e, func(t *testing.T) {
			ok, err := ValidateEmail(e)
			require.True(t, ok)
			require.Nil(t, err)
		})
	}
}

func TestValidateEmail_Empty(t *testing.T) {
	ok, err := ValidateEmail("")
	require.False(t, ok)
	require.NotNil(t, err)
	require.Equal(t, "field_required", err.Code)
	require.Equal(t, "email", err.Field)
}

func TestValidateEmail_OnlySpaces(t *testing.T) {
	ok, err := ValidateEmail("   ")
	require.False(t, ok)
	require.NotNil(t, err)
	require.Equal(t, "field_required", err.Code)
}

func TestValidateEmail_InvalidFormats(t *testing.T) {
	invalid := []string{
		"notanemail",
		"@nodomain",
		"user@",
		"user@domain",      // missing TLD
		"user @domain.com", // space before @
		"user@ domain.com", // space after @
		"user@@domain.com", // double @
		"user@domain..com", // double dot
	}

	for _, e := range invalid {
		t.Run(e, func(t *testing.T) {
			ok, err := ValidateEmail(e)
			require.False(t, ok)
			require.NotNil(t, err)
			require.Equal(t, "field_invalid", err.Code)
			require.Equal(t, "email", err.Field)
		})
	}
}

// ── ValidatePassword ──────────────────────────────────────────────────────────

func TestValidatePassword_Valid(t *testing.T) {
	valid := []string{
		"Password1",
		"P4ssword",
		"12345678A",
		"abcABC123",
		strings.Repeat("Aa1", 3), // 9 chars — above min
	}

	for _, p := range valid {
		t.Run(p, func(t *testing.T) {
			require.Nil(t, ValidatePassword(p))
		})
	}
}

func TestValidatePassword_TooShort(t *testing.T) {
	cases := []string{"", "A1", "Abc123", "Pass1"} // all < 8 chars
	for _, p := range cases {
		t.Run(fmt.Sprintf("len=%d", len(p)), func(t *testing.T) {
			err := ValidatePassword(p)
			require.NotNil(t, err)
			require.Equal(t, "field_too_short", err.Code)
			require.Equal(t, "password", err.Field)
		})
	}
}

func TestValidatePassword_ExactlyMinLength(t *testing.T) {
	// 8 chars with letters + digits = valid.
	require.Nil(t, ValidatePassword("Pass1234"))
}

func TestValidatePassword_TooLong(t *testing.T) {
	// 129 chars (MaxPasswordLength + 1)
	long := strings.Repeat("A1", 65) // 130 chars
	err := ValidatePassword(long)
	require.NotNil(t, err)
	require.Equal(t, "field_too_long", err.Code)
	require.Equal(t, "password", err.Field)
}

func TestValidatePassword_ExactlyMaxLength(t *testing.T) {
	// 128 chars — right at the limit.
	max := strings.Repeat("A", 64) + strings.Repeat("1", 64) // 128 chars
	require.Nil(t, ValidatePassword(max))
}

func TestValidatePassword_NoDigits(t *testing.T) {
	cases := []string{
		"OnlyLetters",
		"ALLCAPS!!?",
		"mixedCase!!",
	}
	for _, p := range cases {
		t.Run(p, func(t *testing.T) {
			err := ValidatePassword(p)
			require.NotNil(t, err)
			require.Equal(t, "field_weak", err.Code)
			require.Equal(t, "password", err.Field)
		})
	}
}

func TestValidatePassword_NoLetters(t *testing.T) {
	cases := []string{
		"12345678",
		"11223344",
		"99999999",
	}
	for _, p := range cases {
		t.Run(p, func(t *testing.T) {
			err := ValidatePassword(p)
			require.NotNil(t, err)
			require.Equal(t, "field_weak", err.Code)
			require.Equal(t, "password", err.Field)
		})
	}
}

// ── ValidateRegisterUser ──────────────────────────────────────────────────────

func TestValidateRegisterUser_AllValid(t *testing.T) {
	req := RegisterUserDto{
		Username: "validuser",
		Email:    "user@example.com",
		Password: "Password1",
	}

	errs := ValidateRegisterUser(req)
	require.Empty(t, errs)
}

func TestValidateRegisterUser_CollectsAllErrors(t *testing.T) {
	// All three fields invalid → all three errors collected in one pass.
	req := RegisterUserDto{
		Username: "ab",    // too short
		Email:    "",      // required
		Password: "short", // too short
	}

	errs := ValidateRegisterUser(req)
	require.Len(t, errs, 3)

	byField := make(map[string]string)
	for _, e := range errs {
		byField[e.Field] = e.Code
	}

	require.Equal(t, "field_too_short", byField["username"])
	require.Equal(t, "field_required", byField["email"])
	require.Equal(t, "field_too_short", byField["password"])
}

func TestValidateRegisterUser_SingleFieldError(t *testing.T) {
	req := RegisterUserDto{
		Username: "validuser",
		Email:    "notanemail",
		Password: "Password1",
	}

	errs := ValidateRegisterUser(req)
	require.Len(t, errs, 1)
	require.Equal(t, "email", errs[0].Field)
	require.Equal(t, "field_invalid", errs[0].Code)
}

func TestValidateRegisterUser_WeakPassword(t *testing.T) {
	req := RegisterUserDto{
		Username: "validuser",
		Email:    "user@example.com",
		Password: "onlyletters",
	}

	errs := ValidateRegisterUser(req)
	require.Len(t, errs, 1)
	require.Equal(t, "password", errs[0].Field)
	require.Equal(t, "field_weak", errs[0].Code)
}

func TestValidateRegisterUser_ErrorsAreUnique(t *testing.T) {
	// Each field must appear at most once in the result.
	req := RegisterUserDto{
		Username: "ab",
		Email:    "bad",
		Password: "bad",
	}

	errs := ValidateRegisterUser(req)
	seen := make(map[string]int)
	for _, e := range errs {
		seen[e.Field]++
	}
	for field, count := range seen {
		require.Equal(t, 1, count, "field %q appeared %d times", field, count)
	}
}
