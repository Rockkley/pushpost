package usecase

import (
	"context"
	"crypto/subtle"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/rockkley/pushpost/clients/user_api"
	apperr "github.com/rockkley/pushpost/services/auth_service/internal/apperror"
	"github.com/rockkley/pushpost/services/auth_service/internal/domain"
	"github.com/rockkley/pushpost/services/auth_service/internal/domain/dto"
	"github.com/rockkley/pushpost/services/auth_service/internal/email"
	"github.com/rockkley/pushpost/services/auth_service/internal/otp"
	"github.com/rockkley/pushpost/services/auth_service/internal/repository"
	redisotp "github.com/rockkley/pushpost/services/auth_service/internal/repository/redis"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/ctxlog"
	"github.com/rockkley/pushpost/services/common_service/jwt"
	passwordTools "github.com/rockkley/pushpost/services/common_service/password"
)

const (
	otpTTL         = 5 * time.Minute
	resendCooldown = 60 * time.Second
	maxOTPAttempts = 5
)

type AuthUsecase struct {
	userClient   user_api.Client
	sessionStore repository.SessionStore
	otpStore     repository.OTPStore
	emailSender  email.Sender
	jwtManager   *jwt.Manager
}

func NewAuthUsecase(
	userClient user_api.Client,
	sessionStore repository.SessionStore,
	otpStore repository.OTPStore,
	emailSender email.Sender,
	jwtManager *jwt.Manager,
) *AuthUsecase {
	return &AuthUsecase{
		userClient:   userClient,
		sessionStore: sessionStore,
		otpStore:     otpStore,
		emailSender:  emailSender,
		jwtManager:   jwtManager,
	}
}

func (s *AuthUsecase) Register(ctx context.Context, data dto.RegisterUserDto) (*dto.RegisterResponseDto, error) {
	log := ctxlog.From(ctx).With(slog.String("op", "AuthUsecase.Register"))

	hash, err := passwordTools.Hash(data.Password)
	if err != nil {
		log.Error("failed to hash password", slog.Any("error", err))
		return nil, commonapperr.Internal("failed to hash password", err)
	}

	created, err := s.userClient.CreateUser(ctx, user_api.CreateUserRequest{
		Username:     data.Username,
		Email:        data.Email,
		PasswordHash: hash,
	})

	if err != nil {
		log.Warn("user creation failed", slog.Any("error", err))
		return nil, err
	}

	if err = s.sendOTP(ctx, data.Email, log); err != nil {
		log.Error("failed to send verification OTP, user can resend",
			slog.String("user_id", created.ID.String()),
			slog.Any("error", err),
		)
	}

	log.Info("user registered", slog.String("user_id", created.ID.String()))

	return &dto.RegisterResponseDto{
		ID:       created.ID,
		Username: created.Username,
		Email:    created.Email,
	}, nil
}

func (s *AuthUsecase) Login(ctx context.Context, req dto.LoginUserDTO) (string, error) {
	log := ctxlog.From(ctx).With(slog.String("op", "AuthUsecase.Login"))

	user, err := s.userClient.GetUserByEmail(ctx, req.Email)
	if err != nil {
		log.Debug("login attempt: user not found")
		return "", apperr.InvalidCredentials()
	}

	if err = passwordTools.Compare(req.Password, user.PasswordHash); err != nil {
		log.Debug("login attempt: password mismatch", slog.String("user_id", user.ID.String()))
		return "", apperr.InvalidCredentials()
	}

	// Блокируем вход до подтверждения email
	if !user.IsActive() {
		log.Debug("login attempt: account not verified", slog.String("user_id", user.ID.String()))
		return "", apperr.AccountNotVerified()
	}

	deviceID := req.DeviceID
	if deviceID == uuid.Nil {
		deviceID = uuid.New()
	}
	sessionID := uuid.New()

	session := &domain.Session{
		SessionID: sessionID,
		UserID:    user.ID,
		DeviceID:  deviceID,
		Expires:   time.Now().Add(24 * time.Hour).Unix(),
	}

	if err = s.sessionStore.Save(ctx, session); err != nil {
		log.Error("failed to save session", slog.Any("error", err))
		return "", commonapperr.Internal("failed to create session", err)
	}

	token, err := s.jwtManager.Generate(user.ID, deviceID, sessionID)
	if err != nil {
		if delErr := s.sessionStore.Delete(ctx, sessionID); delErr != nil {
			log.Error("failed to delete orphaned session",
				slog.String("session_id", sessionID.String()),
				slog.Any("error", delErr),
			)
		}
		log.Error("failed to generate token", slog.Any("error", err))
		return "", commonapperr.Internal("failed to generate token", err)
	}

	log.Info("user logged in",
		slog.String("user_id", user.ID.String()),
		slog.String("session_id", sessionID.String()),
	)

	return token, nil
}

func (s *AuthUsecase) Logout(ctx context.Context, sessionID uuid.UUID) error {
	log := ctxlog.From(ctx).With(
		slog.String("op", "AuthUsecase.Logout"),
		slog.String("session_id", sessionID.String()),
	)

	session, err := s.sessionStore.Get(ctx, sessionID)
	if err != nil {
		log.Debug("session not found during logout, treating as already logged out")
		return nil
	}

	if err = s.sessionStore.Delete(ctx, sessionID); err != nil {
		log.Error("failed to delete session", slog.Any("error", err))
		return commonapperr.Internal("failed to delete session", err)
	}

	log.Info("user logged out", slog.String("user_id", session.UserID.String()))
	return nil
}

func (s *AuthUsecase) AuthenticateRequest(ctx context.Context, tokenStr string) (*domain.Session, error) {
	log := ctxlog.From(ctx).With(slog.String("op", "AuthUsecase.AuthenticateRequest"))

	claims, err := s.jwtManager.Parse(tokenStr)
	if err != nil {
		log.Debug("token parse failed", slog.Any("error", err))
		return nil, commonapperr.Unauthorized(commonapperr.CodeUnauthorized, "invalid token")
	}

	sessionIDStr, ok := claims["sid"].(string)
	if !ok {
		return nil, commonapperr.Unauthorized(commonapperr.CodeUnauthorized, "token missing session id")
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return nil, commonapperr.Unauthorized(commonapperr.CodeUnauthorized, "invalid session id format")
	}

	session, err := s.sessionStore.Get(ctx, sessionID)
	if err != nil {
		return nil, commonapperr.Unauthorized(commonapperr.CodeUnauthorized, "session not found")
	}

	if time.Now().Unix() > session.Expires {
		if delErr := s.sessionStore.Delete(ctx, sessionID); delErr != nil {
			log.Error("failed to delete expired session", slog.Any("error", delErr))
		}
		return nil, apperr.SessionExpired()
	}

	return session, nil
}

func (s *AuthUsecase) VerifyEmail(ctx context.Context, email, code string) error {
	log := ctxlog.From(ctx).With(slog.String("op", "AuthUsecase.VerifyEmail"))

	stored, err := s.otpStore.Get(ctx, email)
	if err != nil {
		if errors.Is(err, redisotp.ErrOTPNotFound) {
			return apperr.OTPExpired()
		}
		log.Error("failed to get otp", slog.Any("error", err))
		return commonapperr.Internal("get otp", err)
	}

	// Constant-time сравнение защищает от timing-атак
	if subtle.ConstantTimeCompare([]byte(stored), []byte(code)) != 1 {
		attempts, incrErr := s.otpStore.IncrAttempts(ctx, email, otpTTL)
		if incrErr != nil {
			log.Error("failed to increment otp attempts", slog.Any("error", incrErr))
		}

		if attempts >= maxOTPAttempts {
			if delErr := s.otpStore.Delete(ctx, email); delErr != nil {
				log.Error("failed to delete otp after max attempts", slog.Any("error", delErr))
			}
			return apperr.TooManyOTPAttempts()
		}

		return apperr.OTPInvalid()
	}

	if err = s.userClient.ActivateUser(ctx, email); err != nil {
		log.Error("failed to activate user", slog.String("email", email), slog.Any("error", err))
		return err
	}

	if err = s.otpStore.Delete(ctx, email); err != nil {
		// Не критично — OTP истечёт сам по себе
		log.Warn("failed to delete otp after successful verification", slog.Any("error", err))
	}

	log.Info("email verified", slog.String("email", email))
	return nil
}

func (s *AuthUsecase) ResendOTP(ctx context.Context, userEmail string) error {
	log := ctxlog.From(ctx).With(slog.String("op", "AuthUsecase.ResendOTP"))

	onCooldown, err := s.otpStore.HasCooldown(ctx, userEmail)
	if err != nil {
		log.Error("failed to check otp cooldown", slog.Any("error", err))
		return commonapperr.Internal("check otp cooldown", err)
	}
	if onCooldown {
		return apperr.OTPResendCooldown()
	}

	// Проверяем, что пользователь существует и не верифицирован.
	// Отвечаем всегда одинаково (предотвращаем user enumeration).
	user, err := s.userClient.GetUserByEmail(ctx, userEmail)
	if err != nil || user.IsActive() {
		// Успешный ответ, даже если пользователя нет или он уже активен
		return nil
	}

	if err = s.sendOTP(ctx, userEmail, log); err != nil {
		log.Error("failed to resend OTP", slog.Any("error", err))
		return commonapperr.Internal("send otp", err)
	}

	return nil
}

// sendOTP генерирует код, сохраняет в Redis, отправляет письмо.
func (s *AuthUsecase) sendOTP(ctx context.Context, userEmail string, log *slog.Logger) error {
	code, err := otp.Generate()
	if err != nil {
		return err
	}

	if err = s.otpStore.Save(ctx, userEmail, code, otpTTL); err != nil {
		return err
	}

	if err = s.otpStore.SetCooldown(ctx, userEmail, resendCooldown); err != nil {
		log.Warn("failed to set otp resend cooldown", slog.Any("error", err))
	}

	if err = s.emailSender.SendOTP(ctx, userEmail, code); err != nil {
		return err
	}

	return nil
}
