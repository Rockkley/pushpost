package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/internal/domain"
	"github.com/rockkley/pushpost/internal/handler/http/dto"
	"github.com/rockkley/pushpost/internal/repository"
	"github.com/rockkley/pushpost/pkg/jwt"
	passwordTools "github.com/rockkley/pushpost/pkg/password"
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

		return nil, err
	}

	user := &domain.User{
		Id:           uuid.New(),
		Username:     dto.Username,
		Email:        dto.Email,
		PasswordHash: hashedPassword,
	}

	err = s.userRepo.Create(ctx, user)

	var de domain.DomainError

	if err != nil {
		if errors.As(err, &de) {

			return nil, err
		}

		return nil, fmt.Errorf("internal error - %w", err)
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, dto dto.LoginUserDTO) (string, error) {
	user, err := s.userRepo.FindByEmail(ctx, dto.Email)
	if err != nil {
		return "", err
	}

	if user.IsDeleted() {
		return "", errors.New("account is deleted")
	}

	if err = passwordTools.Compare(user.PasswordHash, dto.Password); err != nil {
		return "", errors.New("invalid password")
	}

	sessionID := uuid.NewString()
	session := &domain.Session{
		SessionID: sessionID,
		UserID:    user.Id.String(),
		DeviceID:  dto.DeviceID,
		Expires:   time.Now().Add(24 * time.Hour).Unix(),
	}

	if err = s.sessionStore.Save(ctx, session); err != nil {
		return "", err
	}

	token, err := s.jwtManager.Generate(user.Id.String(), dto.DeviceID, sessionID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	return s.sessionStore.Delete(ctx, sessionID)
}

func (s *AuthService) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	session, err := s.sessionStore.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *AuthService) AuthenticateRequest(ctx context.Context, tokenStr string) (*domain.Session, error) {
	claims, err := s.jwtManager.Parse(tokenStr)
	if err != nil {
		return nil, fmt.Errorf("invalid token signature: %w", err)
	}
	sessionID, ok := claims["sid"].(string)
	if !ok {
		return nil, errors.New("token missing session ID")
	}

	return s.GetSession(ctx, sessionID)
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
