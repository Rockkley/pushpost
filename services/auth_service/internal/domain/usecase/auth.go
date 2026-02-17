package usecase

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/pkg/clients/user_api"
	"github.com/rockkley/pushpost/pkg/jwt"
	passwordTools "github.com/rockkley/pushpost/pkg/password"
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
	var err error

	data.Password, err = passwordTools.Hash(data.Password)
	if err != nil {

		return nil, apperror.Internal("failed to hash password", err)
	}

	req := user_api.CreateUserRequest{
		Username:     data.Username,
		Email:        data.Email,
		PasswordHash: data.Password,
	}

	createResponse, err := s.userClient.CreateUser(ctx, req)

	if err != nil {

		return nil, err
	}

	regResponse := &dto2.RegisterResponseDto{
		ID:       createResponse.ID,
		Username: createResponse.Username,
		Email:    createResponse.Email,
	}
	return regResponse, nil
}

func (s *AuthUsecase) Login(ctx context.Context, dto dto2.LoginUserDTO) (string, error) {

	user, err := s.userClient.AuthenticateUser(ctx, dto.Email, dto.PasswordHash) // fix me
	if err != nil {
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
		return "", apperror.Internal("failed to create session", err)
	}

	token, err := s.jwtManager.Generate(user.ID, dto.DeviceID, sessionID)
	if err != nil {
		return "", apperror.Internal("failed to generate token", err)
	}

	return token, nil
}

func (s *AuthUsecase) Logout(ctx context.Context, sessionID uuid.UUID) error {
	session, err := s.sessionStore.Get(ctx, sessionID)
	if err != nil {
		slog.Debug("session not found during logout")
		slog.String("session_id", sessionID.String())
		return nil
	}

	if err = s.sessionStore.Delete(ctx, sessionID); err != nil {
		return apperror.Internal("failed to delete session", err)
	}

	slog.Info("user_service logged out",
		slog.String("user_id", session.UserID.String()),
		slog.String("session_id", session.SessionID.String()),
	)

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
	claims, err := s.jwtManager.Parse(tokenStr)
	if err != nil {
		slog.Debug("invalid token signature", slog.Any("error", err))
		return nil, apperror.Unauthorized(apperror.CodeUnauthorized, "invalid token signature")
	}

	sessionIDStr, ok := claims["sid"].(string)
	if !ok {
		return nil, apperror.Unauthorized(apperror.CodeUnauthorized, "token missing session id")
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return nil, apperror.Unauthorized(apperror.CodeUnauthorized, "invalid session id format")
	}
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, apperror.Unauthorized(apperror.CodeUnauthorized, "session not found")
	}

	if time.Now().Unix() > session.Expires {
		if err = s.sessionStore.Delete(ctx, sessionID); err != nil {
			return nil, apperror.Internal("failed to delete session", err)
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
