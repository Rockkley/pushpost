package apperror

import "github.com/rockkley/pushpost/services/common_service/apperror"

func InvalidErrorEnvelope(err error) apperror.AppError {
	return apperror.BadRequest(CodeInvalidEventEnvelope, err.Error())
}

func InvalidEventType(err error) apperror.AppError {
	return apperror.BadRequest(CodeInvalidEventType, err.Error())
}
