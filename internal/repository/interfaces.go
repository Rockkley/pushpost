package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

type RelationshipRepository interface {
	CreateRequest(ctx context.Context, senderID, receiverID uuid.UUID) error
	AcceptRequest(ctx context.Context, senderID, receiverID uuid.UUID) error
	RejectRequest(ctx context.Context, senderID, receiverID uuid.UUID) error
	CancelRequest(ctx context.Context, senderID, receiverID uuid.UUID) error
	DeleteFriendship(ctx context.Context, userID, friendID uuid.UUID) error
	AreFriends(ctx context.Context, user1, user2 uuid.UUID) (bool, error)
	GetFriendsIds(ctx context.Context, userID uuid.UUID) ([]*uuid.UUID, error)
	RequestExists(ctx context.Context, senderID, receiverID uuid.UUID) (bool, error)
}

type MessageRepository interface {
	Create(ctx context.Context, message *domain.Message) (*domain.Message, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Message, error)
	GetConversationWithUsers(ctx context.Context, userID, otherUserID uuid.UUID, limit, offset int) ([]*domain.Message, error)
	MarkAsRead(ctx context.Context, messageID, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, senderID, receiverID uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
	GetUnreadMessages(ctx context.Context, userID uuid.UUID) ([]*domain.Message, error)
}

type SessionStore interface {
	Save(ctx context.Context, session *domain.Session) error
	Get(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error)
	Delete(ctx context.Context, sessionID uuid.UUID) error
}
