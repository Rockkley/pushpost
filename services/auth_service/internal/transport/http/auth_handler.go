package http

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/auth_service/internal/domain"
	"github.com/rockkley/pushpost/services/auth_service/internal/domain/dto"
	httpDto "github.com/rockkley/pushpost/services/auth_service/internal/transport/http/dto"
	"github.com/rockkley/pushpost/services/auth_service/internal/transport/http/middleware"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/httperror"
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
		errMap := make(map[string]string, len(validationErrors))

		for _, e := range validationErrors {
			errMap[e.Field] = e.Code
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

	if err := req.Validate(); err != nil {
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

	return httperror.WriteJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) error {
	var req httpDto.VerifyEmailRequestDTO

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return commonapperr.BadRequest(commonapperr.CodeValidationFailed, "invalid JSON")
	}

	if err := req.Validate(); err != nil {
		return commonapperr.BadRequest(commonapperr.CodeValidationFailed, err.Error())
	}

	if err := h.authUseCase.VerifyEmail(r.Context(), req.Email, req.Code); err != nil {
		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]string{"message": "email verified successfully"})
}

func (h *AuthHandler) ResendOTP(w http.ResponseWriter, r *http.Request) error {
	var req httpDto.ResendOTPRequestDTO

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return commonapperr.BadRequest(commonapperr.CodeValidationFailed, "invalid JSON")
	}

	if err := req.Validate(); err != nil {
		return commonapperr.BadRequest(commonapperr.CodeValidationFailed, err.Error())
	}

	if err := h.authUseCase.ResendOTP(r.Context(), req.Email); err != nil {
		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "if this email is registered and unverified, you will receive a new code",
	})
}
