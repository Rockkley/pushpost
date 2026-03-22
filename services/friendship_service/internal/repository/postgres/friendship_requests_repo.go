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
	"log"
	"log/slog"
	"reflect"
	"time"
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
	slog.Debug("friendshipRequestRepository UpdateStatus")
	log.Println(status, reflect.TypeOf(status))
	query := `
		UPDATE friendship_requests
		SET    status = $3
		WHERE  sender_id = $1 AND receiver_id = $2 AND status = 'pending'`

	result, err := r.exec.ExecContext(ctx, query, senderID, receiverID, status)

	if err != nil {
		slog.Debug("UPDATE REQUEST STATUS FUCKED UP", err.Error())
		return commonapperr.MapPostgresError(err, "update request status")
	}
	rows, _ := result.RowsAffected()

	if rows == 0 {
		slog.Debug("ROWS 0")
		return apperr.FriendRequestNotFound()
	}
	slog.Debug("friendshipRequestRepository UpdateStatus DONE")
	return nil
}

func (r *friendshipRequestRepository) HasRecentRejected(
	ctx context.Context,
	senderID, receiverID uuid.UUID,
	since time.Time,
) (bool, error) {
	query := `
        SELECT EXISTS (
            SELECT 1
            FROM   friendship_requests
            WHERE  sender_id   = $1
              AND  receiver_id = $2
              AND  status      = 'rejected'
              AND  updated_at  > $3
        )`

	var exists bool
	err := r.exec.QueryRowContext(ctx, query, senderID, receiverID, since).Scan(&exists)
	if err != nil {
		return false, commonapperr.MapPostgresError(err, "check cooldown")
	}
	return exists, nil
}

func (r *friendshipRequestRepository) GetIncoming(
	ctx context.Context,
	receiverID uuid.UUID,
) ([]*entity.FriendshipRequest, error) {
	query := `
		SELECT id, sender_id, receiver_id, status, created_at, updated_at
		FROM   friendship_requests
		WHERE  receiver_id = $1
		  AND  status      = 'pending'
		ORDER BY created_at DESC`

	rows, err := r.exec.QueryContext(ctx, query, receiverID)
	if err != nil {

		return nil, commonapperr.MapPostgresError(err, "get incoming requests")
	}
	defer rows.Close()

	var result []*entity.FriendshipRequest
	for rows.Next() {
		var req entity.FriendshipRequest
		if err = rows.Scan(
			&req.ID, &req.SenderID, &req.ReceiverID,
			&req.Status, &req.CreatedAt, &req.UpdatedAt,
		); err != nil {

			return nil, commonapperr.MapPostgresError(err, "scan incoming request")
		}
		result = append(result, &req)
	}
	if err = rows.Err(); err != nil {

		return nil, commonapperr.MapPostgresError(err, "iterate incoming requests")
	}
	return result, nil
}
