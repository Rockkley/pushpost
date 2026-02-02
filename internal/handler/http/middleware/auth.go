package middleware

import (
	"context"
	"errors"
	http2 "github.com/rockkley/pushpost/internal/handler/http"
	"github.com/rockkley/pushpost/internal/service"
	"net/http"
	"strings"
)

type contextKey string

const (
	ctxUserIDKey    contextKey = "userID"
	ctxSessionIDKey contextKey = "sessionID"
	ctxDeviceIDKey  contextKey = "deviceID"
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
			http2.WriteJSON(w, http.StatusUnauthorized, http2.ErrorResponse{Code: "unauthorized"})
			return
		}

		session, err := m.authService.AuthenticateRequest(r.Context(), tokenStr)
		if err != nil {
			http2.WriteJSON(w, http.StatusUnauthorized, http2.ErrorResponse{Code: "unauthorized"})
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserIDKey, session.UserID)
		ctx = context.WithValue(ctx, ctxSessionIDKey, session.SessionID)
		ctx = context.WithValue(ctx, ctxDeviceIDKey, session.DeviceID)

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
