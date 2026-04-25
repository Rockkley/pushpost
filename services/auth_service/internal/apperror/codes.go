package apperror

const (
	CodeInvalidCredentials = "invalid_credentials"
	CodeSessionExpired     = "session_expired"
	CodeAccountDeleted     = "account_deleted"
	CodeAccountNotVerified = "account_not_verified"
	CodeOTPExpired         = "otp_expired"
	CodeOTPInvalid         = "otp_invalid"
	CodeOTPResendCooldown  = "otp_resend_cooldown"
	CodeTooManyOTPAttempts = "too_many_otp_attempts"
)
