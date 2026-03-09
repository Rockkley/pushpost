package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/database"
	"github.com/rockkley/pushpost/services/user_service/internal/apperror"
	"github.com/rockkley/pushpost/services/user_service/internal/entity"
	"strings"
	"time"
)

type UserRepository struct {
	exec database.Executor
}

func NewUserRepository(exec database.Executor) *UserRepository {
	return &UserRepository{exec: exec}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
			INSERT INTO users (id, username, email, password_hash) 
			VALUES ($1,$2,$3,$4) RETURNING created_at, updated_at`

	err := r.exec.QueryRowContext(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash).Scan(&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return commonapperr.MapPostgresError(err, "create user_service", apperror.MapConstraint)
	}

	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, userId uuid.UUID) (*entity.User, error) {
	query := `
			SELECT id, username, email, password_hash, created_at, deleted_at
			FROM users
			WHERE id = $1`

	var user entity.User

	err := r.exec.QueryRowContext(ctx, query, userId).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.DeletedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.UserNotFound()
		}
		return nil, commonapperr.MapPostgresError(err, "find user by id")
	}

	return &user, err
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {

	email = strings.ToLower(strings.TrimSpace(email))
	query := `
			SELECT id, username, email, password_hash, created_at, deleted_at
			FROM users
			WHERE email = $1 AND deleted_at IS NULL`

	var user entity.User

	err := r.exec.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.DeletedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.InvalidCredentials()
		}
		return nil, commonapperr.MapPostgresError(err, "get user by email")
	}

	return &user, err
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	query := `
			SELECT id, username, email, password_hash, created_at, deleted_at
			FROM users
			WHERE username = $1 AND deleted_at IS NULL`

	var user entity.User

	err := r.exec.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.DeletedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.UserNotFound()
		}
		return nil, commonapperr.MapPostgresError(err, "find user by username")
	}

	return &user, err
}

func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool

	query := `
		SELECT EXISTS(
			SELECT 1 FROM users
			WHERE LOWER(email)=LOWER($1) AND deleted_at IS NULL
			)`

	err := r.exec.QueryRowContext(ctx, query, email).Scan(&exists)

	if err != nil {
		return false, commonapperr.MapPostgresError(err, "check email exists")
	}

	return exists, nil
}

func (r *UserRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool

	query := `
		SELECT EXISTS(
			SELECT 1 FROM users
			WHERE LOWER(username)=LOWER($1) AND deleted_at IS NULL
			)`

	err := r.exec.QueryRowContext(ctx, query, username).Scan(&exists)

	if err != nil {
		return false, commonapperr.MapPostgresError(err, "check username exists")
	}

	return exists, nil
}

func (r *UserRepository) SoftDelete(ctx context.Context, userId uuid.UUID) error {
	query := `
		UPDATE users 
		SET deleted_at = $1
		WHERE id=$2 and deleted_at IS NULL`

	result, err := r.exec.ExecContext(ctx, query, time.Now(), userId)

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
	return err
}
