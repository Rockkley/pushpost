package services

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/internal/apperror"
	"github.com/rockkley/pushpost/internal/domain"
	"github.com/rockkley/pushpost/internal/handler/http/dto"
	"github.com/rockkley/pushpost/internal/repository"
	"github.com/rockkley/pushpost/pkg/jwt"
	passwordTools "github.com/rockkley/pushpost/pkg/password"
	"log/slog"
	"time"
)

type AuthService struct {
	userRepo     repository.UserRepository
	sessionStore repository.SessionStore
	jwtManager   *jwt.Manager
}

func NewAuthService(
	userRepo repository.UserRepository,
	sessionStore repository.SessionStore,
	jwtManager *jwt.Manager) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		sessionStore: sessionStore,
		jwtManager:   jwtManager}
}

func (s *AuthService) Register(ctx context.Context, dto dto.RegisterUserDto) (*domain.User, error) {

	hashedPassword, err := passwordTools.Hash(dto.Password)
	if err != nil {

		return nil, apperror.Internal("failed to hash password", err)
	}

	user := &domain.User{
		Id:           uuid.New(),
		Username:     dto.Username,
		Email:        dto.Email,
		PasswordHash: hashedPassword,
	}

	err = s.userRepo.Create(ctx, user)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, dto dto.LoginUserDTO) (string, error) {
	user, err := s.userRepo.FindByEmail(ctx, dto.Email)
	if err != nil {
		return "", err
	}

	if user.IsDeleted() {
		return "", apperror.AccountDeleted()
	}

	if err = passwordTools.Compare(dto.Password, user.PasswordHash); err != nil {
		return "", apperror.InvalidCredentials()
	}

	deviceID := dto.DeviceID
	if deviceID == uuid.Nil {
		deviceID = uuid.New()
	}
	sessionID := uuid.New()
	session := &domain.Session{
		SessionID: sessionID,
		UserID:    user.Id,
		DeviceID:  deviceID,
		Expires:   time.Now().Add(24 * time.Hour).Unix(),
	}

	if err = s.sessionStore.Save(ctx, session); err != nil {
		return "", apperror.Internal("failed to create session", err)
	}

	token, err := s.jwtManager.Generate(user.Id, dto.DeviceID, sessionID)
	if err != nil {
		return "", apperror.Internal("failed to generate token", err)
	}

	return token, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	session, err := s.sessionStore.Get(ctx, sessionID)
	if err != nil {
		slog.Debug("session not found during logout")
		slog.String("session_id", sessionID.String())
		return nil
	}

	if err = s.sessionStore.Delete(ctx, sessionID); err != nil {
		return apperror.Internal("failed to delete session", err)
	}

	slog.Info("user logged out",
		slog.String("user_id", session.UserID.String()),
		slog.String("session_id", session.SessionID.String()),
	)

	return nil
}

func (s *AuthService) GetSession(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error) {
	session, err := s.sessionStore.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *AuthService) AuthenticateRequest(ctx context.Context, tokenStr string) (*domain.Session, error) {
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

//func (s *AuthService) extractTokenClaims() todo?
//func (s *AuthService) ValidateTokenStructure(tokenStr string) (tokenID string, err error) {
//	claims, err := s.jwtManager.Parse(tokenStr)
//	if err != nil {
//		return "", fmt.Errorf("invalid token signature: %w", err)
//	}
//
//
//
//	return tokenID, nil
//}
