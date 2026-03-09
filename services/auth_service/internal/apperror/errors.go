package apperror

import "github.com/rockkley/pushpost/services/common_service/apperror"

func InvalidCredentials() apperror.AppError {
	return apperror.Unauthorized(CodeInvalidCredentials, "invalid credentials")
}

func SessionExpired() apperror.AppError {
	return apperror.Unauthorized(CodeSessionExpired, "session expired")
}

func AccountDeleted() apperror.AppError {
	return apperror.Forbidden(CodeAccountDeleted, "account has been deleted")
}
