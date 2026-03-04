package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	handlerhttp "github.com/rockkley/pushpost/services/common/http"
	"github.com/rockkley/pushpost/services/common/httperror"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/auth_service/internal/domain"
	"github.com/rockkley/pushpost/services/auth_service/internal/transport/http/dto"
	"github.com/rockkley/pushpost/services/auth_service/internal/transport/http/middleware"
	"github.com/rockkley/pushpost/services/common/apperror"
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

type mockAuthUsecase struct {
	registerFunc            func(ctx context.Context, req dto.RegisterUserDto) (*dto.RegisterResponseDto, error)
	loginFunc               func(ctx context.Context, req dto.LoginUserDTO) (string, error)
	logoutFunc              func(ctx context.Context, tokenID uuid.UUID) error
	authenticateRequestFunc func(ctx context.Context, tokenStr string) (*domain.Session, error)
}

func (m *mockAuthUsecase) Register(ctx context.Context, req dto.RegisterUserDto) (*dto.RegisterResponseDto, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockAuthUsecase) Login(ctx context.Context, req dto.LoginUserDTO) (string, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, req)
	}
	return "", nil
}

func (m *mockAuthUsecase) Logout(ctx context.Context, tokenID uuid.UUID) error {
	if m.logoutFunc != nil {
		return m.logoutFunc(ctx, tokenID)
	}
	return nil
}

func (m *mockAuthUsecase) AuthenticateRequest(ctx context.Context, tokenStr string) (*domain.Session, error) {
	if m.authenticateRequestFunc != nil {
		return m.authenticateRequestFunc(ctx, tokenStr)
	}
	return nil, nil
}

// ── Register ──────────────────────────────────────────────────────────────────

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	req := newJSONRequest(t, http.MethodPost, "/register", `{"broken":`)
	resp := httptest.NewRecorder()

	handlerhttp.MakeHandler(NewAuthHandler(&mockAuthUsecase{}).Register).ServeHTTP(resp, req)

	var body httperror.ErrorResponse
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, http.StatusBadRequest, resp.Code)
	require.Equal(t, apperror.CodeValidationFailed, body.Code)
}

func TestAuthHandler_Register_EmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/register", nil)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	handlerhttp.MakeHandler(NewAuthHandler(&mockAuthUsecase{}).Register).ServeHTTP(resp, req)

	var body httperror.ErrorResponse
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, http.StatusBadRequest, resp.Code)
	require.Equal(t, apperror.CodeValidationFailed, body.Code)
}

func TestAuthHandler_Register_ValidationErrors(t *testing.T) {
	tests := []struct {
		name       string
		payload    string
		wantFields map[string]string
	}{
		{
			name:       "username too short",
			payload:    `{"username":"ab","email":"user@example.com","password":"Password1"}`,
			wantFields: map[string]string{"username": "field_too_short"},
		},
		{
			name:       "username too long",
			payload:    fmt.Sprintf(`{"username":%q,"email":"user@example.com","password":"Password1"}`, strings.Repeat("a", 31)),
			wantFields: map[string]string{"username": "field_too_long"},
		},
		{
			name:       "username with space",
			payload:    `{"username":"user name","email":"user@example.com","password":"Password1"}`,
			wantFields: map[string]string{"username": "field_invalid"},
		},
		{
			name:       "username with at-sign",
			payload:    `{"username":"user@name","email":"user@example.com","password":"Password1"}`,
			wantFields: map[string]string{"username": "field_invalid"},
		},
		{
			name:       "username with emoji",
			payload:    `{"username":"user😀","email":"user@example.com","password":"Password1"}`,
			wantFields: map[string]string{"username": "field_invalid"},
		},
		{
			name:       "username with cyrillic",
			payload:    `{"username":"пользователь","email":"user@example.com","password":"Password1"}`,
			wantFields: map[string]string{"username": "field_invalid"},
		},
		{
			name:       "email empty",
			payload:    `{"username":"validuser","email":"","password":"Password1"}`,
			wantFields: map[string]string{"email": "field_required"},
		},
		{
			name:       "email invalid — no @",
			payload:    `{"username":"validuser","email":"notanemail","password":"Password1"}`,
			wantFields: map[string]string{"email": "field_invalid"},
		},
		{
			name:       "email invalid — missing TLD",
			payload:    `{"username":"validuser","email":"user@domain","password":"Password1"}`,
			wantFields: map[string]string{"email": "field_invalid"},
		},
		{
			name:       "password too short",
			payload:    `{"username":"validuser","email":"user@example.com","password":"Ab1"}`,
			wantFields: map[string]string{"password": "field_too_short"},
		},
		{
			name:       "password no digits",
			payload:    `{"username":"validuser","email":"user@example.com","password":"OnlyLetters"}`,
			wantFields: map[string]string{"password": "field_weak"},
		},
		{
			name:       "password no letters",
			payload:    `{"username":"validuser","email":"user@example.com","password":"12345678"}`,
			wantFields: map[string]string{"password": "field_weak"},
		},
		{
			name:       "password too long",
			payload:    fmt.Sprintf(`{"username":"validuser","email":"user@example.com","password":%q}`, strings.Repeat("A1", 65)),
			wantFields: map[string]string{"password": "field_too_long"},
		},
		{
			name:    "multiple fields invalid at once",
			payload: `{"username":"ab","email":"","password":"short"}`,
			wantFields: map[string]string{
				"username": "field_too_short",
				"email":    "field_required",
				"password": "field_too_short",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newJSONRequest(t, http.MethodPost, "/register", tt.payload)
			resp := httptest.NewRecorder()

			handlerhttp.MakeHandler(NewAuthHandler(&mockAuthUsecase{}).Register).ServeHTTP(resp, req)

			var body httperror.ErrorResponse
			decodeJSONResponse(t, resp, &body)
			require.Equal(t, http.StatusUnprocessableEntity, resp.Code)
			require.Equal(t, apperror.CodeValidationFailed, body.Code)
			require.Equal(t, tt.wantFields, body.Fields)
		})
	}
}

func TestAuthHandler_Register_Success(t *testing.T) {
	userID := uuid.New()
	called := false
	svc := &mockAuthUsecase{
		registerFunc: func(_ context.Context, req dto.RegisterUserDto) (*dto.RegisterResponseDto, error) {
			called = true
			require.Equal(t, "validuser", req.Username)
			require.Equal(t, "user@example.com", req.Email)
			require.Equal(t, "Password1", req.Password)
			return &dto.RegisterResponseDto{ID: userID, Username: "validuser", Email: "user@example.com"}, nil
		},
	}

	req := newJSONRequest(t, http.MethodPost, "/register",
		`{"username":"validuser","email":"user@example.com","password":"Password1"}`)
	resp := httptest.NewRecorder()

	handlerhttp.MakeHandler(NewAuthHandler(svc).Register).ServeHTTP(resp, req)

	require.True(t, called)
	require.Equal(t, http.StatusCreated, resp.Code)
	var body dto.RegisterResponseDto
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, userID, body.ID)
	require.Equal(t, "validuser", body.Username)
	require.Equal(t, "user@example.com", body.Email)
}

func TestAuthHandler_Register_NoPasswordInResponse(t *testing.T) {
	svc := &mockAuthUsecase{
		registerFunc: func(_ context.Context, _ dto.RegisterUserDto) (*dto.RegisterResponseDto, error) {
			return &dto.RegisterResponseDto{ID: uuid.New(), Username: "user", Email: "u@e.com"}, nil
		},
	}
	req := newJSONRequest(t, http.MethodPost, "/register",
		`{"username":"validuser","email":"user@example.com","password":"Password1"}`)
	resp := httptest.NewRecorder()

	handlerhttp.MakeHandler(NewAuthHandler(svc).Register).ServeHTTP(resp, req)

	var body map[string]any
	decodeJSONResponse(t, resp, &body)
	for _, key := range []string{"password", "password_hash", "passwordHash", "PasswordHash"} {
		_, found := body[key]
		require.Falsef(t, found, "field %q must not appear in response", key)
	}
}

func TestAuthHandler_Register_ServiceErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
		wantField  string
	}{
		{
			name:       "email already exists",
			err:        apperror.EmailAlreadyExists(),
			wantStatus: http.StatusConflict,
			wantCode:   apperror.CodeEmailExists,
			wantField:  "email",
		},
		{
			name:       "username already exists",
			err:        apperror.UsernameAlreadyExists(),
			wantStatus: http.StatusConflict,
			wantCode:   apperror.CodeUsernameExists,
			wantField:  "username",
		},
		{
			name:       "internal error",
			err:        apperror.Internal("something failed", nil),
			wantStatus: http.StatusInternalServerError,
			wantCode:   apperror.CodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockAuthUsecase{
				registerFunc: func(_ context.Context, _ dto.RegisterUserDto) (*dto.RegisterResponseDto, error) {
					return nil, tt.err
				},
			}
			req := newJSONRequest(t, http.MethodPost, "/register",
				`{"username":"validuser","email":"user@example.com","password":"Password1"}`)
			resp := httptest.NewRecorder()

			handlerhttp.MakeHandler(NewAuthHandler(svc).Register).ServeHTTP(resp, req)

			var body httperror.ErrorResponse
			decodeJSONResponse(t, resp, &body)
			require.Equal(t, tt.wantStatus, resp.Code)
			require.Equal(t, tt.wantCode, body.Code)
			if tt.wantField != "" {
				require.Equal(t, tt.wantField, body.Field)
			}
		})
	}
}

// ── Login ─────────────────────────────────────────────────────────────────────

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	req := newJSONRequest(t, http.MethodPost, "/login", `{bad json`)
	resp := httptest.NewRecorder()

	handlerhttp.MakeHandler(NewAuthHandler(&mockAuthUsecase{}).Login).ServeHTTP(resp, req)

	var body httperror.ErrorResponse
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, http.StatusBadRequest, resp.Code)
	require.Equal(t, apperror.CodeValidationFailed, body.Code)
}

func TestAuthHandler_Login_EmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	handlerhttp.MakeHandler(NewAuthHandler(&mockAuthUsecase{}).Login).ServeHTTP(resp, req)

	var body httperror.ErrorResponse
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, http.StatusBadRequest, resp.Code)
	require.Equal(t, apperror.CodeValidationFailed, body.Code)
}

func TestAuthHandler_Login_Success_WithDeviceID(t *testing.T) {
	deviceID := uuid.New()
	called := false
	svc := &mockAuthUsecase{
		loginFunc: func(_ context.Context, req dto.LoginUserDTO) (string, error) {
			called = true
			require.Equal(t, "user@example.com", req.Email)
			require.Equal(t, "Password1", req.Password)
			require.Equal(t, deviceID, req.DeviceID)
			return "jwt.token.value", nil
		},
	}

	payload := fmt.Sprintf(`{"email":"user@example.com","password":"Password1","deviceID":%q}`, deviceID)
	req := newJSONRequest(t, http.MethodPost, "/login", payload)
	resp := httptest.NewRecorder()

	handlerhttp.MakeHandler(NewAuthHandler(svc).Login).ServeHTTP(resp, req)

	require.True(t, called)
	require.Equal(t, http.StatusOK, resp.Code)
	var body map[string]string
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, "jwt.token.value", body["token"])
}

func TestAuthHandler_Login_Success_WithoutDeviceID(t *testing.T) {
	called := false
	svc := &mockAuthUsecase{
		loginFunc: func(_ context.Context, req dto.LoginUserDTO) (string, error) {
			called = true
			require.Equal(t, uuid.Nil, req.DeviceID)
			return "jwt.token.value", nil
		},
	}

	req := newJSONRequest(t, http.MethodPost, "/login",
		`{"email":"user@example.com","password":"Password1"}`)
	resp := httptest.NewRecorder()

	handlerhttp.MakeHandler(NewAuthHandler(svc).Login).ServeHTTP(resp, req)

	require.True(t, called)
	require.Equal(t, http.StatusOK, resp.Code)
}

func TestAuthHandler_Login_ServiceErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{
			name:       "invalid credentials",
			err:        apperror.InvalidCredentials(),
			wantStatus: http.StatusUnauthorized,
			wantCode:   apperror.CodeInvalidCredentials,
		},
		{
			name:       "account deleted",
			err:        apperror.AccountDeleted(),
			wantStatus: http.StatusForbidden,
			wantCode:   apperror.CodeAccountDeleted,
		},
		{
			name:       "internal error",
			err:        apperror.Internal("db down", nil),
			wantStatus: http.StatusInternalServerError,
			wantCode:   apperror.CodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockAuthUsecase{
				loginFunc: func(_ context.Context, _ dto.LoginUserDTO) (string, error) {
					return "", tt.err
				},
			}
			req := newJSONRequest(t, http.MethodPost, "/login",
				`{"email":"user@example.com","password":"Password1"}`)
			resp := httptest.NewRecorder()

			handlerhttp.MakeHandler(NewAuthHandler(svc).Login).ServeHTTP(resp, req)

			var body httperror.ErrorResponse
			decodeJSONResponse(t, resp, &body)
			require.Equal(t, tt.wantStatus, resp.Code)
			require.Equal(t, tt.wantCode, body.Code)
		})
	}
}

// ── Logout ────────────────────────────────────────────────────────────────────

func TestAuthHandler_Logout_NoSessionInContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	resp := httptest.NewRecorder()

	handlerhttp.MakeHandler(NewAuthHandler(&mockAuthUsecase{}).Logout).ServeHTTP(resp, req)

	var body httperror.ErrorResponse
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, http.StatusUnauthorized, resp.Code)
	require.Equal(t, apperror.CodeUnauthorized, body.Code)
}

func TestAuthHandler_Logout_WrongTypeInContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.CtxSessionIDKey, "not-a-uuid"))
	resp := httptest.NewRecorder()

	handlerhttp.MakeHandler(NewAuthHandler(&mockAuthUsecase{}).Logout).ServeHTTP(resp, req)

	var body httperror.ErrorResponse
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, http.StatusUnauthorized, resp.Code)
	require.Equal(t, apperror.CodeUnauthorized, body.Code)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	sessionID := uuid.New()
	called := false
	svc := &mockAuthUsecase{
		logoutFunc: func(_ context.Context, tokenID uuid.UUID) error {
			called = true
			require.Equal(t, sessionID, tokenID)
			return nil
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.CtxSessionIDKey, sessionID))
	resp := httptest.NewRecorder()

	handlerhttp.MakeHandler(NewAuthHandler(svc).Logout).ServeHTTP(resp, req)

	require.True(t, called)
	require.Equal(t, http.StatusOK, resp.Code)
	var body map[string]string
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, "logged out successfully", body["message"])
}

func TestAuthHandler_Logout_ServiceErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{
			name:       "session expired",
			err:        apperror.SessionExpired(),
			wantStatus: http.StatusUnauthorized,
			wantCode:   apperror.CodeSessionExpired,
		},
		{
			name:       "internal error",
			err:        apperror.Internal("failed to delete session", nil),
			wantStatus: http.StatusInternalServerError,
			wantCode:   apperror.CodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockAuthUsecase{
				logoutFunc: func(_ context.Context, _ uuid.UUID) error { return tt.err },
			}
			req := httptest.NewRequest(http.MethodPost, "/logout", nil)
			req = req.WithContext(context.WithValue(req.Context(), middleware.CtxSessionIDKey, uuid.New()))
			resp := httptest.NewRecorder()

			handlerhttp.MakeHandler(NewAuthHandler(svc).Logout).ServeHTTP(resp, req)

			var body httperror.ErrorResponse
			decodeJSONResponse(t, resp, &body)
			require.Equal(t, tt.wantStatus, resp.Code)
			require.Equal(t, tt.wantCode, body.Code)
		})
	}
}
