package postgres

import (
	"bytes"
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

type friendshipRepo struct {
	exec database.Executor
}

func NewFriendshipRepository(exec database.Executor) repository.FriendshipRepository {
	return &friendshipRepo{exec: exec}
}

func (r *friendshipRepo) Create(ctx context.Context, friendship *entity.Friendship) error {
	u1, u2 := orderUUIDs(friendship.User1ID, friendship.User2ID)

	query := `INSERT INTO friendships (id, user1_id, user2_id) VALUES ($1,$2,$3) RETURNING created_at`

	err := r.exec.QueryRowContext(ctx, query, friendship.ID, u1, u2).Scan(&friendship.CreatedAt)

	if err != nil {

		return commonapperr.MapPostgresError(err, "create friendship", apperr.MapConstraint)
	}

	friendship.User1ID, friendship.User2ID = u1, u2

	return nil
}

func (r *friendshipRepo) Exists(ctx context.Context, user1, user2 uuid.UUID) (bool, error) {
	u1, u2 := orderUUIDs(user1, user2)

	query := `SELECT EXISTS (SELECT 1 FROM friendships WHERE user1_id = $1 AND user2_id = $2)`

	var exists bool

	err := r.exec.QueryRowContext(ctx, query, u1, u2).Scan(&exists)

	if err != nil {

		return false, commonapperr.MapPostgresError(err, "check friendship existence")
	}

	return exists, nil
}

func (r *friendshipRepo) Delete(ctx context.Context, userID, user2 uuid.UUID) error {
	u1, u2 := orderUUIDs(userID, user2)

	query := `DELETE FROM friendships WHERE user1_id = $1 AND user2_id = $2`

	result, err := r.exec.ExecContext(ctx, query, u1, u2)

	if err != nil {

		return commonapperr.MapPostgresError(err, "delete friendship")
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {

		return commonapperr.MapPostgresError(err, "delete friendship rows affected")
	}

	if rowsAffected == 0 {

		return apperr.NotFriends()
	}

	return nil
}

func (r *friendshipRepo) GetFriendIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT CASE WHEN user1_id=$1 THEN user2_id ELSE user1_id END AS friend_id
	FROM friendships WHERE user1_id=$1 OR user2_id=$1
	ORDER BY created_at DESC`

	rows, err := r.exec.QueryContext(ctx, query, userID)

	if err != nil {

		return nil, commonapperr.MapPostgresError(err, "get friend ids")
	}

	defer rows.Close()

	var ids []uuid.UUID

	for rows.Next() {
		var id uuid.UUID

		if scanErr := rows.Scan(&id); scanErr != nil {

			return nil, commonapperr.MapPostgresError(scanErr, "scan friend id")
		}

		ids = append(ids, id)
	}

	if err = rows.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {

			return nil, nil
		}

		return nil, commonapperr.MapPostgresError(err, "iterate friend ids")
	}

	return ids, nil
}

func orderUUIDs(a, b uuid.UUID) (uuid.UUID, uuid.UUID) {
	if bytes.Compare(a[:], b[:]) < 0 {
		return a, b
	}
	return b, a
}
