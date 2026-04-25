package repository

import (
	"context"
	"github.com/rockkley/pushpost/services/profile_service/internal/entity"
)

type ProfileRepositoryInterface interface {
	Create(ctx context.Context, profile *entity.Profile) error
	FindByUsername(ctx context.Context, username string) (*entity.Profile, error)
	Update(ctx context.Context, profile *entity.Profile) error
}
