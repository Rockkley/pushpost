package usecase

import (
	"context"
	"github.com/rockkley/pushpost/internal/apperror"
	"github.com/rockkley/pushpost/internal/services/user_service/internal/entity"
	"github.com/rockkley/pushpost/internal/services/user_service/internal/repository"
	"github.com/rockkley/pushpost/internal/validator"
)

type UserUseCase struct {
	repo repository.UserRepository
}

func NewUserUseCase(repo repository.UserRepository) *UserUseCase {
	return &UserUseCase{repo: repo}
}

func (u *UserUseCase) AuthenticateUser(ctx context.Context, email, password string) (*entity.User, error) {
	user, err := u.repo.FindByEmail(ctx, email)
	if err != nil {

		return nil, apperror.InvalidCredentials()
	}

	if user.IsDeleted() {

		return nil, apperror.InvalidCredentials() // don't reveal if the user exists
	}

	if err = validator.ValidatePassword(password); err != nil {

		return nil, apperror.InvalidCredentials()
	}

	return user, nil
}
