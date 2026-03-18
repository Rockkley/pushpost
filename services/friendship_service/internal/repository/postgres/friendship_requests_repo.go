package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/database"
	apperr "github.com/rockkley/pushpost/services/friendship_service/internal/apperror"
	"github.com/rockkley/pushpost/services/friendship_service/internal/entity"
	"github.com/rockkley/pushpost/services/friendship_service/internal/repository"
)

type friendshipRequestRepository struct {
	exec database.Executor
}

func NewFriendshipRequestRepository(exec database.Executor) repository.FriendshipRequestRepository {
	return &friendshipRequestRepository{exec: exec}
}

func (r *friendshipRequestRepository) Create(ctx context.Context, req *entity.FriendshipRequest) error {
	query := `
		INSERT INTO friendship_requests (id, sender_id, receiver_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())`

	_, err := r.exec.ExecContext(ctx, query,
		req.ID, req.SenderID, req.ReceiverID, req.Status)

	if err != nil {

		return commonapperr.MapPostgresError(err, "create friendship request", apperr.MapConstraint)
	}

	return nil
}

func (r *friendshipRequestRepository) FindPending(
	ctx context.Context, senderID, receiverID uuid.UUID,
) (*entity.FriendshipRequest, error) {
	query := `
		SELECT id, sender_id, receiver_id, status, created_at, updated_at
		FROM   friendship_requests
		WHERE  sender_id = $1 AND receiver_id = $2 AND status = 'pending'`

	var req entity.FriendshipRequest
	err := r.exec.QueryRowContext(ctx, query, senderID, receiverID).Scan(
		&req.ID, &req.SenderID, &req.ReceiverID, &req.Status, &req.CreatedAt, &req.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {

			return nil, apperr.FriendRequestNotFound()
		}

		return nil, commonapperr.MapPostgresError(err, "find pending request")
	}

	return &req, nil
}

func (r *friendshipRequestRepository) FindPendingBetween(
	ctx context.Context, user1, user2 uuid.UUID,
) (*entity.FriendshipRequest, error) {
	const query = `
		SELECT id, sender_id, receiver_id, status, created_at, updated_at
		FROM   friendship_requests
		WHERE  ((sender_id = $1 AND receiver_id = $2)
		    OR  (sender_id = $2 AND receiver_id = $1))
		  AND  status = 'pending'`

	var req entity.FriendshipRequest
	err := r.exec.QueryRowContext(ctx, query, user1, user2).Scan(
		&req.ID, &req.SenderID, &req.ReceiverID, &req.Status, &req.CreatedAt, &req.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // no pending request - not an error
		}
		return nil, commonapperr.MapPostgresError(err, "find pending between")
	}
	return &req, nil
}

func (r *friendshipRequestRepository) UpdateStatus(
	ctx context.Context, senderID, receiverID uuid.UUID, status entity.FriendshipReqStatus,
) error {
	const query = `
		UPDATE friendship_requests
		SET    status = $3
		WHERE  sender_id = $1 AND receiver_id = $2 AND status = 'pending'`

	result, err := r.exec.ExecContext(ctx, query, senderID, receiverID, status)

	if err != nil {

		return commonapperr.MapPostgresError(err, "update request status")
	}
	rows, _ := result.RowsAffected()

	if rows == 0 {

		return apperr.FriendRequestNotFound()
	}
	return nil
}

func (r *friendshipRequestRepository) Delete(ctx context.Context, senderID, receiverID uuid.UUID) error {
	const query = `
		DELETE FROM friendship_requests
		WHERE  sender_id = $1 AND receiver_id = $2 AND status = 'pending'`

	result, err := r.exec.ExecContext(ctx, query, senderID, receiverID)

	if err != nil {

		return commonapperr.MapPostgresError(err, "delete friendship request")
	}
	rows, _ := result.RowsAffected()

	if rows == 0 {

		return apperr.FriendRequestNotFound()
	}

	return nil
}
