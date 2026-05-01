package http

import (
	"encoding/json"
	"github.com/rockkley/pushpost/services/post_service/internal/apperror"
	"github.com/rockkley/pushpost/services/post_service/internal/domain"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/httperror"
	commonmiddleware "github.com/rockkley/pushpost/services/common_service/middleware"
	"github.com/rockkley/pushpost/services/post_service/internal/entity"
)

const maxPostBodySize = 64 * 1024 // 64KB

type feedResponse struct {
	Posts      []*entity.Post `json:"posts"`
	NextCursor string         `json:"next_cursor"`
	TopCursor  string         `json:"top_cursor"`
}

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

	r.Body = http.MaxBytesReader(w, r.Body, maxPostBodySize)
	var body struct {
		Content string `json:"content"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		return commonapperr.BadRequest(commonapperr.CodeValidationFailed, "invalid request body")
	}

	post, err := h.uc.CreatePost(r.Context(), authorID, body.Content)
	if err != nil {
		return err
	}
	return httperror.WriteJSON(w, http.StatusCreated, post)
}

func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) error {
	authorID, ok := commonmiddleware.UserIDFromContext(r.Context())
	if !ok {
		return commonapperr.Unauthorized(commonapperr.CodeUnauthorized, "missing user id")
	}

	postID, err := uuid.Parse(chi.URLParam(r, "postID"))
	if err != nil {
		return commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid post id")
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxPostBodySize)
	var body struct {
		Content string `json:"content"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err = dec.Decode(&body); err != nil {
		return commonapperr.BadRequest(commonapperr.CodeValidationFailed, "invalid request body")
	}

	post, err := h.uc.UpdatePost(r.Context(), postID, authorID, body.Content)
	if err != nil {
		return err
	}
	return httperror.WriteJSON(w, http.StatusOK, post)
}

func (h *PostHandler) GetFeed(w http.ResponseWriter, r *http.Request) error {
	userID, ok := commonmiddleware.UserIDFromContext(r.Context())
	if !ok {
		return commonapperr.Unauthorized(commonapperr.CodeUnauthorized, "missing user id")
	}

	limit, err := parseOptionalIntQuery(r, "limit")

	if err != nil {
		return err
	}

	cursorToken := r.URL.Query().Get("cursor")
	sinceToken := r.URL.Query().Get("since")

	var (
		resp domain.FeedResponse
	)

	if sinceToken != "" {
		resp, err = h.uc.GetFeedSince(r.Context(), userID, limit, sinceToken)
	} else {
		resp, err = h.uc.GetFeed(r.Context(), userID, limit, cursorToken)
	}

	if err != nil {
		return err
	}

	posts := resp.Posts

	if posts == nil {
		posts = []*entity.Post{}
	}

	return httperror.WriteJSON(w, http.StatusOK, feedResponse{
		Posts:      posts,
		NextCursor: resp.NextCursor,
		TopCursor:  resp.TopCursor,
	})
}

func (h *PostHandler) GetUserPosts(w http.ResponseWriter, r *http.Request) error {
	authorID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		return commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid user id")
	}

	limit, err := parseOptionalIntQuery(r, "limit")

	if err != nil {
		return err
	}

	cursorToken := r.URL.Query().Get("cursor")

	resp, err := h.uc.GetUserPosts(r.Context(), authorID, limit, cursorToken)
	if err != nil {
		return err
	}

	posts := resp.Posts
	if posts == nil {
		posts = []*entity.Post{}
	}

	return httperror.WriteJSON(w, http.StatusOK, feedResponse{
		Posts:      posts,
		NextCursor: resp.NextCursor,
		TopCursor:  resp.TopCursor,
	})
}

func (h *PostHandler) GetPostsByIDs(w http.ResponseWriter, r *http.Request) error {
	rawIDs := r.URL.Query().Get("ids")
	if rawIDs == "" {
		return commonapperr.BadRequest(commonapperr.CodeFieldRequired, "ids query param is required")
	}

	parts := strings.Split(rawIDs, ",")
	ids := make([]uuid.UUID, 0, len(parts))
	for _, p := range parts {
		id, err := uuid.Parse(strings.TrimSpace(p))
		if err != nil {
			return commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid id: "+p)
		}
		ids = append(ids, id)
	}

	posts, err := h.uc.GetPostsByIDs(r.Context(), ids)
	if err != nil {
		return err
	}
	if posts == nil {
		posts = []*entity.Post{}
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]any{"posts": posts})
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

func (h *PostHandler) LikePost(w http.ResponseWriter, r *http.Request) error {
	userID, ok := commonmiddleware.UserIDFromContext(r.Context())

	if !ok {
		return commonapperr.Unauthorized(commonapperr.CodeUnauthorized, "missing user id")
	}

	postID, err := uuid.Parse(chi.URLParam(r, "postID"))

	if err != nil {
		return commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid post id")
	}

	post, err := h.uc.GetPostByID(r.Context(), postID)

	if err != nil {
		return err
	}

	if post.AuthorID == userID {
		return commonapperr.BadRequest(apperror.CannotVoteOwnPost().Error(), "cannot like your own post")
	}

	post, err = h.uc.LikePost(r.Context(), postID, userID)

	if err != nil {
		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, post)
}

func (h *PostHandler) DislikePost(w http.ResponseWriter, r *http.Request) error {
	userID, ok := commonmiddleware.UserIDFromContext(r.Context())

	if !ok {
		return commonapperr.Unauthorized(commonapperr.CodeUnauthorized, "missing user id")
	}

	postID, err := uuid.Parse(chi.URLParam(r, "postID"))

	if err != nil {
		return commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid post id")
	}

	post, err := h.uc.GetPostByID(r.Context(), postID)

	if err != nil {
		return err
	}

	if post.AuthorID == userID {
		return commonapperr.BadRequest(apperror.CannotVoteOwnPost().Error(), "cannot like your own post")
	}

	post, err = h.uc.DislikePost(r.Context(), postID, userID)

	if err != nil {
		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, post)
}

func (h *PostHandler) RemoveVote(w http.ResponseWriter, r *http.Request) error {
	userID, ok := commonmiddleware.UserIDFromContext(r.Context())

	if !ok {
		return commonapperr.Unauthorized(commonapperr.CodeUnauthorized, "missing user id")
	}

	postID, err := uuid.Parse(chi.URLParam(r, "postID"))

	if err != nil {
		return commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid post id")
	}

	post, err := h.uc.RemovePostVote(r.Context(), postID, userID)

	if err != nil {
		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, post)
}

func parseOptionalIntQuery(r *http.Request, key string) (int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return 0, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid "+key+" - must be an integer")
	}

	return value, nil
}
