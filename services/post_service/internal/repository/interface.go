package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/post_service/internal/entity"
	"time"
)

type PostRepositoryInterface interface {
	Create(ctx context.Context, post *entity.Post) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Post, error)
	GetByAuthors(ctx context.Context, authorIDs []uuid.UUID, limit int, before time.Time, beforeID uuid.UUID) ([]*entity.Post, error)
	GetByAuthor(ctx context.Context, authorID uuid.UUID, limit int, before time.Time, beforeID uuid.UUID) ([]*entity.Post, error)
	SoftDelete(ctx context.Context, postID, authorID uuid.UUID) error
}
