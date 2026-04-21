package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/httperror"
)

const HeaderUserID = "X-User-ID"

type userIDKey struct{}

func RequireUserID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawID := r.Header.Get(HeaderUserID)
		if rawID == "" {
			httperror.HandleError(w, r, apperror.Unauthorized(
				apperror.CodeUnauthorized, "missing X-User-ID header",
			))
			return
		}

		userID, err := uuid.Parse(rawID)
		if err != nil || userID == uuid.Nil {
			httperror.HandleError(w, r, apperror.Unauthorized(
				apperror.CodeUnauthorized, "invalid X-User-ID header",
			))
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey{}, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey{}).(uuid.UUID)
	return userID, ok && userID != uuid.Nil
}
