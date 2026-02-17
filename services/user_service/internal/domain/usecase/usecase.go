package usecase

import (
	"context"
	"github.com/google/uuid"
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

func (u *UserUseCase) AuthenticateUser(ctx context.Context, email, password string) (*entity.User, error) {
	user, err := u.repo.FindByEmail(ctx, email)
	if err != nil {

		return nil, apperror.InvalidCredentials()
	}

	if user.IsDeleted() {

		return nil, apperror.InvalidCredentials() // don't reveal if the user exists
	}

	//if err = validator.ValidatePassword(password); err != nil {
	//
	//	return nil, apperror.InvalidCredentials()
	//}

	return user, nil
}

func (u *UserUseCase) CreateUser(ctx context.Context, dto dto.CreateUserDTO) (*entity.User, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	user := &entity.User{
		Id:           uuid.New(),
		Username:     dto.Username,
		Email:        dto.Email,
		PasswordHash: dto.PasswordHash,
	}

	if err := u.repo.Create(ctx, user); err != nil {

		return nil, err
	}

	return user, nil
}
