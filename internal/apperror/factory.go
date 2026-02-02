package apperror

import "net/http"

// Client error codes

func NotFound(code, message string) AppError {
	return &appError{
		errType:    ErrorTypeClient,
		httpStatus: http.StatusNotFound,
		code:       code,
		message:    message,
	}
}

func Conflict(code, field, message string) AppError {
	return &appError{
		errType:    ErrorTypeClient,
		httpStatus: http.StatusConflict,
		code:       code,
		field:      field,
		message:    message,
	}
}

func BadRequest(code, message string) AppError {
	return &appError{
		errType:    ErrorTypeClient,
		httpStatus: http.StatusBadRequest,
		code:       code,
		message:    message,
	}
}

func Forbidden(code, message string) AppError {
	return &appError{
		errType:    ErrorTypeClient,
		httpStatus: http.StatusForbidden,
		code:       code,
		message:    message,
	}
}

func Unauthorized(code, message string) AppError {
	return &appError{
		errType:    ErrorTypeClient,
		httpStatus: http.StatusUnauthorized,
		code:       code,
		message:    message,
	}
}

func Validation(code, field, message string) AppError {
	return &appError{
		errType:    ErrorTypeValidation,
		httpStatus: http.StatusUnprocessableEntity,
		code:       code,
		field:      field,
		message:    message,
	}
}
func ValidationFields(fields map[string]string) AppError {
	return &appError{
		errType:    ErrorTypeValidation,
		httpStatus: http.StatusUnprocessableEntity,
		code:       CodeValidationFailed,
		fields:     fields,
		message:    "validation failed",
	}
}

// Server error codes

func Internal(message string, cause error) AppError {
	return &appError{
		errType:    ErrorTypeServer,
		httpStatus: http.StatusInternalServerError,
		code:       CodeInternalError,
		message:    message,
		cause:      cause,
	}
}

func Database(message string, cause error) AppError {
	return &appError{
		errType:    ErrorTypeServer,
		httpStatus: http.StatusInternalServerError,
		code:       CodeDatabaseError,
		message:    message,
		cause:      cause,
	}
}

func Service(message string, cause error) AppError {
	return &appError{
		errType:    ErrorTypeServer,
		httpStatus: http.StatusInternalServerError,
		code:       CodeServiceError,
		message:    message,
		cause:      cause,
	}
}

// Concrete errors

func UserNotFound() AppError {
	return NotFound(CodeUserNotFound, "user not found")
}

func EmailAlreadyExists() AppError {
	return Conflict(CodeEmailExists, "email", "email already exists")
}

func UsernameAlreadyExists() AppError {
	return Conflict(CodeUsernameExists, "username", "username already exists")
}

func InvalidCredentials() AppError {
	return Unauthorized(CodeInvalidCredentials, "invalid credentials")
}

func SessionExpired() AppError {
	return Unauthorized(CodeSessionExpired, "session expired")
}

func AccountDeleted() AppError {
	return Forbidden(CodeAccountDeleted, "account has been deleted")
}

func CannotMessageSelf() AppError {
	return BadRequest(CodeCannotMessageSelf, "cannot send message to yourself")
}

func MessageNotFound() AppError {
	return NotFound(CodeMessageNotFound, "message not found")
}
