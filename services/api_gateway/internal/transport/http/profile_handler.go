package http

import (
	"errors"
	"github.com/rockkley/pushpost/services/common_service/ctxlog"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/rockkley/pushpost/clients/profile_api"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/clients/friendship_api"
	gwmiddleware "github.com/rockkley/pushpost/services/api_gateway/internal/middleware"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/httperror"
)

var usernamePathRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

type ProfileHandler struct {
	profileClient    profile_api.Client
	friendshipClient friendship_api.Client
}

type ProfileHandlerDeps struct {
	ProfileClient    profile_api.Client
	FriendshipClient friendship_api.Client
}

type ProfileResponse struct {
	ID               uuid.UUID `json:"id"`
	Username         string    `json:"username"`
	CreatedAt        string    `json:"created_at"`
	FriendshipStatus string    `json:"friendship_status,omitempty"`
}

func NewProfileHandler(deps ProfileHandlerDeps) *ProfileHandler {
	return &ProfileHandler{
		profileClient:    deps.ProfileClient,
		friendshipClient: deps.FriendshipClient,
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
		if errors.Is(err, profile_api.ErrNotFound) {
			return commonapperr.NotFound("user_not_found", "user not found")
		}
		return commonapperr.Service("failed to fetch profile", err)
	}

	resp := ProfileResponse{
		ID:        profile.UserID,
		Username:  profile.Username,
		CreatedAt: profile.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	viewerID, ok := gwmiddleware.UserIDFromContext(r.Context())
	if ok && viewerID != profile.UserID {
		rel, relErr := h.friendshipClient.GetRelationship(r.Context(), viewerID, profile.UserID)
		if relErr == nil {
			resp.FriendshipStatus = friendship_api.ResolveStatus(rel)
		} else {
			log.Warn("failed to fetch friendship status",
				slog.String("viewer_id", viewerID.String()),
				slog.String("target_id", profile.UserID.String()),
				slog.Any("error", relErr),
			)
		}
	}

	return httperror.WriteJSON(w, http.StatusOK, resp)
}
