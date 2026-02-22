package usecase

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/pkg/clients/user_api"
	"github.com/rockkley/pushpost/pkg/jwt"
	passwordTools "github.com/rockkley/pushpost/pkg/password"
	"github.com/rockkley/pushpost/services/auth_service/internal/ctxlog"
	"github.com/rockkley/pushpost/services/auth_service/internal/domain"
	"github.com/rockkley/pushpost/services/auth_service/internal/repository"
	dto2 "github.com/rockkley/pushpost/services/auth_service/internal/transport/http/dto"
	"github.com/rockkley/pushpost/services/common/apperror"
	"log/slog"
	"time"
)

type AuthUsecase struct {
	userClient   user_api.Client
	sessionStore repository.SessionStore
	jwtManager   *jwt.Manager
}

func NewAuthUsecase(
	userClient user_api.Client,
	sessionStore repository.SessionStore,
	jwtManager *jwt.Manager) *AuthUsecase {
	return &AuthUsecase{
		userClient:   userClient,
		sessionStore: sessionStore,
		jwtManager:   jwtManager}
}

func (s *AuthUsecase) Register(ctx context.Context, data dto2.RegisterUserDto) (*dto2.RegisterResponseDto, error) {
	log := ctxlog.From(ctx).With(slog.String("method", "AuthUsecase.Register"))
	hash, err := passwordTools.Hash(data.Password)

	if err != nil {
		log.Error("failed to hash password", slog.Any("error", err))

		return nil, apperror.Internal("failed to hash password", err)
	}

	req := user_api.CreateUserRequest{
		Username:     data.Username,
		Email:        data.Email,
		PasswordHash: hash,
	}

	createdResponse, err := s.userClient.CreateUser(ctx, req)

	if err != nil {
		log.Warn("user creation failed", slog.Any("error", err))

		return nil, err
	}

	log.Info("user registered", slog.String("user_id", createdResponse.ID.String()))

	return &dto2.RegisterResponseDto{
		ID:       createdResponse.ID,
		Username: createdResponse.Username,
		Email:    createdResponse.Email,
	}, nil
}

func (s *AuthUsecase) Login(ctx context.Context, dto dto2.LoginUserDTO) (string, error) {
	log := ctxlog.From(ctx).With(slog.String("op", "AuthUsecase.Login"))
	hashedPassword, err := passwordTools.Hash(dto.Password)
	user, err := s.userClient.AuthenticateUser(ctx, dto.Email, hashedPassword) // fix me
	if err != nil {
		log.Warn("authentication failed", slog.Any("error", err))
		return "", err
	}

	//if s.userClient.IsDeleted(ctx, user.Id) {
	//	return "", apperror.AccountDeleted() // TODO think of a good way to check it, comment out for now just for tests
	//}

	deviceID := dto.DeviceID
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
		return "", apperror.Internal("failed to create session", err)
	}

	token, err := s.jwtManager.Generate(user.ID, dto.DeviceID, sessionID)
	if err != nil {
		log.Error("failed to generate token", slog.Any("error", err))
		return "", apperror.Internal("failed to generate token", err)
	}

	log.Info("user logged in",
		slog.String("user_id", user.ID.String()),
		slog.String("session_id", sessionID.String()),
		slog.String("device_id", deviceID.String()),
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
		log.Debug(fmt.Sprintf(
			"session_id (%s) not found during logout, treating as already logged out", sessionID.String()),
			slog.Any("error", err))

		return nil
	}

	if err = s.sessionStore.Delete(ctx, sessionID); err != nil {
		log.Error("failed to delete session", slog.Any("error", err))
		return apperror.Internal("failed to delete session", err)
	}

	log.Info("user logged out", slog.String("user_id", session.UserID.String()))

	return nil
}

func (s *AuthUsecase) GetSession(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error) {
	session, err := s.sessionStore.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *AuthUsecase) AuthenticateRequest(ctx context.Context, tokenStr string) (*domain.Session, error) {
	log := ctxlog.From(ctx).With(slog.String("op", "AuthUsecase.AuthenticateRequest"))

	claims, err := s.jwtManager.Parse(tokenStr)
	if err != nil {
		log.Debug("token parse failed", slog.Any("error", err))
		return nil, apperror.Unauthorized(apperror.CodeUnauthorized, "invalid token")
	}

	sessionIDStr, ok := claims["sid"].(string)
	if !ok {
		log.Warn("token missing sid claim")
		return nil, apperror.Unauthorized(apperror.CodeUnauthorized, "token missing session id")
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		log.Warn("invalid session id in token", slog.String("sid", sessionIDStr))
		return nil, apperror.Unauthorized(apperror.CodeUnauthorized, "invalid session id format")
	}
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		log.Debug("session not found", slog.String("session_id", sessionID.String()))
		return nil, apperror.Unauthorized(apperror.CodeUnauthorized, "session not found")
	}

	if time.Now().Unix() > session.Expires {
		log.Info("session expired", slog.String("session_id", sessionID.String()))
		if delErr := s.sessionStore.Delete(ctx, sessionID); delErr != nil {
			log.Error("failed to delete expired session", slog.Any("error", delErr))
			return nil, apperror.Internal("failed to delete expired session", delErr)
		}
		return nil, apperror.SessionExpired()
	}

	return session, nil
}

//func (s *AuthUsecase) extractTokenClaims() todo?
//func (s *AuthUsecase) ValidateTokenStructure(tokenStr string) (tokenID string, err error) {
//	claims, err := s.jwtManager.Parse(tokenStr)
//	if err != nil {
//		return "", fmt.Errorf("invalid token signature: %w", err)
//	}
//
//
//
//	return tokenID, nil
//}
