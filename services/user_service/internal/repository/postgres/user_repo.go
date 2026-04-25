package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/database"
	"github.com/rockkley/pushpost/services/user_service/internal/apperror"
	"github.com/rockkley/pushpost/services/user_service/internal/entity"
)

type UserRepository struct {
	exec database.Executor
}

func NewUserRepository(exec database.Executor) *UserRepository {
	return &UserRepository{exec: exec}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	const query = `
		INSERT INTO users (id, username, email, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at`

	err := r.exec.QueryRowContext(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return commonapperr.MapPostgresError(err, "create user", apperror.MapConstraint)
	}
	user.Status = entity.StatusInactive
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	const query = `
		SELECT id, username, email, password_hash, status, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1`

	var u entity.User
	err := r.exec.QueryRowContext(ctx, query, userID).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash,
		&u.Status, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.UserNotFound()
		}
		return nil, commonapperr.MapPostgresError(err, "find user by id")
	}
	return &u, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	const query = `
		SELECT id, username, email, password_hash, status, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL`

	var u entity.User
	err := r.exec.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash,
		&u.Status, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.InvalidCredentials()
		}
		return nil, commonapperr.MapPostgresError(err, "get user by email")
	}
	return &u, nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	username = strings.TrimSpace(username)
	const query = `
		SELECT id, username, email, password_hash, status, created_at, updated_at, deleted_at
		FROM users
		WHERE LOWER(username) = LOWER($1) AND deleted_at IS NULL`

	var u entity.User
	err := r.exec.QueryRowContext(ctx, query, username).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash,
		&u.Status, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.UserNotFound()
		}
		return nil, commonapperr.MapPostgresError(err, "find user by username")
	}
	return &u, nil
}

func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1 FROM users
			WHERE LOWER(email) = LOWER($1) AND deleted_at IS NULL
		)`
	var exists bool
	if err := r.exec.QueryRowContext(ctx, query, email).Scan(&exists); err != nil {
		return false, commonapperr.MapPostgresError(err, "check email exists")
	}
	return exists, nil
}

func (r *UserRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1 FROM users
			WHERE LOWER(username) = LOWER($1) AND deleted_at IS NULL
		)`
	var exists bool
	if err := r.exec.QueryRowContext(ctx, query, username).Scan(&exists); err != nil {
		return false, commonapperr.MapPostgresError(err, "check username exists")
	}
	return exists, nil
}

func (r *UserRepository) SoftDelete(ctx context.Context, userID uuid.UUID) error {
	const query = `
		UPDATE users SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.exec.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return commonapperr.MapPostgresError(err, "soft delete user")
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return commonapperr.Internal("failed to get rows affected", err)
	}
	if rows == 0 {
		return apperror.UserNotFound()
	}
	return nil
}

func (r *UserRepository) ActivateUser(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	const query = `
		UPDATE users
		SET    status = 'active', updated_at = NOW()
		WHERE  email = $1 AND status = 'inactive' AND deleted_at IS NULL`

	result, err := r.exec.ExecContext(ctx, query, email)
	if err != nil {
		return commonapperr.MapPostgresError(err, "activate user")
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return commonapperr.Internal("failed to get rows affected", err)
	}
	if rows == 0 {
		// Пользователь не найден или уже активен — оба случая безопасны
		return apperror.UserNotFound()
	}
	return nil
}
