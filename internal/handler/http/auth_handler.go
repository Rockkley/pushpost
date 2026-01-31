package http

import (
	"encoding/json"
	"github.com/rockkley/pushpost/internal/handler/http/dto"
	"github.com/rockkley/pushpost/internal/service"
	"github.com/rockkley/pushpost/pkg/validator"
	"net/http"
)

type AuthHandler struct {
	authService service.AuthService
}

type ErrorResponse struct {
	Field string `json:"field"`
	Code  string `json:"code"`
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) error {
	var req dto.RegisterUserDto

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return InvalidJSON()
	}

	validationErrors := validator.ValidateRegisterUser(req)

	if len(validationErrors) > 0 {
		errMap := make(map[string]string)
		for _, e := range validationErrors {
			errMap[e.Field] = e.Code
		}
		return InvalidRequestData(errMap)
	}

	user, err := h.authService.Register(r.Context(), req)

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusCreated, user)

}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) error {
	var req dto.LoginUserDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {

		return InvalidJSON()
	}

	if req.DeviceID == "" {
		req.DeviceID = "web-browser"
	}

	token, err := h.authService.Login(r.Context(), req)

	if err != nil {
		return WriteJSON(w, http.StatusUnauthorized, err.Error())
	}

	WriteJSON(w, http.StatusOK, map[string]string{"token": token})
	return nil
}
