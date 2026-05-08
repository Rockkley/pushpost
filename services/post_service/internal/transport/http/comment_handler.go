package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/httperror"
	commonmiddleware "github.com/rockkley/pushpost/services/common_service/middleware"
	"github.com/rockkley/pushpost/services/post_service/internal/domain"
	"github.com/rockkley/pushpost/services/post_service/internal/entity"
)

type CommentHandler struct {
	uc domain.CommentUseCaseInterface
}

func NewCommentHandler(uc domain.CommentUseCaseInterface) *CommentHandler {
	return &CommentHandler{uc: uc}
}

func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) error {
	userID, ok := commonmiddleware.UserIDFromContext(r.Context())
	if !ok {
		return commonapperr.Unauthorized(commonapperr.CodeUnauthorized, "missing user id")
	}
	postID, err := uuid.Parse(chi.URLParam(r, "postID"))
	if err != nil {
		return commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid post id")
	}
	var body struct {
		Content  string  `json:"content"`
		ParentID *string `json:"parent_id,omitempty"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxPostBodySize)).Decode(&body); err != nil {
		return commonapperr.BadRequest(commonapperr.CodeValidationFailed, "invalid request body")
	}
	var parentID *uuid.UUID
	if body.ParentID != nil && *body.ParentID != "" {
		p, err := uuid.Parse(*body.ParentID)
		if err != nil {
			return commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid parent_id")
		}
		parentID = &p
	}
	c, err := h.uc.CreateComment(r.Context(), postID, userID, parentID, body.Content)
	if err != nil {
		return err
	}
	return httperror.WriteJSON(w, http.StatusCreated, c)
}

func (h *CommentHandler) GetPostComments(w http.ResponseWriter, r *http.Request) error {
	postID, err := uuid.Parse(chi.URLParam(r, "postID"))
	if err != nil {
		return commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid post id")
	}
	limit, err := parseOptionalIntQuery(r, "limit")
	if err != nil {
		return err
	}
	resp, err := h.uc.GetPostComments(r.Context(), postID, limit, r.URL.Query().Get("cursor"))
	if err != nil {
		return err
	}
	comments := resp.Comments
	if comments == nil {
		comments = []*entity.Comment{}
	}
	return httperror.WriteJSON(w, http.StatusOK, domain.CommentsResponse{Comments: comments, NextCursor: resp.NextCursor})
}

func (h *CommentHandler) UpdateComment(w http.ResponseWriter, r *http.Request) error {
	userID, ok := commonmiddleware.UserIDFromContext(r.Context())
	if !ok {
		return commonapperr.Unauthorized(commonapperr.CodeUnauthorized, "missing user id")
	}
	id, err := uuid.Parse(chi.URLParam(r, "commentID"))
	if err != nil {
		return commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid comment id")
	}
	var b struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxPostBodySize)).Decode(&b); err != nil {
		return commonapperr.BadRequest(commonapperr.CodeValidationFailed, "invalid request body")
	}
	c, err := h.uc.UpdateComment(r.Context(), id, userID, b.Content)
	if err != nil {
		return err
	}
	return httperror.WriteJSON(w, http.StatusOK, c)
}
func (h *CommentHandler) UpvoteComment(w http.ResponseWriter, r *http.Request) error {
	return h.vote(w, r, 1)
}
func (h *CommentHandler) DownvoteComment(w http.ResponseWriter, r *http.Request) error {
	return h.vote(w, r, -1)
}
func (h *CommentHandler) vote(w http.ResponseWriter, r *http.Request, v int) error {
	userID, ok := commonmiddleware.UserIDFromContext(r.Context())

	if !ok {
		return commonapperr.Unauthorized(commonapperr.CodeUnauthorized, "missing user id")
	}
	id, err := uuid.Parse(chi.URLParam(r, "commentID"))

	if err != nil {
		return commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid comment id")
	}

	var c *entity.Comment

	if v > 0 {
		c, err = h.uc.UpvoteComment(r.Context(), id, userID)
	} else {
		c, err = h.uc.DownvoteComment(r.Context(), id, userID)
	}

	if err != nil {
		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, c)
}
func (h *CommentHandler) RemoveCommentVote(w http.ResponseWriter, r *http.Request) error {
	userID, ok := commonmiddleware.UserIDFromContext(r.Context())
	if !ok {
		return commonapperr.Unauthorized(commonapperr.CodeUnauthorized, "missing user id")
	}
	id, err := uuid.Parse(chi.URLParam(r, "commentID"))

	if err != nil {
		return commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid comment id")
	}
	c, err := h.uc.RemoveCommentVote(r.Context(), id, userID)

	if err != nil {
		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, c)
}
func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) error {
	userID, ok := commonmiddleware.UserIDFromContext(r.Context())
	if !ok {
		return commonapperr.Unauthorized(commonapperr.CodeUnauthorized, "missing user id")
	}
	id, err := uuid.Parse(chi.URLParam(r, "commentID"))
	if err != nil {
		return commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid comment id")
	}
	if err := h.uc.DeleteComment(r.Context(), id, userID); err != nil {
		return err
	}
	return httperror.WriteJSON(w, http.StatusOK, map[string]string{"message": "comment deleted"})
}
