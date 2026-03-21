package http

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/httperror"
	"github.com/rockkley/pushpost/services/profile_service/internal/domain"
)

type ProfileHandler struct {
	uc domain.ProfileUseCase
}

type ProfileResponse struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

func NewProfileHandler(uc domain.ProfileUseCase) *ProfileHandler {
	return &ProfileHandler{uc: uc}
}

func (h *ProfileHandler) GetByUsername(w http.ResponseWriter, r *http.Request) error {
	username := strings.TrimSpace(chi.URLParam(r, "username"))
	if username == "" {
		return commonapperr.Validation(commonapperr.CodeFieldRequired, "username", "username is required")
	}

	profile, err := h.uc.GetByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, domain.ErrProfileNotFound) {
			return commonapperr.NotFound("profile_not_found", "profile not found")
		}
		return commonapperr.Service("failed to fetch profile", err)
	}

	return httperror.WriteJSON(w, http.StatusOK, ProfileResponse{
		UserID:    profile.UserID,
		Username:  profile.Username,
		CreatedAt: profile.CreatedAt,
	})
}
