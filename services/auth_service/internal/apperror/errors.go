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

func AccountNotVerified() apperror.AppError {
	return apperror.Forbidden(CodeAccountNotVerified, "email not verified, please check your inbox")
}

func OTPExpired() apperror.AppError {
	return apperror.BadRequest(CodeOTPExpired, "OTP has expired, please request a new one")
}

func OTPInvalid() apperror.AppError {
	return apperror.BadRequest(CodeOTPInvalid, "invalid OTP code")
}

func OTPResendCooldown() apperror.AppError {
	return apperror.BadRequest(CodeOTPResendCooldown, "please wait before requesting a new OTP")
}

func TooManyOTPAttempts() apperror.AppError {
	return apperror.BadRequest(CodeTooManyOTPAttempts, "too many incorrect attempts, please request a new OTP")
}
