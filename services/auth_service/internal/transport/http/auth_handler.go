package http

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/internal/handler/httperror"
	"github.com/rockkley/pushpost/services/auth_service/internal/domain"
	"github.com/rockkley/pushpost/services/auth_service/internal/transport/http/dto"
	"github.com/rockkley/pushpost/services/auth_service/internal/transport/http/middleware"
	"github.com/rockkley/pushpost/services/common/apperror"
	"net/http"
)

type AuthHandler struct {
	authUseCase domain.AuthUsecase
}

func NewAuthHandler(authUseCase domain.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUseCase: authUseCase}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) error {
	var req dto.RegisterUserDto

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {

		return apperror.BadRequest(apperror.CodeValidationFailed, "invalid JSON")
	}

	validationErrors := dto.ValidateRegisterUser(req)

	if len(validationErrors) > 0 {
		errMap := make(map[string]string)
		for _, err := range validationErrors {
			errMap[err.Field] = err.Code
		}

		return apperror.ValidationFields(errMap)
	}

	user, err := h.authUseCase.Register(r.Context(), req)

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

	token, err := h.authUseCase.Login(r.Context(), req)

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

	if err := h.authUseCase.Logout(r.Context(), sessionID); err != nil {

		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "logged out successfully",
	})
}
