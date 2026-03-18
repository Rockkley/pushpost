package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/httperror"
	"github.com/rockkley/pushpost/services/common_service/jwt"
)

type contextKey string

const CtxUserIDKey contextKey = "userID"

type JwtAuthMiddleware struct {
	jwtManager *jwt.Manager
}

func NewJwtAuthMiddleware(jwtManager *jwt.Manager) *JwtAuthMiddleware {
	return &JwtAuthMiddleware{jwtManager: jwtManager}
}

func (m *JwtAuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		ctx := context.WithValue(r.Context(), CtxUserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractBearer(header string) string {
	const prefix = "Bearer "
	if len(header) <= len(prefix) {
		return ""
	}
	if header[:len(prefix)] != prefix {
		return ""
	}
	return header[len(prefix):]
}
