package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/internal/apperror"
	"github.com/rockkley/pushpost/internal/domain"
	"strings"
	"time"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
			INSERT INTO users (id, username, email, password_hash) 
			VALUES ($1,$2,$3,$4) RETURNING created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		user.Id, user.Username, user.Email, user.PasswordHash).Scan(&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return apperror.MapPostgresError(err, "create user")
	}

	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, userId uuid.UUID) (*domain.User, error) {
	query := `
			SELECT id, username, email, password_hash, created_at, deleted_at
			FROM users
			WHERE id = $1`

	var user domain.User

	err := r.db.QueryRowContext(ctx, query, userId).Scan(
		&user.Id, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.DeletedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.UserNotFound()
		}
		return nil, apperror.MapPostgresError(err, "find user by id")
	}

	return &user, err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	query := `
			SELECT id, username, email, password_hash, created_at, deleted_at
			FROM users
			WHERE email = $1 AND deleted_at IS NULL`

	var user domain.User

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.Id, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.DeletedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.InvalidCredentials()
		}
		return nil, apperror.MapPostgresError(err, "find user by email")
	}

	return &user, err
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	query := `
			SELECT id, username, email, password_hash, created_at, deleted_at
			FROM users
			WHERE username = $1 AND deleted_at IS NULL`

	var user domain.User

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.Id, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.DeletedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.UserNotFound()
		}
		return nil, apperror.MapPostgresError(err, "find user by username")
	}

	return &user, err
}

func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM users
			WHERE LOWER(email)=LOWER($1) AND deleted_at IS NULL
			)`

	var exists bool

	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)

	if err != nil {
		return false, apperror.MapPostgresError(err, "check email exists")
	}

	return exists, nil
}

func (r *UserRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM users
			WHERE LOWER(username)=LOWER($1) AND deleted_at IS NULL
			)`

	var exists bool

	err := r.db.QueryRowContext(ctx, query, username).Scan(&exists)

	if err != nil {
		return false, apperror.MapPostgresError(err, "check username exists")
	}

	return exists, nil
}

func (r *UserRepository) SoftDelete(ctx context.Context, userId uuid.UUID) error {
	query := `
		UPDATE users 
		SET deleted_at = $1
		WHERE id=$2 and deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, time.Now(), userId)

	if err != nil {
		return apperror.MapPostgresError(err, "soft delete user")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperror.Database("failed to get rows affected", err)
	}

	if rows == 0 {
		return apperror.UserNotFound()
	}
	return err
}
