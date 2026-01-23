package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/internal/domain"
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
			INSERT INTO users (id, username, email, password_hash, created_at, deleted_at) 
			VALUES ($1,$2,$3,$4,$5)`

	_, err := r.db.ExecContext(ctx, query,
		user.Id, user.Username, user.Email, user.PasswordHash, user.CreatedAt)

	// TODO error handling
	return err
}

func (r *UserRepository) FindByID(ctx context.Context, userId uuid.UUID) (*domain.User, error) {
	query := `
			SELECT id, username, email, password_hash, created_at, deleted_at
			FROM users
			WHERE id = $1`

	var user domain.User

	err := r.db.QueryRowContext(ctx, query, userId).Scan(
		&user.Id, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.DeletedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	return &user, err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
			SELECT id, username, email, password_hash, created_at, deleted_at
			FROM users
			WHERE LOWER(email) = LOWER($1) AND deleted_at IS NULL`

	var user domain.User

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.Id, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.DeletedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return &user, err
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
			SELECT id, username, email, password_hash, created_at, deleted_at
			FROM users
			WHERE LOWER(username) = LOWER($1) AND deleted_at IS NULL`

	var user domain.User

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.Id, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.DeletedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find user by username: %w", err)
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
		return false, fmt.Errorf("failed to check email exists:%w", err)
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
		return false, fmt.Errorf("failed to check username exists:%w", err)
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
		return fmt.Errorf("failed to soft delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("faied to get rows affected")
	}

	if rows == 0 {
		return domain.ErrUserNotFound
	}
}
