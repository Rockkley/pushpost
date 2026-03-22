package middleware

import (
	"context"
	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/ctxlog"
	"github.com/rockkley/pushpost/services/common_service/httperror"
	"github.com/rockkley/pushpost/services/common_service/jwt"
	"log/slog"
	"net/http"
)

const HeaderUserID = "X-User-ID"

type contextKey string

const ctxUserIDKey contextKey = "userID"

type AuthMiddleware struct {
	jwtManager *jwt.Manager
}

func NewAuthMiddleware(jwtManager *jwt.Manager) *AuthMiddleware {
	return &AuthMiddleware{jwtManager: jwtManager}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del(HeaderUserID)
		tokenStr := extractBearer(r.Header.Get("Authorization"))

		if tokenStr == "" {
			httperror.HandleError(w, r, commonapperr.Unauthorized(
				commonapperr.CodeUnauthorized, "missing or invalid authorization header",
			))

			return
		}

		claims, err := m.jwtManager.Parse(tokenStr)

		if err != nil {
			httperror.HandleError(w, r, commonapperr.Unauthorized(
				commonapperr.CodeUnauthorized, "invalid token",
			))

			return
		}

		subStr, ok := claims["sub"].(string)

		if !ok || subStr == "" {
			httperror.HandleError(w, r, commonapperr.Unauthorized(
				commonapperr.CodeUnauthorized, "token missing subject",
			))

			return
		}

		userID, err := uuid.Parse(subStr)

		if err != nil {
			httperror.HandleError(w, r, commonapperr.Unauthorized(
				commonapperr.CodeUnauthorized, "invalid user id in token",
			))

			return
		}

		r.Header.Set(HeaderUserID, userID.String())

		ctx := context.WithValue(r.Context(), ctxUserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del(HeaderUserID)
		tokenStr := extractBearer(r.Header.Get("Authorization"))
		if tokenStr == "" {
			next.ServeHTTP(w, r)
			return
		}

		claims, err := m.jwtManager.Parse(tokenStr)
		if err != nil {
			ctxlog.From(r.Context()).Debug("optional auth: invalid token", slog.Any("error", err))
			next.ServeHTTP(w, r)
			return
		}

		subStr, ok := claims["sub"].(string)
		if !ok || subStr == "" {
			next.ServeHTTP(w, r)
			return
		}

		userID, err := uuid.Parse(subStr)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		r.Header.Set(HeaderUserID, userID.String())
		ctx := context.WithValue(r.Context(), ctxUserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(ctxUserIDKey).(uuid.UUID)
	return userID, ok && userID != uuid.Nil
}

func extractBearer(header string) string {
	const prefix = "Bearer "
	if len(header) <= len(prefix) || header[:len(prefix)] != prefix {

		return ""
	}

	return header[len(prefix):]
}
