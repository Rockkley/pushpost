package usecase

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/rockkley/pushpost/pkg/ctxlog"
	"github.com/rockkley/pushpost/services/common/apperror"
	"github.com/rockkley/pushpost/services/user_service/internal/domain/dto"
	"github.com/rockkley/pushpost/services/user_service/internal/entity"
	"github.com/rockkley/pushpost/services/user_service/internal/repository"
)

type UserUseCase struct {
	repo repository.UserRepository
}

func NewUserUseCase(repo repository.UserRepository) *UserUseCase {
	return &UserUseCase{repo: repo}
}

func (u *UserUseCase) AuthenticateUser(ctx context.Context, req dto.AuthenticateUserRequestDTO) (*entity.User, error) {
	log := ctxlog.From(ctx).With(slog.String("op", "UserUseCase.AuthenticateUser"))

	user, err := u.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		log.Debug("user not found by email")
		return nil, apperror.InvalidCredentials()
	}

	if user.IsDeleted() {
		log.Warn("auth attempt on deleted account", slog.String("user_id", user.Id.String()))
		return nil, apperror.InvalidCredentials()
	}

	return user, nil
}

func (u *UserUseCase) CreateUser(ctx context.Context, req dto.CreateUserDTO) (*entity.User, error) {
	log := ctxlog.From(ctx).With(slog.String("op", "UserUseCase.CreateUser"))

	if err := req.Validate(); err != nil {
		log.Debug("dto validation failed", slog.Any("error", err))
		return nil, err
	}

	user := &entity.User{
		Id:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: req.PasswordHash,
	}

	if err := u.repo.Create(ctx, user); err != nil {
		log.Warn("failed to create user", slog.Any("error", err))
		return nil, err
	}

	log.Info("user created", slog.String("user_id", user.Id.String()))
	return user, nil
}

func (u *UserUseCase) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	log := ctxlog.From(ctx).With(slog.String("op", "UserUseCase.GetUserByEmail"))

	user, err := u.repo.GetUserByEmail(ctx, email)
	if err != nil {
		log.Debug("user not found by email")
		return nil, apperror.NotFound(apperror.CodeUserNotFound, "user not found")
	}

	if user.IsDeleted() {
		log.Warn("attempt to find deleted account by email", slog.String("user_id", user.Id.String()))
		return nil, apperror.NotFound(apperror.CodeUserDeleted, "user is deleted")
	}

	return user, nil
}
