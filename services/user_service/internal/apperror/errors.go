package apperror

import (
	"github.com/rockkley/pushpost/services/common_service/apperror"
)

func UserNotFound() apperror.AppError {
	return apperror.NotFound(CodeUserNotFound, "user not found")
}

func UserDeleted() apperror.AppError {
	return apperror.Forbidden(CodeUserDeleted, "account has been deleted")
}

func EmailAlreadyExists() apperror.AppError {
	return apperror.Conflict(CodeEmailExists, "email", "email already exists")
}

func UsernameAlreadyExists() apperror.AppError {
	return apperror.Conflict(CodeUsernameExists, "username", "username already exists")
}

func InvalidCredentials() apperror.AppError {
	return apperror.Unauthorized(CodeInvalidCredentials, "invalid credentials")
}

func SessionExpired() apperror.AppError {
	return apperror.Unauthorized(CodeSessionExpired, "session expired")
}

func AccountDeleted() apperror.AppError {
	return apperror.Forbidden(CodeAccountDeleted, "account has been deleted")
}

// -- Postgres constraint mapper

func MapConstraint(constraintName string) apperror.AppError {
	switch constraintName {
	case "users_email_key", "idx_users_email_unique":
		return EmailAlreadyExists()
	case "users_username_key", "idx_users_username_unique":
		return UsernameAlreadyExists()
	default:
		return nil
	}
}
