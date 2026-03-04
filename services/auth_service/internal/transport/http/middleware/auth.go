package middleware

import (
	"context"
	"github.com/rockkley/pushpost/services/common/httperror"
	"net/http"
	"strings"

	"github.com/rockkley/pushpost/services/auth_service/internal/domain"
	"github.com/rockkley/pushpost/services/common/apperror"
)

type contextKey string

const (
	CtxUserIDKey    contextKey = "userID"
	CtxSessionIDKey contextKey = "sessionID"
	CtxDeviceIDKey  contextKey = "deviceID"
)

type AuthMiddleware struct {
	authUsecase domain.AuthUsecase
}

func NewAuthMiddleware(authUsecase domain.AuthUsecase) *AuthMiddleware {
	return &AuthMiddleware{authUsecase: authUsecase}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr, ok := extractBearerToken(r.Header.Get("Authorization"))
		if !ok {
			httperror.HandleError(w, r, apperror.Unauthorized(apperror.CodeUnauthorized, "missing or invalid authorization header"))
			return
		}

		session, err := m.authUsecase.AuthenticateRequest(r.Context(), tokenStr)
		if err != nil {
			httperror.HandleError(w, r, err)
			return
		}

		ctx := context.WithValue(r.Context(), CtxUserIDKey, session.UserID)
		ctx = context.WithValue(ctx, CtxSessionIDKey, session.SessionID)
		ctx = context.WithValue(ctx, CtxDeviceIDKey, session.DeviceID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractBearerToken(header string) (string, bool) {
	if header == "" {
		return "", false
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", false
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}

	return token, true
}
