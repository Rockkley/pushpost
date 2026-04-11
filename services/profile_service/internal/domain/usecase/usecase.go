package usecase

import (
	"context"
	"github.com/rockkley/pushpost/services/profile_service/internal/repository"

	"github.com/rockkley/pushpost/services/profile_service/internal/entity"
)

type ProfileUseCase struct {
	profileRepo repository.ProfileRepositoryInterface
}

func NewProfileUseCase(profileRepo repository.ProfileRepositoryInterface) *ProfileUseCase {
	return &ProfileUseCase{profileRepo: profileRepo}
}

func (u *ProfileUseCase) GetByUsername(ctx context.Context, username string) (*entity.Profile, error) {
	return u.profileRepo.FindByUsername(ctx, username)
}

func (u *ProfileUseCase) CreateProfile(ctx context.Context, profile *entity.Profile) error {
	return u.profileRepo.Create(ctx, profile)
}
