package apperror

import "net/http"

const (
	CodeInternalError = "internal_error"
	CodeDatabaseError = "database_error"
	CodeServiceError  = "service_error"

	// validation errors
	CodeValidationFailed = "validation_failed"
	CodeFieldRequired    = "field_required"
	CodeFieldInvalid     = "field_invalid"
	CodeFieldTooShort    = "field_too_short"
	CodeFieldTooLong     = "field_too_long"
	CodeFieldWeak        = "field_weak"

	// generic errors
	CodeAlreadyExists = "already_exists"
	CodeUnauthorized  = "unauthorized"
)

// 4xx ----------------
func BadRequest(code, message string) AppError {
	return &appError{httpStatus: http.StatusBadRequest, code: code, message: message}
}

func Unauthorized(code, message string) AppError {
	return &appError{httpStatus: http.StatusUnauthorized, code: code, message: message}
}

func Forbidden(code, message string) AppError {
	return &appError{httpStatus: http.StatusForbidden, code: code, message: message}
}

func NotFound(code, message string) AppError {
	return &appError{httpStatus: http.StatusNotFound, code: code, message: message}
}

func Conflict(code, field, message string) AppError {
	return &appError{httpStatus: http.StatusConflict, code: code, field: field, message: message}
}

func Validation(code, field, message string) AppError {
	return &appError{
		httpStatus: http.StatusUnprocessableEntity,
		code:       code,
		field:      field,
		message:    message,
	}
}

func ValidationFields(fields map[string]string) AppError {
	return &appError{
		httpStatus: http.StatusUnprocessableEntity,
		code:       CodeValidationFailed,
		fields:     fields,
		message:    "validation failed",
	}
}

// 5xx ----------------

func Internal(message string, cause error) AppError {
	return &appError{
		httpStatus: http.StatusInternalServerError,
		code:       CodeInternalError,
		message:    message,
		cause:      cause,
	}
}

func Database(message string, cause error) AppError {
	return &appError{
		httpStatus: http.StatusInternalServerError,
		code:       CodeDatabaseError,
		message:    message,
		cause:      cause,
	}
}

func Service(message string, cause error) AppError {
	return &appError{
		httpStatus: http.StatusInternalServerError,
		code:       CodeServiceError,
		message:    message,
		cause:      cause,
	}
}
