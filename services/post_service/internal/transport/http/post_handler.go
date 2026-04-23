package http

import (
	"encoding/json"
	"github.com/rockkley/pushpost/services/post_service/internal/entity"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/httperror"
	commonmiddleware "github.com/rockkley/pushpost/services/common_service/middleware"
	"github.com/rockkley/pushpost/services/post_service/internal/domain"
)

type PostHandler struct {
	uc domain.PostUseCaseInterface
}

func NewPostHandler(uc domain.PostUseCaseInterface) *PostHandler {
	return &PostHandler{uc: uc}
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) error {
	authorID, ok := commonmiddleware.UserIDFromContext(r.Context())

	if !ok {
		return commonapperr.Unauthorized(commonapperr.CodeUnauthorized, "missing user id")
	}

	var body struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return commonapperr.BadRequest(commonapperr.CodeValidationFailed, "invalid JSON")
	}

	post, err := h.uc.CreatePost(r.Context(), authorID, body.Content)
	if err != nil {
		return err
	}
	return httperror.WriteJSON(w, http.StatusCreated, post)
}

func (h *PostHandler) GetFeed(w http.ResponseWriter, r *http.Request) error {
	userID, ok := commonmiddleware.UserIDFromContext(r.Context())
	if !ok {
		return commonapperr.Unauthorized(commonapperr.CodeUnauthorized, "missing user id")
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	cursor := r.URL.Query().Get("cursor")

	posts, nextCursor, err := h.uc.GetFeed(r.Context(), userID, limit, cursor)
	if err != nil {
		return err
	}
	if posts == nil {
		posts = []*entity.Post{}
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]any{
		"posts":       posts,
		"next_cursor": nextCursor,
	})
}

func (h *PostHandler) GetUserPosts(w http.ResponseWriter, r *http.Request) error {
	rawID := chi.URLParam(r, "userID")
	authorID, err := uuid.Parse(rawID)
	if err != nil {
		return commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid user id")
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	cursor := r.URL.Query().Get("cursor")

	posts, nextCursor, err := h.uc.GetUserPosts(r.Context(), authorID, limit, cursor)
	if err != nil {
		return err
	}
	if posts == nil {
		posts = []*entity.Post{}
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]any{
		"posts":       posts,
		"next_cursor": nextCursor,
	})
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) error {
	authorID, ok := commonmiddleware.UserIDFromContext(r.Context())
	if !ok {
		return commonapperr.Unauthorized(commonapperr.CodeUnauthorized, "missing user id")
	}

	postID, err := uuid.Parse(chi.URLParam(r, "postID"))
	if err != nil {
		return commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid post id")
	}

	if err = h.uc.DeletePost(r.Context(), postID, authorID); err != nil {
		return err
	}
	return httperror.WriteJSON(w, http.StatusOK, map[string]string{"message": "post deleted"})
}

func (h *PostHandler) GetPostByID(w http.ResponseWriter, r *http.Request) error {
	postID, err := uuid.Parse(chi.URLParam(r, "postID"))
	if err != nil {
		return commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid post id")
	}

	post, err := h.uc.GetPostByID(r.Context(), postID)
	if err != nil {
		return err
	}
	return httperror.WriteJSON(w, http.StatusOK, post)
}
