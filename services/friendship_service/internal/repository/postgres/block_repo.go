package postgres

import (
	"context"
	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/database"
	apperr "github.com/rockkley/pushpost/services/friendship_service/internal/apperror"
	"github.com/rockkley/pushpost/services/friendship_service/internal/entity"
)

type blockRepo struct {
	exec database.Executor
}

func NewBlockRepository(exec database.Executor) *blockRepo {
	return &blockRepo{exec: exec}
}

func (r *blockRepo) Create(ctx context.Context, block *entity.Block) error {
	query := `
		INSERT INTO blocks (user_id, target_id)
		VALUES ($1, $2) RETURNING created_at`

	if err := r.exec.QueryRowContext(ctx, query, block.UserID, block.TargetID).Scan(&block.CreatedAt); err != nil {
		return commonapperr.MapPostgresError(err, "create block", apperr.MapConstraint)
	}

	return nil
}

func (r *blockRepo) Delete(ctx context.Context, userID, targetID uuid.UUID) error {
	query := `DELETE FROM blocks WHERE user_id = $1 AND target_id = $2`

	result, err := r.exec.ExecContext(ctx, query, userID, targetID)

	if err != nil {
		return commonapperr.MapPostgresError(err, "delete block")
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return commonapperr.MapPostgresError(err, "delete block rows affected")
	}

	if rowsAffected == 0 {
		return apperr.BlockNotFound()
	}

	return nil
}

func (r *blockRepo) Exists(ctx context.Context, userID, targetID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM blocks WHERE user_id = $1 AND target_id = $2)`

	var exists bool

	if err := r.exec.QueryRowContext(ctx, query, userID, targetID).Scan(&exists); err != nil {
		return false, commonapperr.MapPostgresError(err, "check block existence")
	}

	return exists, nil
}

func (r *blockRepo) GetBlockedUserIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT target_id FROM blocks WHERE user_id = $1`

	rows, err := r.exec.QueryContext(ctx, query, userID)

	if err != nil {
		return nil, commonapperr.MapPostgresError(err, "get blocked user ids")
	}

	defer rows.Close()

	var blockedIDs []uuid.UUID

	for rows.Next() {
		var targetID uuid.UUID
		if err = rows.Scan(&targetID); err != nil {
			return nil, commonapperr.MapPostgresError(err, "scan blocked user id")
		}
		blockedIDs = append(blockedIDs, targetID)
	}

	if err = rows.Err(); err != nil {
		return nil, commonapperr.MapPostgresError(err, "iterate blocked user ids")
	}

	return blockedIDs, nil
}
