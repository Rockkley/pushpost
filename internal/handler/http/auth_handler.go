package http

import "github.com/rockkley/pushpost/internal/service"

type AuthHandler struct {
	authService *service.AuthService
}

type ErrorResponse struct {
	Field string `json:"field"`
	Code  string `json:"code"`
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register()
