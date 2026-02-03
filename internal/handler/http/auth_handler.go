package http

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/internal/apperror"
	"github.com/rockkley/pushpost/internal/handler/http/dto"
	"github.com/rockkley/pushpost/internal/handler/http/middleware"
	"github.com/rockkley/pushpost/internal/handler/httperror"
	"github.com/rockkley/pushpost/internal/service"
	"github.com/rockkley/pushpost/pkg/validator"
	"net/http"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) error {
	var req dto.RegisterUserDto

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apperror.BadRequest(apperror.CodeValidationFailed, "invalid JSON")
	}

	validationErrors := validator.ValidateRegisterUser(req)

	if len(validationErrors) > 0 {
		errMap := make(map[string]string)
		for _, err := range validationErrors {
			errMap[err.Field] = err.Code
		}
		return apperror.ValidationFields(errMap)
	}

	user, err := h.authService.Register(r.Context(), req)

	if err != nil {
		return err
	}

	return httperror.WriteJSON(w, http.StatusCreated, user)

}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) error {
	var req dto.LoginUserDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {

		return apperror.BadRequest(apperror.CodeValidationFailed, "invalid JSON")
	}

	token, err := h.authService.Login(r.Context(), req)

	if err != nil {
		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]string{"token": token})

}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) error {
	sessionID, ok := r.Context().Value(middleware.CtxSessionIDKey).(uuid.UUID)
	if !ok {
		return apperror.Unauthorized(apperror.CodeUnauthorized, "invalid session")
	}

	if err := h.authService.Logout(r.Context(), sessionID); err != nil {
		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "logged out successfully",
	})
}
