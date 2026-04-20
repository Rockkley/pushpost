package apperror

import "github.com/rockkley/pushpost/services/common_service/apperror"

func MessageNotFound() apperror.AppError {
	return apperror.NotFound(CodeMessageNotFound, "message not found")
}

func CannotMessageSelf() apperror.AppError {
	return apperror.BadRequest(CodeCannotMessageSelf, "cannot send message to yourself")
}

func MessageTooLong() apperror.AppError {
	return apperror.Validation(CodeMessageTooLong, "content", "message exceeds maximum length of 10000 characters")
}

func MessageEmpty() apperror.AppError {
	return apperror.Validation(CodeMessageEmpty, "content", "message cannot be empty")
}

func NotReceiver() apperror.AppError {
	return apperror.Forbidden(CodeNotReceiver, "you are not the receiver of this message")
}
