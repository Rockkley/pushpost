package postgres

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/common/apperror"
	"github.com/rockkley/pushpost/services/friendship/entity"
)

type FriendshipRequestRepository struct {
	db *sql.DB
}

func (r FriendshipRequestRepository) CreateRequest(ctx context.Context, request *entity.FriendshipRequest) error {
	query := `
INSERT INTO friendship_requests (id, sender_id, receiver_id, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query,
		request.ID, request.SenderID, request.ReceiverID, request.Status, request.CreatedAt, request.UpdatedAt,
	)
	if err != nil {
		return apperror.MapPostgresError(err, "create friendship request")
	}
	return nil
}

func (r FriendshipRequestRepository) AcceptRequest(ctx context.Context, senderID, receiverID uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (r FriendshipRequestRepository) RejectRequest(ctx context.Context, senderID, receiverID uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (r FriendshipRequestRepository) CancelRequest(ctx context.Context, senderID, receiverID uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (r FriendshipRequestRepository) RequestExists(ctx context.Context, senderID, receiverID uuid.UUID) (bool, error) {
	//TODO implement me
	panic("implement me")
}
