package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/httperror"
	"github.com/rockkley/pushpost/services/profile_service/internal/domain"
	"github.com/rockkley/pushpost/services/profile_service/internal/entity"
)

const userIDHeader = "X-User-ID"

type ProfileHandler struct {
	uc domain.ProfileUseCaseInterface
}

func NewProfileHandler(uc domain.ProfileUseCaseInterface) *ProfileHandler {
	return &ProfileHandler{uc: uc}
}

func (h *ProfileHandler) GetByUsername(w http.ResponseWriter, r *http.Request) error {
	username := strings.TrimSpace(chi.URLParam(r, "username"))
	if username == "" {
		return commonapperr.BadRequest(commonapperr.CodeFieldRequired, "username is required")
	}

	profile, err := h.uc.GetByUsername(r.Context(), username)
	if err != nil {
		if err == domain.ErrProfileNotFound {
			return commonapperr.NotFound("profile_not_found", "profile not found")
		}

		return commonapperr.Service("failed to get profile", err)
	}

	return httperror.WriteJSON(w, http.StatusOK, profile)
}

func (h *ProfileHandler) UpdateMe(w http.ResponseWriter, r *http.Request) error {
	userIDRaw := strings.TrimSpace(r.Header.Get(userIDHeader))
	if userIDRaw == "" {
		return commonapperr.Unauthorized(commonapperr.CodeUnauthorized, "missing user id")
	}

	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid X-User-ID")
	}

	var body struct {
		DisplayName  string `json:"display_name"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		BirthDate    string `json:"birth_date"`
		AvatarURL    string `json:"avatar_url"`
		Bio          string `json:"bio"`
		TelegramLink string `json:"telegram_link"`
		IsPrivate    bool   `json:"is_private"`
	}

	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		return commonapperr.BadRequest(commonapperr.CodeValidationFailed, "invalid JSON")
	}

	birthDate, err := parseBirthDate(strings.TrimSpace(body.BirthDate))
	if err != nil {
		return err
	}

	profile := &entity.Profile{
		UserID:       userID,
		DisplayName:  normalizeOptional(body.DisplayName),
		FirstName:    normalizeOptional(body.FirstName),
		LastName:     normalizeOptional(body.LastName),
		BirthDate:    birthDate,
		AvatarURL:    normalizeOptional(body.AvatarURL),
		Bio:          normalizeOptional(body.Bio),
		TelegramLink: normalizeOptional(body.TelegramLink),
		IsPrivate:    body.IsPrivate,
	}

	if err = validateProfileFields(profile); err != nil {
		return err
	}

	if err = h.uc.UpdateProfile(r.Context(), profile); err != nil {
		if err == domain.ErrProfileNotFound {
			return commonapperr.NotFound("profile_not_found", "profile not found")
		}

		return commonapperr.Service("failed to update profile", err)
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]string{"message": "profile updated"})
}

func normalizeOptional(v string) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}

	return &v
}

func parseBirthDate(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, commonapperr.Validation(
			commonapperr.CodeFieldInvalid,
			"birth_date",
			"birth_date must be in YYYY-MM-DD format",
		)
	}

	if parsed.After(time.Now()) {
		return nil, commonapperr.Validation(
			commonapperr.CodeFieldInvalid,
			"birth_date",
			"birth_date cannot be in the future",
		)
	}

	return &parsed, nil
}

func validateProfileFields(profile *entity.Profile) error {
	if profile.DisplayName != nil && len(*profile.DisplayName) > 60 {
		return commonapperr.Validation(commonapperr.CodeFieldTooLong, "display_name", "display_name is too long")
	}

	if profile.FirstName != nil && len(*profile.FirstName) > 60 {
		return commonapperr.Validation(commonapperr.CodeFieldTooLong, "first_name", "first_name is too long")
	}

	if profile.LastName != nil && len(*profile.LastName) > 60 {
		return commonapperr.Validation(commonapperr.CodeFieldTooLong, "last_name", "last_name is too long")
	}

	if profile.Bio != nil && len(*profile.Bio) > 500 {
		return commonapperr.Validation(commonapperr.CodeFieldTooLong, "bio", "bio is too long")
	}

	if profile.TelegramLink != nil && len(*profile.TelegramLink) > 255 {
		return commonapperr.Validation(commonapperr.CodeFieldTooLong, "telegram_link", "telegram_link is too long")
	}

	return nil
}
