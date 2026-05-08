package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/common_service/outbox"
	"github.com/rockkley/pushpost/services/post_service/internal/entity"
	"github.com/rockkley/pushpost/services/post_service/internal/repository"
)

type OutboxWriterInterface interface {
	outbox.WriterInterface
}

type Tx interface {
	Comments() repository.CommentRepositoryInterface
	Posts() repository.PostRepositoryInterface
	Outbox() OutboxWriterInterface
}

type UnitOfWorkInterface interface {
	Do(ctx context.Context, fn func(Tx) error) error
	Reader() repository.PostRepositoryInterface
	CommentReader() repository.CommentRepositoryInterface
}

type FriendshipClient interface {
	GetFriendIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

type CommentsResponse struct {
	Comments   []*entity.Comment
	NextCursor string
}

type FeedResponse struct {
	Posts      []*entity.Post
	NextCursor string
	TopCursor  string
}

type PostUseCaseInterface interface {
	CreatePost(ctx context.Context, authorID uuid.UUID, content string) (*entity.Post, error)
	UpdatePost(ctx context.Context, postID, authorID uuid.UUID, content string) (*entity.Post, error)
	GetFeed(ctx context.Context, userID uuid.UUID, limit int, cursor string) (FeedResponse, error)
	GetFeedSince(ctx context.Context, userID uuid.UUID, limit int, since string) (FeedResponse, error)
	GetUserPosts(ctx context.Context, authorID uuid.UUID, limit int, cursor string) (FeedResponse, error)
	DeletePost(ctx context.Context, postID, authorID uuid.UUID) error
	GetPostByID(ctx context.Context, postID uuid.UUID) (*entity.Post, error)
	GetPostsByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.Post, error)
	LikePost(ctx context.Context, postID, userID uuid.UUID) (*entity.Post, error)
	DislikePost(ctx context.Context, postID, userID uuid.UUID) (*entity.Post, error)
	RemovePostVote(ctx context.Context, postID, userID uuid.UUID) (*entity.Post, error)
}

type CommentUseCaseInterface interface {
	CreateComment(ctx context.Context, postID, authorID uuid.UUID, parentID *uuid.UUID, content string) (*entity.Comment, error)
	GetPostComments(ctx context.Context, postID uuid.UUID, limit int, cursor string) (CommentsResponse, error)
	UpdateComment(ctx context.Context, commentID, authorID uuid.UUID, content string) (*entity.Comment, error)
	UpvoteComment(ctx context.Context, commentID, userID uuid.UUID) (*entity.Comment, error)
	DownvoteComment(ctx context.Context, commentID, userID uuid.UUID) (*entity.Comment, error)
	RemoveCommentVote(ctx context.Context, commentID, userID uuid.UUID) (*entity.Comment, error)
	DeleteComment(ctx context.Context, commentID, authorID uuid.UUID) error
}

type Cursor struct {
	Before   time.Time
	BeforeID uuid.UUID
}
