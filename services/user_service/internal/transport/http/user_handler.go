package http

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/httperror"
	"github.com/rockkley/pushpost/services/user_service/internal/domain"
	"github.com/rockkley/pushpost/services/user_service/internal/mapper"
	"github.com/rockkley/pushpost/services/user_service/internal/transport/http/dto"
	"net/http"
)

type UserHandler struct {
	userUseCase domain.UserUseCaseInterface
}

func NewUserHandler(userUseCase domain.UserUseCaseInterface) *UserHandler {

	return &UserHandler{userUseCase: userUseCase}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) error {
	var req dto.CreateUserRequestDTO

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return commonapperr.BadRequest(commonapperr.CodeValidationFailed, "invalid JSON")
	}

	if err := req.Validate(); err != nil {
		return commonapperr.BadRequest(commonapperr.CodeValidationFailed, err.Error())
	}

	user, err := h.userUseCase.CreateUser(r.Context(), *mapper.CreateUserFromRequestToUseCase(req))

	if err != nil {
		return err
	}

	return httperror.WriteJSON(w, http.StatusCreated, user)
}

func (h *UserHandler) GetUserByEmail(w http.ResponseWriter, r *http.Request) error {
	email := r.URL.Query().Get("email")

	if email == "" {
		return commonapperr.Validation(
			commonapperr.CodeFieldRequired, "email", "email query parameter is required",
		)
	}

	user, err := h.userUseCase.GetUserByEmail(r.Context(), email)

	if err != nil {
		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, user)

}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) error {
	rawID := chi.URLParam(r, "id")
	id, err := uuid.Parse(rawID)

	if err != nil {
		return commonapperr.BadRequest(commonapperr.CodeFieldInvalid, "invalid user id")
	}

	user, err := h.userUseCase.GetUserByID(r.Context(), id)

	if err != nil {
		return err // UserNotFound / UserDeleted
	}

	return httperror.WriteJSON(w, http.StatusOK, user)
}

func (h *UserHandler) GetUserByUsername(w http.ResponseWriter, r *http.Request) error {
	username := chi.URLParam(r, "username")

	if username == "" {
		return commonapperr.Validation(
			commonapperr.CodeFieldRequired, "username", "username is required",
		)
	}

	user, err := h.userUseCase.GetUserByUsername(r.Context(), username)

	if err != nil {
		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, user)
}

func (h *UserHandler) ActivateUser(w http.ResponseWriter, r *http.Request) error {
	var body struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return commonapperr.BadRequest(commonapperr.CodeValidationFailed, "invalid JSON")
	}

	if body.Email == "" {
		return commonapperr.Validation(commonapperr.CodeFieldRequired, "email", "email is required")
	}

	if err := h.userUseCase.ActivateUser(r.Context(), body.Email); err != nil {
		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]string{"message": "user activated"})
}
