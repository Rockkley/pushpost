package domain

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/common_service/outbox"
	"github.com/rockkley/pushpost/services/post_service/internal/entity"
	"github.com/rockkley/pushpost/services/post_service/internal/repository"
	"time"
)

type OutboxWriterInterface interface {
	outbox.WriterInterface
}

type Tx interface {
	Posts() repository.PostRepositoryInterface
	Outbox() OutboxWriterInterface
}

type UnitOfWorkInterface interface {
	Do(ctx context.Context, fn func(Tx) error) error
	Reader() repository.PostRepositoryInterface
}

// FriendshipClient — зависимость usecase на внешний сервис.
// Определяем здесь, чтобы domain не зависел от конкретного gRPC клиента.
type FriendshipClient interface {
	GetFriendIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

type PostUseCaseInterface interface {
	CreatePost(ctx context.Context, authorID uuid.UUID, content string) (*entity.Post, error)
	GetFeed(ctx context.Context, userID uuid.UUID, limit int, cursor string) ([]*entity.Post, string, error)
	GetUserPosts(ctx context.Context, authorID uuid.UUID, limit int, cursor string) ([]*entity.Post, string, error)
	DeletePost(ctx context.Context, postID, authorID uuid.UUID) error
	GetPostByID(ctx context.Context, postID uuid.UUID) (*entity.Post, error)
}

// Cursor helpers
type Cursor struct {
	Before   time.Time
	BeforeID uuid.UUID
}
