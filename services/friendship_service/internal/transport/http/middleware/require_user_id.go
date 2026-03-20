package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/httperror"
)

const HeaderUserID = "X-User-ID"

type contextKey string

const CtxUserIDKey contextKey = "userID"

func RequireUserID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawID := r.Header.Get(HeaderUserID)
		if rawID == "" {
			httperror.HandleError(w, r, commonapperr.Unauthorized(
				commonapperr.CodeUnauthorized, "missing X-User-ID header",
			))
			return
		}

		userID, err := uuid.Parse(rawID)
		if err != nil || userID == uuid.Nil {
			httperror.HandleError(w, r, commonapperr.Unauthorized(
				commonapperr.CodeUnauthorized, "invalid X-User-ID header",
			))
			return
		}

		ctx := context.WithValue(r.Context(), CtxUserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
