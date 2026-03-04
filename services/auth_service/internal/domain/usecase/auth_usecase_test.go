package usecase

import (
	"context"
	"errors"
	user_api2 "github.com/rockkley/pushpost/clients/user_api"
	jwtpkg "github.com/rockkley/pushpost/services/common/jwt"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/rockkley/pushpost/services/auth_service/internal/domain"
	authrep "github.com/rockkley/pushpost/services/auth_service/internal/repository"
	"github.com/rockkley/pushpost/services/auth_service/internal/transport/http/dto"
	"github.com/rockkley/pushpost/services/common/apperror"
	"github.com/stretchr/testify/require"
)

// ── Helpers ───────────────────────────────────────────────────────────────────

const testJWTSecret = "test-secret-that-is-exactly-32-chars!!"

// lowCostHash generates a bcrypt hash at MinCost (4) for fast test execution.
func lowCostHash(t testing.TB, password string) string {
	t.Helper()
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)
	return string(b)
}

func newTestUsecase(client user_api2.Client, store authrep.SessionStore) *AuthUsecase {
	jm := jwtpkg.NewManager(testJWTSecret, nil)
	return NewAuthUsecase(client, store, jm)
}

// ── Mock: user_api.Client ──────────────────────────────────────────────────────

type mockUserClient struct {
	createUserFunc       func(ctx context.Context, req user_api2.CreateUserRequest) (*user_api2.UserResponse, error)
	getUserByIDFunc      func(ctx context.Context, id uuid.UUID) (*user_api2.UserResponse, error)
	getUserByEmailFunc   func(ctx context.Context, email string) (*user_api2.UserResponse, error)
	authenticateUserFunc func(ctx context.Context, email, password string) (*user_api2.UserResponse, error)
}

func (m *mockUserClient) CreateUser(ctx context.Context, req user_api2.CreateUserRequest) (*user_api2.UserResponse, error) {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockUserClient) GetUserByID(ctx context.Context, id uuid.UUID) (*user_api2.UserResponse, error) {
	if m.getUserByIDFunc != nil {
		return m.getUserByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockUserClient) GetUserByEmail(ctx context.Context, email string) (*user_api2.UserResponse, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *mockUserClient) AuthenticateUser(ctx context.Context, email, password string) (*user_api2.UserResponse, error) {
	if m.authenticateUserFunc != nil {
		return m.authenticateUserFunc(ctx, email, password)
	}
	return nil, nil
}

// ── Mock: repository.SessionStore ─────────────────────────────────────────────

type mockSessionStore struct {
	saveFunc   func(ctx context.Context, session *domain.Session) error
	getFunc    func(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error)
	deleteFunc func(ctx context.Context, sessionID uuid.UUID) error
}

func (m *mockSessionStore) Save(ctx context.Context, session *domain.Session) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, session)
	}
	return nil
}

func (m *mockSessionStore) Get(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, sessionID)
	}
	return nil, errors.New("session not found")
}

func (m *mockSessionStore) Delete(ctx context.Context, sessionID uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, sessionID)
	}
	return nil
}

// ── Register ──────────────────────────────────────────────────────────────────

func TestAuthUsecase_Register_Success(t *testing.T) {
	userID := uuid.New()
	client := &mockUserClient{
		createUserFunc: func(_ context.Context, req user_api2.CreateUserRequest) (*user_api2.UserResponse, error) {
			require.Equal(t, "alice", req.Username)
			require.Equal(t, "alice@example.com", req.Email)
			// Password must be hashed, not plain-text.
			require.NotEqual(t, "Password1", req.PasswordHash)
			require.NoError(t, bcrypt.CompareHashAndPassword([]byte(req.PasswordHash), []byte("Password1")))
			return &user_api2.UserResponse{ID: userID, Username: "alice", Email: "alice@example.com"}, nil
		},
	}

	uc := newTestUsecase(client, &mockSessionStore{})
	resp, err := uc.Register(context.Background(), dto.RegisterUserDto{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "Password1",
	})

	require.NoError(t, err)
	require.Equal(t, userID, resp.ID)
	require.Equal(t, "alice", resp.Username)
	require.Equal(t, "alice@example.com", resp.Email)
}

func TestAuthUsecase_Register_UserClientError(t *testing.T) {
	client := &mockUserClient{
		createUserFunc: func(_ context.Context, _ user_api2.CreateUserRequest) (*user_api2.UserResponse, error) {
			return nil, apperror.EmailAlreadyExists()
		},
	}

	uc := newTestUsecase(client, &mockSessionStore{})
	_, err := uc.Register(context.Background(), dto.RegisterUserDto{
		Username: "bob",
		Email:    "bob@example.com",
		Password: "Password1",
	})

	require.Error(t, err)
	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, apperror.CodeEmailExists, appErr.Code())
}

func TestAuthUsecase_Register_PasswordIsHashedBeforeSend(t *testing.T) {
	const plainPwd = "MySecret99"
	var capturedHash string
	client := &mockUserClient{
		createUserFunc: func(_ context.Context, req user_api2.CreateUserRequest) (*user_api2.UserResponse, error) {
			capturedHash = req.PasswordHash
			return &user_api2.UserResponse{ID: uuid.New(), Username: "u", Email: "u@e.com"}, nil
		},
	}

	uc := newTestUsecase(client, &mockSessionStore{})
	_, err := uc.Register(context.Background(), dto.RegisterUserDto{
		Username: "validuser",
		Email:    "u@e.com",
		Password: plainPwd,
	})

	require.NoError(t, err)
	require.NotEqual(t, plainPwd, capturedHash, "plain password must not be sent to user service")
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(capturedHash), []byte(plainPwd)))
}

// ── Login ─────────────────────────────────────────────────────────────────────

func TestAuthUsecase_Login_Success(t *testing.T) {
	userID := uuid.New()
	deviceID := uuid.New()
	hash := lowCostHash(t, "Password1")

	client := &mockUserClient{
		getUserByEmailFunc: func(_ context.Context, email string) (*user_api2.UserResponse, error) {
			require.Equal(t, "user@example.com", email)
			return &user_api2.UserResponse{ID: userID, Username: "user", Email: email, PasswordHash: hash}, nil
		},
	}
	store := &mockSessionStore{}

	uc := newTestUsecase(client, store)
	token, err := uc.Login(context.Background(), dto.LoginUserDTO{
		Email:    "user@example.com",
		Password: "Password1",
		DeviceID: deviceID,
	})

	require.NoError(t, err)
	require.NotEmpty(t, token)
}

func TestAuthUsecase_Login_GeneratesDeviceIDWhenNil(t *testing.T) {
	hash := lowCostHash(t, "Password1")
	client := &mockUserClient{
		getUserByEmailFunc: func(_ context.Context, _ string) (*user_api2.UserResponse, error) {
			return &user_api2.UserResponse{ID: uuid.New(), PasswordHash: hash}, nil
		},
	}

	var savedSession *domain.Session
	store := &mockSessionStore{
		saveFunc: func(_ context.Context, s *domain.Session) error {
			savedSession = s
			return nil
		},
	}

	uc := newTestUsecase(client, store)
	_, err := uc.Login(context.Background(), dto.LoginUserDTO{
		Email:    "u@e.com",
		Password: "Password1",
		DeviceID: uuid.Nil, // not provided
	})

	require.NoError(t, err)
	require.NotNil(t, savedSession)
	require.NotEqual(t, uuid.Nil, savedSession.DeviceID, "DeviceID must be generated when not provided")
}

func TestAuthUsecase_Login_UserNotFound(t *testing.T) {
	client := &mockUserClient{
		getUserByEmailFunc: func(_ context.Context, _ string) (*user_api2.UserResponse, error) {
			return nil, errors.New("user not found")
		},
	}

	uc := newTestUsecase(client, &mockSessionStore{})
	_, err := uc.Login(context.Background(), dto.LoginUserDTO{Email: "x@x.com", Password: "Password1"})

	require.Error(t, err)
	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, apperror.CodeInvalidCredentials, appErr.Code())
	require.Equal(t, 401, appErr.HTTPStatus())
}

func TestAuthUsecase_Login_WrongPassword(t *testing.T) {
	hash := lowCostHash(t, "CorrectPass1")
	client := &mockUserClient{
		getUserByEmailFunc: func(_ context.Context, _ string) (*user_api2.UserResponse, error) {
			return &user_api2.UserResponse{ID: uuid.New(), PasswordHash: hash}, nil
		},
	}

	uc := newTestUsecase(client, &mockSessionStore{})
	_, err := uc.Login(context.Background(), dto.LoginUserDTO{Email: "u@e.com", Password: "WrongPass1"})

	require.Error(t, err)
	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, apperror.CodeInvalidCredentials, appErr.Code())
}

func TestAuthUsecase_Login_SessionSaveFails(t *testing.T) {
	hash := lowCostHash(t, "Password1")
	client := &mockUserClient{
		getUserByEmailFunc: func(_ context.Context, _ string) (*user_api2.UserResponse, error) {
			return &user_api2.UserResponse{ID: uuid.New(), PasswordHash: hash}, nil
		},
	}
	store := &mockSessionStore{
		saveFunc: func(_ context.Context, _ *domain.Session) error {
			return errors.New("redis down")
		},
	}

	uc := newTestUsecase(client, store)
	_, err := uc.Login(context.Background(), dto.LoginUserDTO{Email: "u@e.com", Password: "Password1"})

	require.Error(t, err)
	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, apperror.CodeInternalError, appErr.Code())
	require.Equal(t, 500, appErr.HTTPStatus())
}

func TestAuthUsecase_Login_TokenWrittenToResponse(t *testing.T) {
	hash := lowCostHash(t, "Password1")
	userID := uuid.New()
	client := &mockUserClient{
		getUserByEmailFunc: func(_ context.Context, _ string) (*user_api2.UserResponse, error) {
			return &user_api2.UserResponse{ID: userID, PasswordHash: hash}, nil
		},
	}

	uc := newTestUsecase(client, &mockSessionStore{})
	token, err := uc.Login(context.Background(), dto.LoginUserDTO{Email: "u@e.com", Password: "Password1"})

	require.NoError(t, err)
	// Token is a non-empty JWT (three dot-separated segments).
	require.Contains(t, token, ".")
	parts := len(splitDots(token))
	require.Equal(t, 3, parts, "JWT must have 3 parts separated by '.'")
}

func splitDots(s string) []string {
	var parts []string
	start := 0
	for i, c := range s {
		if c == '.' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

// ── Logout ────────────────────────────────────────────────────────────────────

func TestAuthUsecase_Logout_Success(t *testing.T) {
	sessionID := uuid.New()
	userID := uuid.New()
	deleteCalled := false

	store := &mockSessionStore{
		getFunc: func(_ context.Context, id uuid.UUID) (*domain.Session, error) {
			require.Equal(t, sessionID, id)
			return &domain.Session{SessionID: sessionID, UserID: userID, Expires: time.Now().Add(time.Hour).Unix()}, nil
		},
		deleteFunc: func(_ context.Context, id uuid.UUID) error {
			deleteCalled = true
			require.Equal(t, sessionID, id)
			return nil
		},
	}

	uc := newTestUsecase(&mockUserClient{}, store)
	err := uc.Logout(context.Background(), sessionID)

	require.NoError(t, err)
	require.True(t, deleteCalled)
}

func TestAuthUsecase_Logout_SessionNotFound_ReturnsNil(t *testing.T) {
	// If session is not found, Logout treats it as already logged out → no error.
	store := &mockSessionStore{
		getFunc: func(_ context.Context, _ uuid.UUID) (*domain.Session, error) {
			return nil, errors.New("session not found")
		},
	}

	uc := newTestUsecase(&mockUserClient{}, store)
	err := uc.Logout(context.Background(), uuid.New())

	require.NoError(t, err)
}

func TestAuthUsecase_Logout_DeleteFails(t *testing.T) {
	sessionID := uuid.New()
	store := &mockSessionStore{
		getFunc: func(_ context.Context, _ uuid.UUID) (*domain.Session, error) {
			return &domain.Session{SessionID: sessionID, Expires: time.Now().Add(time.Hour).Unix()}, nil
		},
		deleteFunc: func(_ context.Context, _ uuid.UUID) error {
			return errors.New("redis unreachable")
		},
	}

	uc := newTestUsecase(&mockUserClient{}, store)
	err := uc.Logout(context.Background(), sessionID)

	require.Error(t, err)
	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, apperror.CodeInternalError, appErr.Code())
}

// ── AuthenticateRequest ───────────────────────────────────────────────────────

func TestAuthUsecase_AuthenticateRequest_Success(t *testing.T) {
	jm := jwtpkg.NewManager(testJWTSecret, nil)
	userID := uuid.New()
	deviceID := uuid.New()
	sessionID := uuid.New()

	token, err := jm.Generate(userID, deviceID, sessionID)
	require.NoError(t, err)

	store := &mockSessionStore{
		getFunc: func(_ context.Context, id uuid.UUID) (*domain.Session, error) {
			require.Equal(t, sessionID, id)
			return &domain.Session{
				SessionID: sessionID,
				UserID:    userID,
				DeviceID:  deviceID,
				Expires:   time.Now().Add(time.Hour).Unix(),
			}, nil
		},
	}

	uc := newTestUsecase(&mockUserClient{}, store)
	session, err := uc.AuthenticateRequest(context.Background(), token)

	require.NoError(t, err)
	require.Equal(t, sessionID, session.SessionID)
	require.Equal(t, userID, session.UserID)
}

func TestAuthUsecase_AuthenticateRequest_InvalidToken(t *testing.T) {
	uc := newTestUsecase(&mockUserClient{}, &mockSessionStore{})
	_, err := uc.AuthenticateRequest(context.Background(), "not.a.valid.jwt")

	require.Error(t, err)
	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, 401, appErr.HTTPStatus())
}

func TestAuthUsecase_AuthenticateRequest_EmptyToken(t *testing.T) {
	uc := newTestUsecase(&mockUserClient{}, &mockSessionStore{})
	_, err := uc.AuthenticateRequest(context.Background(), "")

	require.Error(t, err)
	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, 401, appErr.HTTPStatus())
}

func TestAuthUsecase_AuthenticateRequest_TokenSignedWithWrongSecret(t *testing.T) {
	otherJM := jwtpkg.NewManager("completely-different-secret-value!!", nil)
	token, err := otherJM.Generate(uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)

	uc := newTestUsecase(&mockUserClient{}, &mockSessionStore{})
	_, err = uc.AuthenticateRequest(context.Background(), token)

	require.Error(t, err)
	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, 401, appErr.HTTPStatus())
}

func TestAuthUsecase_AuthenticateRequest_SessionNotFound(t *testing.T) {
	jm := jwtpkg.NewManager(testJWTSecret, nil)
	token, err := jm.Generate(uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)

	store := &mockSessionStore{
		getFunc: func(_ context.Context, _ uuid.UUID) (*domain.Session, error) {
			return nil, errors.New("session not found")
		},
	}

	uc := newTestUsecase(&mockUserClient{}, store)
	_, err = uc.AuthenticateRequest(context.Background(), token)

	require.Error(t, err)
	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, 401, appErr.HTTPStatus())
}

func TestAuthUsecase_AuthenticateRequest_ExpiredSession(t *testing.T) {
	jm := jwtpkg.NewManager(testJWTSecret, nil)
	userID := uuid.New()
	deviceID := uuid.New()
	sessionID := uuid.New()
	token, err := jm.Generate(userID, deviceID, sessionID)
	require.NoError(t, err)

	deleteCalled := false
	store := &mockSessionStore{
		getFunc: func(_ context.Context, _ uuid.UUID) (*domain.Session, error) {
			return &domain.Session{
				SessionID: sessionID,
				UserID:    userID,
				Expires:   time.Now().Add(-time.Hour).Unix(), // already expired
			}, nil
		},
		deleteFunc: func(_ context.Context, id uuid.UUID) error {
			deleteCalled = true
			require.Equal(t, sessionID, id)
			return nil
		},
	}

	uc := newTestUsecase(&mockUserClient{}, store)
	_, err = uc.AuthenticateRequest(context.Background(), token)

	require.Error(t, err)
	require.True(t, deleteCalled, "expired session must be deleted")
	var appErr apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, apperror.CodeSessionExpired, appErr.Code())
}
