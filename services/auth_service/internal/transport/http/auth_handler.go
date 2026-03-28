package http

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/auth_service/internal/domain"
	"github.com/rockkley/pushpost/services/auth_service/internal/domain/dto"
	"github.com/rockkley/pushpost/services/auth_service/internal/transport/http/middleware"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/httperror"
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

		return commonapperr.BadRequest(commonapperr.CodeValidationFailed, "invalid JSON")
	}

	validationErrors := dto.ValidateRegisterUser(req)

	if len(validationErrors) > 0 {
		errMap := make(map[string]string)
		for _, err := range validationErrors {
			errMap[err.Field] = err.Code
		}

		return commonapperr.ValidationFields(errMap)
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

		return commonapperr.BadRequest(commonapperr.CodeValidationFailed, "invalid JSON")
	}

	err := req.Validate()

	if err != nil {
		return commonapperr.BadRequest(commonapperr.CodeValidationFailed, err.Error())
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

		return commonapperr.Unauthorized(commonapperr.CodeUnauthorized, "invalid session")
	}

	if err := h.authUseCase.Logout(r.Context(), sessionID); err != nil {

		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]string{
		"auth": "logged out successfully",
	})
}
