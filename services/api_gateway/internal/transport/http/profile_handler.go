package http

import (
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/clients/friendship_api"
	"github.com/rockkley/pushpost/clients/profile_grpc"
	gwmiddleware "github.com/rockkley/pushpost/services/api_gateway/internal/middleware"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/ctxlog"
	"github.com/rockkley/pushpost/services/common_service/httperror"
)

var usernamePathRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

type ProfileHandler struct {
	profileClient    *profile_grpc.Client
	friendshipClient friendship_api.Client
}

type ProfileResponse struct {
	ID               uuid.UUID `json:"id"`
	Username         string    `json:"username"`
	CreatedAt        string    `json:"created_at"`
	DisplayName      string    `json:"display_name,omitempty"`
	FirstName        string    `json:"first_name,omitempty"`
	LastName         string    `json:"last_name,omitempty"`
	BirthDate        string    `json:"birth_date,omitempty"`
	AvatarURL        string    `json:"avatar_url,omitempty"`
	Bio              string    `json:"bio,omitempty"`
	TelegramLink     string    `json:"telegram_link,omitempty"`
	IsPrivate        bool      `json:"is_private"`
	FriendshipStatus string    `json:"friendship_status,omitempty"`
}

func NewProfileHandler(profileClient *profile_grpc.Client, friendshipClient friendship_api.Client) *ProfileHandler {
	return &ProfileHandler{
		profileClient:    profileClient,
		friendshipClient: friendshipClient,
	}
}

func (h *ProfileHandler) GetProfileByUsername(w http.ResponseWriter, r *http.Request) error {
	log := ctxlog.From(r.Context()).With(slog.String("op", "ProfileHandler.GetProfileByUsername"))

	username := strings.TrimSpace(chi.URLParam(r, "username"))

	if !usernamePathRegex.MatchString(username) {
		return commonapperr.NotFound("user_not_found", "user not found")
	}

	profile, err := h.profileClient.GetByUsername(r.Context(), username)

	if err != nil {
		if errors.Is(err, profile_grpc.ErrNotFound) {
			return commonapperr.NotFound("user_not_found", "user not found")
		}

		return commonapperr.Service("failed to fetch profile", err)
	}

	userID, parseErr := uuid.Parse(profile.UserID)

	if parseErr != nil {
		return commonapperr.Service("invalid profile user_id", parseErr)
	}

	resp := ProfileResponse{
		ID:           userID,
		Username:     profile.Username,
		CreatedAt:    profile.CreatedAt,
		DisplayName:  profile.DisplayName,
		FirstName:    profile.FirstName,
		LastName:     profile.LastName,
		BirthDate:    profile.BirthDate,
		AvatarURL:    profile.AvatarURL,
		Bio:          profile.Bio,
		TelegramLink: profile.TelegramLink,
		IsPrivate:    profile.IsPrivate,
	}

	viewerID, ok := gwmiddleware.UserIDFromContext(r.Context())

	if ok && viewerID != userID {
		rel, relErr := h.friendshipClient.GetRelationship(r.Context(), viewerID, userID)
		if relErr == nil {
			resp.FriendshipStatus = friendship_api.ResolveStatus(rel)
		} else {
			log.Warn("failed to fetch friendship status",
				slog.String("viewer_id", viewerID.String()),
				slog.String("target_id", userID.String()),
				slog.Any("error", relErr),
			)
		}
	}

	return httperror.WriteJSON(w, http.StatusOK, resp)
}
