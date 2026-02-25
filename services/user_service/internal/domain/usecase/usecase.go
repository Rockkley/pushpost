package usecase

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/common/apperror"
	"github.com/rockkley/pushpost/services/user_service/internal/ctxlog"
	"github.com/rockkley/pushpost/services/user_service/internal/domain/dto"
	"github.com/rockkley/pushpost/services/user_service/internal/entity"
	"github.com/rockkley/pushpost/services/user_service/internal/repository"
	"log/slog"
)

type UserUseCase struct {
	repo repository.UserRepository
}

func NewUserUseCase(repo repository.UserRepository) *UserUseCase {
	return &UserUseCase{repo: repo}
}

func (u *UserUseCase) AuthenticateUser(ctx context.Context, dto dto.AuthenticateUserRequestDTO) (*entity.User, error) {
	log := ctxlog.From(ctx).With(slog.String("op", "UserUseCase.AuthenticateUser"))

	user, err := u.repo.GetUserByEmail(ctx, dto.Email)

	if err != nil {
		log.Debug("user not found by email")
		return nil, apperror.InvalidCredentials()
	}

	if user.IsDeleted() {
		log.Warn("auth attempt on deleted account", slog.String("user_id", user.Id.String()))
		return nil, apperror.InvalidCredentials()
	}

	//if err = validator.ValidatePassword(password); err != nil {
	//
	//	return nil, apperror.InvalidCredentials()
	//}

	return user, nil
}

func (u *UserUseCase) CreateUser(ctx context.Context, dto dto.CreateUserDTO) (*entity.User, error) {
	log := ctxlog.From(ctx).With(slog.String("op", "UserUseCase.CreateUser"))

	if err := dto.Validate(); err != nil {
		log.Debug("dto validation failed", slog.Any("error", err))
		return nil, err
	}

	user := &entity.User{
		Id:           uuid.New(),
		Username:     dto.Username,
		Email:        dto.Email,
		PasswordHash: dto.PasswordHash,
	}

	if err := u.repo.Create(ctx, user); err != nil {
		log.Warn("failed to create user", slog.Any("error", err))
		return nil, err
	}

	log.Info("user created", slog.String("user_id", user.Id.String()))
	return user, nil
}

func (u *UserUseCase) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	log := ctxlog.From(ctx).With(slog.String("op", "UserUseCase.FindUserByEmail"))

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
