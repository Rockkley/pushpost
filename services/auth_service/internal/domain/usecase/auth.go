package usecase

import (
	"context"
	"github.com/rockkley/pushpost/clients/user_api"
	"github.com/rockkley/pushpost/services/common_service/ctxlog"
	"github.com/rockkley/pushpost/services/common_service/jwt"
	passwordTools "github.com/rockkley/pushpost/services/common_service/password"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/rockkley/pushpost/services/auth_service/internal/domain"
	"github.com/rockkley/pushpost/services/auth_service/internal/domain/dto"
	"github.com/rockkley/pushpost/services/auth_service/internal/repository"
	"github.com/rockkley/pushpost/services/common_service/apperror"
)

type AuthUsecase struct {
	userClient   user_api.Client
	sessionStore repository.SessionStore
	jwtManager   *jwt.Manager
}

func NewAuthUsecase(
	userClient user_api.Client,
	sessionStore repository.SessionStore,
	jwtManager *jwt.Manager,
) *AuthUsecase {
	return &AuthUsecase{
		userClient:   userClient,
		sessionStore: sessionStore,
		jwtManager:   jwtManager,
	}
}

func (s *AuthUsecase) Register(ctx context.Context, data dto.RegisterUserDto) (*dto.RegisterResponseDto, error) {
	log := ctxlog.From(ctx).With(slog.String("op", "AuthUsecase.Register"))

	hash, err := passwordTools.Hash(data.Password)

	if err != nil {
		log.Error("failed to hash password", slog.Any("error", err))

		return nil, apperror.Internal("failed to hash password", err)
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
		return "", apperror.InvalidCredentials()
	}

	if err = passwordTools.Compare(req.Password, user.PasswordHash); err != nil {
		log.Debug("login attempt: password mismatch", slog.String("user_id", user.ID.String()))
		return "", apperror.InvalidCredentials()
	}

	// ensure deviceID is never nil - generate one if the client didn't supply it.
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
		return "", apperror.Internal("failed to create session", err)
	}

	token, err := s.jwtManager.Generate(user.ID, deviceID, sessionID)
	if err != nil {
		// session was already persisted - clean it up so it doesn't become orphaned
		if delErr := s.sessionStore.Delete(ctx, sessionID); delErr != nil {
			log.Error("failed to delete orphaned session after token error",
				slog.String("session_id", sessionID.String()),
				slog.Any("error", delErr),
			)
		}
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
		// not an error - session may have already expired or been deleted
		log.Debug("session not found during logout, treating as already logged out",
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

	session, err := s.sessionStore.Get(ctx, sessionID)
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
