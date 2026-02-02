package apperror

const (
	// Authentication and authorization errors
	CodeInvalidCredentials = "invalid_credentials"
	CodeUnauthorized       = "unauthorized"
	CodeSessionExpired     = "session_expired"
	CodeAccountDeleted     = "account_deleted"

	// Validation Errors
	CodeValidationFailed = "validation_failed"
	CodeFieldRequired    = "field_required"
	CodeFieldInvalid     = "field_invalid"
	CodeFieldTooShort    = "field_too_short"
	CodeFieldTooLong     = "field_too_long"
	CodeFieldWeak        = "field_weak"

	// Conflict Errors
	CodeEmailExists    = "email_already_exists"
	CodeUsernameExists = "username_already_exists"
	CodeAlreadyExists  = "already_exists"

	// Not Found Errors
	CodeUserNotFound    = "user_not_found"
	CodeMessageNotFound = "message_not_found"

	// Bussiness Logic Errors
	CodeCannotMessageSelf  = "cannot_message_self"
	CodeCannotBefriendSelf = "cannot_befriend_self"

	// System Errors
	CodeInternalError = "internal_error"
	CodeDatabaseError = "database_error"
	CodeServiceError  = "service_error"
)
