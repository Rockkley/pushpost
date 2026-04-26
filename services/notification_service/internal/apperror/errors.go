package apperror

import commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"

func InvalidTelegramCode() commonapperr.AppError {
	return commonapperr.BadRequest(CodeInvalidTelegramCode, "telegram link code is invalid or expired")
}

func InvalidNotificationID() commonapperr.AppError {
	return commonapperr.BadRequest(CodeInvalidNotificationID, "invalid notification id")
}

func InvalidJSON() commonapperr.AppError {
	return commonapperr.BadRequest(CodeInvalidJSON, "invalid JSON")
}

func TypeRequired() commonapperr.AppError {
	return commonapperr.Validation(CodeTypeRequired, "type", "type is required")
}

func ChannelRequired() commonapperr.AppError {
	return commonapperr.Validation(CodeChannelRequired, "channel", "channel is required")
}

func MissingUserID() commonapperr.AppError {
	return commonapperr.Unauthorized(CodeMissingUserID, "missing authenticated user")
}
