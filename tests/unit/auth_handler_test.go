package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	http2 "github.com/rockkley/pushpost/internal/handler/http"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rockkley/pushpost/internal/apperror"
	"github.com/rockkley/pushpost/internal/domain"
	"github.com/rockkley/pushpost/internal/handler/http/dto"
	"github.com/rockkley/pushpost/internal/handler/http/middleware"
	"github.com/rockkley/pushpost/internal/handler/httperror"
	"github.com/stretchr/testify/require"
)

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

type mockAuthService struct {
	registerFunc            func(ctx context.Context, req dto.RegisterUserDto) (*domain.User, error)
	loginFunc               func(ctx context.Context, req dto.LoginUserDTO) (string, error)
	logoutFunc              func(ctx context.Context, tokenID uuid.UUID) error
	authenticateRequestFunc func(ctx context.Context, tokenStr string) (*domain.Session, error)
}

func (m *mockAuthService) Register(ctx context.Context, req dto.RegisterUserDto) (*domain.User, error) {
	if m.registerFunc == nil {
		return nil, nil
	}
	return m.registerFunc(ctx, req)
}

func (m *mockAuthService) Login(ctx context.Context, req dto.LoginUserDTO) (string, error) {
	if m.loginFunc == nil {
		return "", nil
	}
	return m.loginFunc(ctx, req)
}

func (m *mockAuthService) Logout(ctx context.Context, tokenID uuid.UUID) error {
	if m.logoutFunc == nil {
		return nil
	}
	return m.logoutFunc(ctx, tokenID)
}

func (m *mockAuthService) AuthenticateRequest(ctx context.Context, tokenStr string) (*domain.Session, error) {
	if m.authenticateRequestFunc == nil {
		return nil, nil
	}
	return m.authenticateRequestFunc(ctx, tokenStr)
}

func TestAuthHandlerRegisterInvalidJSON(t *testing.T) {
	handler := http2.NewAuthHandler(&mockAuthService{})
	req := newJSONRequest(t, http.MethodPost, "/register", "{")
	resp := httptest.NewRecorder()

	http2.MakeHandler(handler.Register).ServeHTTP(resp, req)

	var body httperror.ErrorResponse
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, http.StatusBadRequest, resp.Code)
	require.Equal(t, apperror.CodeValidationFailed, body.Code)
}

func TestAuthHandlerRegisterValidationErrors(t *testing.T) {
	tests := []struct {
		name       string
		payload    string
		wantFields map[string]string
	}{
		{
			name:       "multiple fields invalid",
			payload:    `{"username":"ab","email":"","password":"short"}`,
			wantFields: map[string]string{"username": "field_too_short", "email": "field_required", "password": "field_too_short"},
		},
		{
			name:       "invalid email",
			payload:    `{"username":"valid_user","email":"bad","password":"Password123"}`,
			wantFields: map[string]string{"email": "field_invalid"},
		},
		{
			name:       "password without digits",
			payload:    `{"username":"valid_user","email":"user@example.com","password":"OnlyLetters"}`,
			wantFields: map[string]string{"password": "field_weak"},
		},
		{
			name:       "password without letters",
			payload:    `{"username":"valid_user","email":"user@example.com","password":"12345678"}`,
			wantFields: map[string]string{"password": "field_weak"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http2.NewAuthHandler(&mockAuthService{})
			req := newJSONRequest(t, http.MethodPost, "/register", tt.payload)
			resp := httptest.NewRecorder()

			http2.MakeHandler(handler.Register).ServeHTTP(resp, req)

			var body httperror.ErrorResponse
			decodeJSONResponse(t, resp, &body)
			require.Equal(t, http.StatusUnprocessableEntity, resp.Code)
			require.Equal(t, apperror.CodeValidationFailed, body.Code)
			require.Equal(t, tt.wantFields, body.Fields)
		})
	}
}

func TestAuthHandlerRegisterSuccess(t *testing.T) {
	called := false
	expectedReq := dto.RegisterUserDto{Username: "username", Email: "user@example.com", Password: "Password123"}
	userID := uuid.New()
	service := &mockAuthService{
		registerFunc: func(ctx context.Context, req dto.RegisterUserDto) (*domain.User, error) {
			called = true
			require.Equal(t, expectedReq, req)
			return &domain.User{Id: userID, Username: "username", Email: "user@example.com"}, nil
		},
	}
	payload := `{"username":"username","email":"user@example.com","password":"Password123"}`
	handler := http2.NewAuthHandler(service)
	req := newJSONRequest(t, http.MethodPost, "/register", payload)
	resp := httptest.NewRecorder()

	http2.MakeHandler(handler.Register).ServeHTTP(resp, req)

	require.True(t, called)
	require.Equal(t, http.StatusCreated, resp.Code)
	var body domain.User
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, userID, body.Id)
	require.Equal(t, "username", body.Username)
	require.Equal(t, "user@example.com", body.Email)
}

func TestAuthHandlerRegisterNoPasswordHashInResponse(t *testing.T) {
	service := &mockAuthService{
		registerFunc: func(ctx context.Context, req dto.RegisterUserDto) (*domain.User, error) {
			return &domain.User{
				Id:           uuid.New(),
				Username:     "user",
				Email:        "user@example.com",
				PasswordHash: "hashed_password_should_not_be_returned",
			}, nil
		},
	}
	handler := http2.NewAuthHandler(service)
	req := newJSONRequest(t, http.MethodPost, "/register", `{"username":"user","email":"user@example.com","password":"Password123"}`)
	resp := httptest.NewRecorder()

	http2.MakeHandler(handler.Register).ServeHTTP(resp, req)

	var body map[string]any
	decodeJSONResponse(t, resp, &body)
	_, hasPasswordHash := body["password_hash"]
	require.False(t, hasPasswordHash, "password_hash should not be in response")
}

func TestAuthHandlerRegisterEmptyBody(t *testing.T) {
	handler := http2.NewAuthHandler(&mockAuthService{})
	req := httptest.NewRequest(http.MethodPost, "/register", nil)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	http2.MakeHandler(handler.Register).ServeHTTP(resp, req)

	var body httperror.ErrorResponse
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, http.StatusBadRequest, resp.Code)
	require.Equal(t, apperror.CodeValidationFailed, body.Code)
}

func TestAuthHandlerRegisterLongUsername(t *testing.T) {
	handler := http2.NewAuthHandler(&mockAuthService{})
	longUsername := strings.Repeat("a", 1000)
	payload := fmt.Sprintf(`{"username":"%s","email":"user@example.com","password":"Password123"}`, longUsername)
	req := newJSONRequest(t, http.MethodPost, "/register", payload)
	resp := httptest.NewRecorder()

	http2.MakeHandler(handler.Register).ServeHTTP(resp, req)

	var body httperror.ErrorResponse
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, http.StatusUnprocessableEntity, resp.Code)
	require.Equal(t, "field_too_long", body.Fields["username"])
}

func TestAuthHandlerRegisterInvalidCharacters(t *testing.T) {
	tests := []struct {
		name     string
		username string
	}{
		{name: "spaces", username: "user name"},
		{name: "special chars", username: "user@name"},
		{name: "emoji", username: "user😀"},
		{name: "cyrillic", username: "пользователь"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http2.NewAuthHandler(&mockAuthService{})
			payload := fmt.Sprintf(`{"username":"%s","email":"user@example.com","password":"Password123"}`, tt.username)
			req := newJSONRequest(t, http.MethodPost, "/register", payload)
			resp := httptest.NewRecorder()

			http2.MakeHandler(handler.Register).ServeHTTP(resp, req)

			var body httperror.ErrorResponse
			decodeJSONResponse(t, resp, &body)
			require.Equal(t, http.StatusUnprocessableEntity, resp.Code)
			require.Equal(t, "field_invalid", body.Fields["username"])
		})
	}
}

func TestAuthHandlerRegisterServiceError(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &mockAuthService{
				registerFunc: func(ctx context.Context, req dto.RegisterUserDto) (*domain.User, error) {
					return nil, tt.err
				},
			}
			handler := http2.NewAuthHandler(service)
			req := newJSONRequest(t, http.MethodPost, "/register", `{"username":"username","email":"user@example.com","password":"Password123"}`)
			resp := httptest.NewRecorder()

			http2.MakeHandler(handler.Register).ServeHTTP(resp, req)

			var body httperror.ErrorResponse
			decodeJSONResponse(t, resp, &body)
			require.Equal(t, tt.wantStatus, resp.Code)
			require.Equal(t, tt.wantCode, body.Code)
			require.Equal(t, tt.wantField, body.Field)
		})
	}
}

func TestAuthHandlerLoginInvalidJSON(t *testing.T) {
	handler := http2.NewAuthHandler(&mockAuthService{})
	req := newJSONRequest(t, http.MethodPost, "/login", "{")
	resp := httptest.NewRecorder()

	http2.MakeHandler(handler.Login).ServeHTTP(resp, req)

	var body httperror.ErrorResponse
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, http.StatusBadRequest, resp.Code)
	require.Equal(t, apperror.CodeValidationFailed, body.Code)
}

func TestAuthHandlerLoginSuccess(t *testing.T) {
	called := false
	deviceID := uuid.New()
	service := &mockAuthService{
		loginFunc: func(ctx context.Context, req dto.LoginUserDTO) (string, error) {
			called = true
			require.Equal(t, dto.LoginUserDTO{Email: "user@example.com", Password: "Password123", DeviceID: deviceID}, req)
			return "token-value", nil
		},
	}
	handler := http2.NewAuthHandler(service)
	payload := `{"email":"user@example.com","password":"Password123","deviceID":"` + deviceID.String() + `"}`
	req := newJSONRequest(t, http.MethodPost, "/login", payload)
	resp := httptest.NewRecorder()

	http2.MakeHandler(handler.Login).ServeHTTP(resp, req)

	require.True(t, called)
	require.Equal(t, http.StatusOK, resp.Code)
	var body map[string]string
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, map[string]string{"token": "token-value"}, body)
}

func TestAuthHandlerLoginNoDeviceID(t *testing.T) {
	called := false
	service := &mockAuthService{
		loginFunc: func(ctx context.Context, req dto.LoginUserDTO) (string, error) {
			called = true
			require.Equal(t, uuid.Nil, req.DeviceID)
			return "token-value", nil
		},
	}
	handler := http2.NewAuthHandler(service)
	req := newJSONRequest(t, http.MethodPost, "/login", `{"email":"user@example.com","password":"Password123"}`)
	resp := httptest.NewRecorder()

	http2.MakeHandler(handler.Login).ServeHTTP(resp, req)

	require.True(t, called)
	require.Equal(t, http.StatusOK, resp.Code)
}

func TestAuthHandlerLoginServiceError(t *testing.T) {
	service := &mockAuthService{
		loginFunc: func(ctx context.Context, req dto.LoginUserDTO) (string, error) {
			return "", apperror.InvalidCredentials()
		},
	}
	handler := http2.NewAuthHandler(service)
	req := newJSONRequest(t, http.MethodPost, "/login", `{"email":"user@example.com","password":"Password123","deviceID":"`+uuid.New().String()+`"}`)
	resp := httptest.NewRecorder()

	http2.MakeHandler(handler.Login).ServeHTTP(resp, req)

	var body httperror.ErrorResponse
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, http.StatusUnauthorized, resp.Code)
	require.Equal(t, apperror.CodeInvalidCredentials, body.Code)
}

func TestAuthHandlerLogoutUnauthorized(t *testing.T) {
	handler := http2.NewAuthHandler(&mockAuthService{})
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	resp := httptest.NewRecorder()

	http2.MakeHandler(handler.Logout).ServeHTTP(resp, req)

	var body httperror.ErrorResponse
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, http.StatusUnauthorized, resp.Code)
	require.Equal(t, apperror.CodeUnauthorized, body.Code)
}

func TestAuthHandlerLogoutSuccess(t *testing.T) {
	called := false
	sessionID := uuid.New()
	service := &mockAuthService{
		logoutFunc: func(ctx context.Context, tokenID uuid.UUID) error {
			called = true
			require.Equal(t, sessionID, tokenID)
			return nil
		},
	}
	handler := http2.NewAuthHandler(service)
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.CtxSessionIDKey, sessionID))
	resp := httptest.NewRecorder()

	http2.MakeHandler(handler.Logout).ServeHTTP(resp, req)

	require.True(t, called)
	require.Equal(t, http.StatusOK, resp.Code)
	var body map[string]string
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, map[string]string{"message": "logged out successfully"}, body)
}

func TestAuthHandlerLogoutServiceError(t *testing.T) {
	service := &mockAuthService{
		logoutFunc: func(ctx context.Context, tokenID uuid.UUID) error {
			return apperror.SessionExpired()
		},
	}
	handler := http2.NewAuthHandler(service)
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.CtxSessionIDKey, uuid.New()))
	resp := httptest.NewRecorder()

	http2.MakeHandler(handler.Logout).ServeHTTP(resp, req)

	var body httperror.ErrorResponse
	decodeJSONResponse(t, resp, &body)
	require.Equal(t, http.StatusUnauthorized, resp.Code)
	require.Equal(t, apperror.CodeSessionExpired, body.Code)
}
