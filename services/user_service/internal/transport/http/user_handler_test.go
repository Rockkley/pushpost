package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	handlerhttp "github.com/rockkley/pushpost/internal/handler/http"
	"github.com/rockkley/pushpost/internal/handler/httperror"
	"github.com/rockkley/pushpost/services/common/apperror"
	domaindto "github.com/rockkley/pushpost/services/user_service/internal/domain/dto"
	"github.com/rockkley/pushpost/services/user_service/internal/entity"
	userhttp "github.com/rockkley/pushpost/services/user_service/internal/transport/http"
	"github.com/stretchr/testify/require"
)

// ── Helpers ───────────────────────────────────────────────────────────────────

func newJSONRequest(t *testing.T, method, path, payload string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func decodeJSONResponse(t *testing.T, resp *httptest.ResponseRecorder, v any) {
	t.Helper()
	require.Equal(t, "application/json", resp.Header().Get("Content-Type"))
	require.NoError(t, json.NewDecoder(resp.Body).Decode(v))
}

// ── Mock ──────────────────────────────────────────────────────────────────────

type mockUserUsecase struct {
	authenticateUserFunc func(ctx context.Context, req domaindto.AuthenticateUserRequestDTO) (*entity.User, error)
	createUserFunc       func(ctx context.Context, req domaindto.CreateUserDTO) (*entity.User, error)
	getUserByEmailFunc   func(ctx context.Context, email string) (*entity.User, error)
}

func (m *mockUserUsecase) AuthenticateUser(ctx context.Context, req domaindto.AuthenticateUserRequestDTO) (*entity.User, error) {
	if m.authenticateUserFunc != nil {
		return m.authenticateUserFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockUserUsecase) CreateUser(ctx context.Context, req domaindto.CreateUserDTO) (*entity.User, error) {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockUserUsecase) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return nil, nil
}

func makeUser(id uuid.UUID, username, email string) *entity.User {
	return &entity.User{
		Id:        id,
		Username:  username,
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ── CreateUser ────────────────────────────────────────────────────────────────

// NOTE: UserHandler.CreateUser does not wrap json.Decode/validation errors as
// AppError → malformed JSON and empty body result in HTTP 500. Tests below
// document this current (buggy) behaviour so regressions are visible.

func TestUserHandler_CreateUser_InvalidJSON(t *testing.T) {
	req := newJSONRequest(t, http.MethodPost, "/user", `{broken`)
	resp := httptest.NewRecorder()

	handlerhttp.MakeHandler(userhttp.NewUserHandler(&mockUserUsecase{}).CreateUser).ServeHTTP(resp, req)

	require.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestUserHandler_CreateUser_EmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/user", nil)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	handlerhttp.MakeHandler(userhttp.NewUserHandler(&mockUserUsecase{}).CreateUser).ServeHTTP(resp, req)

	require.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestUserHandler_CreateUser_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		payload string
	}{
		{name: "username empty", payload: `{"username":"","email":"user@example.com","passwordHash":"hash"}`},
		{name: "email empty", payload: `{"username":"user","email":"","passwordHash":"hash"}`},
		{name: "passwordHash empty", payload: `{"username":"user","email":"user@example.com","passwordHash":""}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newJSONRequest(t, http.MethodPost, "/user", tt.payload)
			resp := httptest.NewRecorder()

			handlerhttp.MakeHandler(userhttp.NewUserHandler(&mockUserUsecase{}).CreateUser).ServeHTTP(resp, req)

			require.Equal(t, http.StatusInternalServerError, resp.Code)
		})
	}
}

func TestUserHandler_CreateUser_Success(t *testing.T) {
	userID := uuid.New()

	// Capture args outside mock to avoid calling require inside the mock func.
	// require inside a mock calls runtime.Goexit which unwinds the handler
	// before it writes a response, making assertions on resp.Code unreliable.
	var gotUsername, gotEmail, gotPasswordHash string
	svc := &mockUserUsecase{
		createUserFunc: func(_ context.Context, req domaindto.CreateUserDTO) (*entity.User, error) {
			gotUsername = req.Username
			gotEmail = req.Email
			gotPasswordHash = req.PasswordHash
			return makeUser(userID, req.Username, req.Email), nil
		},
	}

	req := newJSONRequest(t, http.MethodPost, "/user",
		`{"username":"newuser","email":"new@example.com","passwordHash":"$2a$10$testhash"}`)
	resp := httptest.NewRecorder()

	handlerhttp.MakeHandler(userhttp.NewUserHandler(svc).CreateUser).ServeHTTP(resp, req)

	// Verify usecase was called with correct args.
	require.Equal(t, "newuser", gotUsername)
	require.Equal(t, "new@example.com", gotEmail)
	require.Equal(t, "$2a$10$testhash", gotPasswordHash)

	// Verify response status.
	require.Equal(t, http.StatusCreated, resp.Code)

	// Deserialize directly into entity.User — avoids assumptions about JSON key casing.
	// entity.User has no JSON tags so fields serialize with their Go names ("Id", "Username", …).
	var body entity.User
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, userID, body.Id)
	require.Equal(t, "newuser", body.Username)
	require.Equal(t, "new@example.com", body.Email)
}

func TestUserHandler_CreateUser_ServiceErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{
			name:       "email already exists",
			err:        apperror.EmailAlreadyExists(),
			wantStatus: http.StatusConflict,
			wantCode:   apperror.CodeEmailExists,
		},
		{
			name:       "username already exists",
			err:        apperror.UsernameAlreadyExists(),
			wantStatus: http.StatusConflict,
			wantCode:   apperror.CodeUsernameExists,
		},
		{
			name:       "internal error",
			err:        apperror.Internal("db error", nil),
			wantStatus: http.StatusInternalServerError,
			wantCode:   apperror.CodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockUserUsecase{
				createUserFunc: func(_ context.Context, _ domaindto.CreateUserDTO) (*entity.User, error) {
					return nil, tt.err
				},
			}
			req := newJSONRequest(t, http.MethodPost, "/user",
				`{"username":"user","email":"user@example.com","passwordHash":"$2a$10$hash"}`)
			resp := httptest.NewRecorder()

			handlerhttp.MakeHandler(userhttp.NewUserHandler(svc).CreateUser).ServeHTTP(resp, req)

			var body httperror.ErrorResponse
			decodeJSONResponse(t, resp, &body)
			require.Equal(t, tt.wantStatus, resp.Code)
			require.Equal(t, tt.wantCode, body.Code)
		})
	}
}

// ── AuthenticateUser ──────────────────────────────────────────────────────────

func TestUserHandler_AuthenticateUser_InvalidJSON(t *testing.T) {
	req := newJSONRequest(t, http.MethodPost, "/users/authenticate-user", `{broken`)
	resp := httptest.NewRecorder()

	handlerhttp.MakeHandler(userhttp.NewUserHandler(&mockUserUsecase{}).AuthenticateUser).ServeHTTP(resp, req)

	require.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestUserHandler_AuthenticateUser_Success(t *testing.T) {
	userID := uuid.New()

	var gotEmail, gotPasswordHash string
	svc := &mockUserUsecase{
		authenticateUserFunc: func(_ context.Context, req domaindto.AuthenticateUserRequestDTO) (*entity.User, error) {
			gotEmail = req.Email
			gotPasswordHash = req.PasswordHash
			return makeUser(userID, "testuser", req.Email), nil
		},
	}

	req := newJSONRequest(t, http.MethodPost, "/users/authenticate-user",
		`{"email":"user@example.com","passwordHash":"$2a$10$hash"}`)
	resp := httptest.NewRecorder()

	handlerhttp.MakeHandler(userhttp.NewUserHandler(svc).AuthenticateUser).ServeHTTP(resp, req)

	require.Equal(t, "user@example.com", gotEmail)
	require.Equal(t, "$2a$10$hash", gotPasswordHash)

	require.Equal(t, http.StatusOK, resp.Code)

	var body entity.User
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, userID, body.Id)
}

func TestUserHandler_AuthenticateUser_InvalidCredentials(t *testing.T) {
	svc := &mockUserUsecase{
		authenticateUserFunc: func(_ context.Context, _ domaindto.AuthenticateUserRequestDTO) (*entity.User, error) {
			return nil, apperror.InvalidCredentials()
		},
	}

	req := newJSONRequest(t, http.MethodPost, "/users/authenticate-user",
		`{"email":"user@example.com","passwordHash":"wronghash"}`)
	resp := httptest.NewRecorder()

	handlerhttp.MakeHandler(userhttp.NewUserHandler(svc).AuthenticateUser).ServeHTTP(resp, req)

	var body httperror.ErrorResponse
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, http.StatusUnauthorized, resp.Code)
	require.Equal(t, apperror.CodeInvalidCredentials, body.Code)
}

// ── GetUserByEmail ────────────────────────────────────────────────────────────

func TestUserHandler_GetUserByEmail_MissingQueryParam(t *testing.T) {
	// Missing ?email= → handler returns errors.New (not AppError) → 500.
	req := httptest.NewRequest(http.MethodGet, "/users/by-email", nil)
	resp := httptest.NewRecorder()

	handlerhttp.MakeHandler(userhttp.NewUserHandler(&mockUserUsecase{}).GetUserByEmail).ServeHTTP(resp, req)

	require.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestUserHandler_GetUserByEmail_Success(t *testing.T) {
	userID := uuid.New()

	var gotEmail string
	svc := &mockUserUsecase{
		getUserByEmailFunc: func(_ context.Context, email string) (*entity.User, error) {
			gotEmail = email
			return makeUser(userID, "testuser", email), nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/users/by-email?email=find@example.com", nil)
	resp := httptest.NewRecorder()

	handlerhttp.MakeHandler(userhttp.NewUserHandler(svc).GetUserByEmail).ServeHTTP(resp, req)

	require.Equal(t, "find@example.com", gotEmail)

	require.Equal(t, http.StatusOK, resp.Code)

	var body entity.User
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, userID, body.Id)
}

func TestUserHandler_GetUserByEmail_NotFound(t *testing.T) {
	svc := &mockUserUsecase{
		getUserByEmailFunc: func(_ context.Context, _ string) (*entity.User, error) {
			return nil, apperror.UserNotFound()
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/users/by-email?email=ghost@example.com", nil)
	resp := httptest.NewRecorder()

	handlerhttp.MakeHandler(userhttp.NewUserHandler(svc).GetUserByEmail).ServeHTTP(resp, req)

	var body httperror.ErrorResponse
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, http.StatusNotFound, resp.Code)
	require.Equal(t, apperror.CodeUserNotFound, body.Code)
}

func TestUserHandler_GetUserByEmail_DeletedUser(t *testing.T) {
	svc := &mockUserUsecase{
		getUserByEmailFunc: func(_ context.Context, _ string) (*entity.User, error) {
			return nil, apperror.NotFound(apperror.CodeUserDeleted, "user is deleted")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/users/by-email?email=deleted@example.com", nil)
	resp := httptest.NewRecorder()

	handlerhttp.MakeHandler(userhttp.NewUserHandler(svc).GetUserByEmail).ServeHTTP(resp, req)

	var body httperror.ErrorResponse
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, http.StatusNotFound, resp.Code)
	require.Equal(t, apperror.CodeUserDeleted, body.Code)
}
