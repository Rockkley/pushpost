package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/post_service/internal/entity"
)

type PostRepositoryInterface interface {
	Create(ctx context.Context, post *entity.Post) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Post, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.Post, error)
	GetByAuthors(ctx context.Context, authorIDs []uuid.UUID, limit int, before time.Time, beforeID uuid.UUID) ([]*entity.Post, error)
	GetByAuthor(ctx context.Context, authorID uuid.UUID, limit int, before time.Time, beforeID uuid.UUID) ([]*entity.Post, error)
	Update(ctx context.Context, post *entity.Post) error
	SoftDelete(ctx context.Context, postID, authorID uuid.UUID) error
}

type FeedRepository interface {
	InsertBatch(ctx context.Context, postID uuid.UUID, userIDs []uuid.UUID, insertedAt time.Time) error
	GetFeed(ctx context.Context, userID uuid.UUID, limit int, before time.Time, beforeID uuid.UUID) ([]*entity.Post, error)
	GetFeedSince(ctx context.Context, userID uuid.UUID, limit int, after time.Time, afterID uuid.UUID) ([]*entity.Post, error)
	FindRecipients(ctx context.Context, postID uuid.UUID) ([]uuid.UUID, error)
	DeleteByPostID(ctx context.Context, postID uuid.UUID) error
	DeleteByAuthor(ctx context.Context, recipientID, authorID uuid.UUID) error
	DeleteUserFeed(ctx context.Context, userID uuid.UUID) error
	DeleteByAuthorFromAllFeeds(ctx context.Context, authorID uuid.UUID) error
}
