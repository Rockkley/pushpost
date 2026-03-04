package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/common/apperror"
	domaindto "github.com/rockkley/pushpost/services/user_service/internal/domain/dto"
	"github.com/rockkley/pushpost/services/user_service/internal/entity"
	"github.com/rockkley/pushpost/services/user_service/internal/repository"
	"github.com/stretchr/testify/require"
)

// ── Mock: repository.UserRepository ──────────────────────────────────────────

type mockUserRepository struct {
	createFunc         func(ctx context.Context, user *entity.User) error
	findByIDFunc       func(ctx context.Context, id uuid.UUID) (*entity.User, error)
	getUserByEmailFunc func(ctx context.Context, email string) (*entity.User, error)
	findByUsernameFunc func(ctx context.Context, username string) (*entity.User, error)
	emailExistsFunc    func(ctx context.Context, email string) (bool, error)
	usernameExistsFunc func(ctx context.Context, username string) (bool, error)
	softDeleteFunc     func(ctx context.Context, id uuid.UUID) error
}

// Verify interface compliance at compile time.
var _ repository.UserRepository = (*mockUserRepository)(nil)

func (m *mockUserRepository) Create(ctx context.Context, user *entity.User) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, user)
	}
	return nil
}

func (m *mockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, apperror.UserNotFound()
}

func (m *mockUserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return nil, apperror.UserNotFound()
}

func (m *mockUserRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	if m.findByUsernameFunc != nil {
		return m.findByUsernameFunc(ctx, username)
	}
	return nil, apperror.UserNotFound()
}

func (m *mockUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	if m.emailExistsFunc != nil {
		return m.emailExistsFunc(ctx, email)
	}
	return false, nil
}

func (m *mockUserRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	if m.usernameExistsFunc != nil {
		return m.usernameExistsFunc(ctx, username)
	}
	return false, nil
}

func (m *mockUserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if m.softDeleteFunc != nil {
		return m.softDeleteFunc(ctx, id)
	}
	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func activeUser(id uuid.UUID) *entity.User {
	return &entity.User{
		Id:           id,
		Username:     "activeuser",
		Email:        "active@example.com",
		PasswordHash: "$2a$04$testhash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func deletedUser(id uuid.UUID) *entity.User {
	u := activeUser(id)
	now := time.Now()
	u.DeletedAt = &now
	return u
}

// ── CreateUser ────────────────────────────────────────────────────────────────

func TestUserUseCase_CreateUser_Success(t *testing.T) {
	var storedUser *entity.User
	repo := &mockUserRepository{
		createFunc: func(_ context.Context, u *entity.User) error {
			storedUser = u
			return nil
		},
	}

	uc := NewUserUseCase(repo)
	result, err := uc.CreateUser(context.Background(), domaindto.CreateUserDTO{
		Username:     "newuser",
		Email:        "new@example.com",
		PasswordHash: "$2a$10$somehash",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotEqual(t, uuid.Nil, result.Id, "UUID must be generated")
	require.Equal(t, "newuser", result.Username)
	require.Equal(t, "new@example.com", result.Email)
	require.Equal(t, "$2a$10$somehash", result.PasswordHash)
	require.NotNil(t, storedUser)
	require.Equal(t, result.Id, storedUser.Id)
}

func TestUserUseCase_CreateUser_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		dto  domaindto.CreateUserDTO
	}{
		{name: "username empty", dto: domaindto.CreateUserDTO{Email: "e@e.com", PasswordHash: "hash"}},
		{name: "email empty", dto: domaindto.CreateUserDTO{Username: "user", PasswordHash: "hash"}},
		{name: "passwordHash empty", dto: domaindto.CreateUserDTO{Username: "user", Email: "e@e.com"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewUserUseCase(&mockUserRepository{})
			_, err := uc.CreateUser(context.Background(), tt.dto)
			require.Error(t, err)
		})
	}
}

func TestUserUseCase_CreateUser_RepositoryError_EmailExists(t *testing.T) {
	repo := &mockUserRepository{
		createFunc: func(_ context.Context, _ *entity.User) error {
			return apperror.EmailAlreadyExists()
		},
	}

	uc := NewUserUseCase(repo)
	_, err := uc.CreateUser(context.Background(), domaindto.CreateUserDTO{
		Username:     "user",
		Email:        "taken@example.com",
		PasswordHash: "$2a$10$hash",
	})

	require.Error(t, err)
	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, apperror.CodeEmailExists, appErr.Code())
	require.Equal(t, 409, appErr.HTTPStatus())
}

func TestUserUseCase_CreateUser_RepositoryError_UsernameExists(t *testing.T) {
	repo := &mockUserRepository{
		createFunc: func(_ context.Context, _ *entity.User) error {
			return apperror.UsernameAlreadyExists()
		},
	}

	uc := NewUserUseCase(repo)
	_, err := uc.CreateUser(context.Background(), domaindto.CreateUserDTO{
		Username:     "taken",
		Email:        "e@example.com",
		PasswordHash: "$2a$10$hash",
	})

	require.Error(t, err)
	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, apperror.CodeUsernameExists, appErr.Code())
}

func TestUserUseCase_CreateUser_RepositoryError_DatabaseError(t *testing.T) {
	repo := &mockUserRepository{
		createFunc: func(_ context.Context, _ *entity.User) error {
			return apperror.Database("connection error", errors.New("pg: connection refused"))
		},
	}

	uc := NewUserUseCase(repo)
	_, err := uc.CreateUser(context.Background(), domaindto.CreateUserDTO{
		Username:     "user",
		Email:        "u@example.com",
		PasswordHash: "$2a$10$hash",
	})

	require.Error(t, err)
	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, apperror.CodeDatabaseError, appErr.Code())
}

func TestUserUseCase_CreateUser_GeneratesNewUUID(t *testing.T) {
	var ids []uuid.UUID
	repo := &mockUserRepository{
		createFunc: func(_ context.Context, u *entity.User) error {
			ids = append(ids, u.Id)
			return nil
		},
	}

	uc := NewUserUseCase(repo)
	for i := 0; i < 3; i++ {
		_, err := uc.CreateUser(context.Background(), domaindto.CreateUserDTO{
			Username:     "user",
			Email:        "u@example.com",
			PasswordHash: "$2a$10$hash",
		})
		require.NoError(t, err)
	}

	require.Len(t, ids, 3)
	require.NotEqual(t, ids[0], ids[1])
	require.NotEqual(t, ids[1], ids[2])
}

// ── AuthenticateUser ──────────────────────────────────────────────────────────

func TestUserUseCase_AuthenticateUser_Success(t *testing.T) {
	id := uuid.New()
	repo := &mockUserRepository{
		getUserByEmailFunc: func(_ context.Context, email string) (*entity.User, error) {
			require.Equal(t, "user@example.com", email)
			return activeUser(id), nil
		},
	}

	uc := NewUserUseCase(repo)
	result, err := uc.AuthenticateUser(context.Background(), domaindto.AuthenticateUserRequestDTO{
		Email:        "user@example.com",
		PasswordHash: "$2a$10$hash",
	})

	require.NoError(t, err)
	require.Equal(t, id, result.Id)
}

func TestUserUseCase_AuthenticateUser_UserNotFound(t *testing.T) {
	repo := &mockUserRepository{
		getUserByEmailFunc: func(_ context.Context, _ string) (*entity.User, error) {
			return nil, apperror.InvalidCredentials()
		},
	}

	uc := NewUserUseCase(repo)
	_, err := uc.AuthenticateUser(context.Background(), domaindto.AuthenticateUserRequestDTO{
		Email:        "ghost@example.com",
		PasswordHash: "$2a$10$hash",
	})

	require.Error(t, err)
	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, apperror.CodeInvalidCredentials, appErr.Code())
	require.Equal(t, 401, appErr.HTTPStatus())
}

func TestUserUseCase_AuthenticateUser_DeletedUser(t *testing.T) {
	id := uuid.New()
	repo := &mockUserRepository{
		getUserByEmailFunc: func(_ context.Context, _ string) (*entity.User, error) {
			return deletedUser(id), nil
		},
	}

	uc := NewUserUseCase(repo)
	_, err := uc.AuthenticateUser(context.Background(), domaindto.AuthenticateUserRequestDTO{
		Email:        "deleted@example.com",
		PasswordHash: "$2a$10$hash",
	})

	require.Error(t, err)
	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	// Deleted users must not reveal account existence → InvalidCredentials.
	require.Equal(t, apperror.CodeInvalidCredentials, appErr.Code())
	require.Equal(t, 401, appErr.HTTPStatus())
}

// ── GetUserByEmail ────────────────────────────────────────────────────────────

func TestUserUseCase_GetUserByEmail_Success(t *testing.T) {
	id := uuid.New()
	repo := &mockUserRepository{
		getUserByEmailFunc: func(_ context.Context, email string) (*entity.User, error) {
			require.Equal(t, "user@example.com", email)
			return activeUser(id), nil
		},
	}

	uc := NewUserUseCase(repo)
	result, err := uc.GetUserByEmail(context.Background(), "user@example.com")

	require.NoError(t, err)
	require.Equal(t, id, result.Id)
}

func TestUserUseCase_GetUserByEmail_NotFound(t *testing.T) {
	repo := &mockUserRepository{
		getUserByEmailFunc: func(_ context.Context, _ string) (*entity.User, error) {
			return nil, apperror.UserNotFound()
		},
	}

	uc := NewUserUseCase(repo)
	_, err := uc.GetUserByEmail(context.Background(), "nobody@example.com")

	require.Error(t, err)
	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, apperror.CodeUserNotFound, appErr.Code())
	require.Equal(t, 404, appErr.HTTPStatus())
}

func TestUserUseCase_GetUserByEmail_DeletedUser(t *testing.T) {
	id := uuid.New()
	repo := &mockUserRepository{
		getUserByEmailFunc: func(_ context.Context, _ string) (*entity.User, error) {
			return deletedUser(id), nil
		},
	}

	uc := NewUserUseCase(repo)
	_, err := uc.GetUserByEmail(context.Background(), "deleted@example.com")

	require.Error(t, err)
	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, apperror.CodeUserDeleted, appErr.Code())
	require.Equal(t, 404, appErr.HTTPStatus())
}

func TestUserUseCase_GetUserByEmail_DatabaseError(t *testing.T) {
	repo := &mockUserRepository{
		getUserByEmailFunc: func(_ context.Context, _ string) (*entity.User, error) {
			return nil, apperror.Database("db unavailable", errors.New("timeout"))
		},
	}

	uc := NewUserUseCase(repo)
	_, err := uc.GetUserByEmail(context.Background(), "any@example.com")

	// GetUserByEmail wraps any repo error as UserNotFound.
	require.Error(t, err)
	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, 404, appErr.HTTPStatus())
}
