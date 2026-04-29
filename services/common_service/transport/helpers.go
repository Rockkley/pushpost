package transport

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/middleware"
)

func RequireUserID(r *http.Request) (uuid.UUID, error) {
	userID, ok := middleware.UserIDFromContext(r.Context())

	if !ok || userID == uuid.Nil {
		return uuid.Nil, apperror.Unauthorized(apperror.CodeUnauthorized, "missing authenticated user")
	}

	return userID, nil
}

func ParsePathUUID(r *http.Request, param string) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, param))

	if err != nil {
		return uuid.Nil, apperror.BadRequest(
			apperror.CodeFieldInvalid, "invalid "+param+" — must be a UUID",
		)
	}

	return id, nil
}

func ParsePagination(r *http.Request) (limit, offset int, err error) {
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if limit, err = strconv.Atoi(raw); err != nil {
			return 0, 0, apperror.BadRequest(apperror.CodeFieldInvalid, "invalid limit")
		}
	}

	if raw := r.URL.Query().Get("offset"); raw != "" {
		if offset, err = strconv.Atoi(raw); err != nil {
			return 0, 0, apperror.BadRequest(apperror.CodeFieldInvalid, "invalid offset")
		}
	}

	return limit, offset, nil
}
