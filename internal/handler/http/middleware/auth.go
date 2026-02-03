package middleware

import (
	"context"
	"errors"
	"github.com/rockkley/pushpost/internal/apperror"
	"github.com/rockkley/pushpost/internal/handler/httperror"
	"github.com/rockkley/pushpost/internal/service"
	"net/http"
	"strings"
)

type contextKey string

const (
	CtxUserIDKey    contextKey = "userID"
	CtxSessionIDKey contextKey = "sessionID"
	CtxDeviceIDKey  contextKey = "deviceID"
)

type AuthMiddleware struct {
	authService service.AuthService
}

func NewAuthMiddleware(authService service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenStr, err := extractBearerToken(r.Header.Get("Authorization"))
		if err != nil {
			writeError(w, apperror.Unauthorized(apperror.CodeUnauthorized, "missing authorization header"))
			return
		}

		session, err := m.authService.AuthenticateRequest(r.Context(), tokenStr)
		if err != nil {
			writeError(w, err)
			return
		}
		ctx := context.WithValue(r.Context(), CtxUserIDKey, session.UserID)
		ctx = context.WithValue(ctx, CtxSessionIDKey, session.SessionID)
		ctx = context.WithValue(ctx, CtxDeviceIDKey, session.DeviceID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractBearerToken(header string) (string, error) {
	if header == "" {

		return "", errors.New("missing authorization header")
	}

	parts := strings.SplitN(header, " ", 2)

	if len(parts) != 2 || parts[0] != "Bearer" {

		return "", errors.New("invalid authorization format")
	}

	token := strings.TrimSpace(parts[1])

	if token == "" {

		return "", errors.New("empty token")
	}

	return token, nil
}

func writeError(w http.ResponseWriter, err error) {
	var appErr apperror.AppError
	if errors.As(err, &appErr) {
		response := httperror.ErrorResponse{
			Code:  appErr.Code(),
			Field: appErr.Field(),
		}
		if fieldErrors, ok := appErr.(interface{ Fields() map[string]string }); ok {
			response.Fields = fieldErrors.Fields()
		}
		_ = httperror.WriteJSON(w, appErr.HTTPStatus(), response)
		return
	}

	_ = httperror.WriteJSON(w, http.StatusInternalServerError, httperror.ErrorResponse{Code: apperror.CodeInternalError})
}
