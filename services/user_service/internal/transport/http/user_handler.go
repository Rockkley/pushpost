package http

import (
	"encoding/json"
	"fmt"
	"github.com/rockkley/pushpost/internal/handler/httperror"
	"github.com/rockkley/pushpost/services/user_service/internal/domain"
	"github.com/rockkley/pushpost/services/user_service/internal/mapper"
	"github.com/rockkley/pushpost/services/user_service/internal/transport/http/dto"
	"net/http"
)

type UserHandler struct {
	userUseCase domain.UserUseCase
}

func NewUserHandler(userUseCase domain.UserUseCase) *UserHandler {
	return &UserHandler{userUseCase: userUseCase}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) error {
	var req dto.CreateUserRequestDTO

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {

		return err
	}

	if err := req.Validate(); err != nil {

		return err
	}

	mappedDTO := mapper.CreateUserFromRequestToUseCase(req)

	user, err := h.userUseCase.CreateUser(r.Context(), *mappedDTO)

	if err != nil {
		return err
	}

	return httperror.WriteJSON(w, http.StatusCreated, user)
}

func (h *UserHandler) AuthenticateUser(w http.ResponseWriter, r *http.Request) error {
	var req dto.AuthenticateUserRequestDTO

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {

		return err
	}
	fmt.Println(req)
	if err := req.Validate(); err != nil {
		return err
	}

	mappedDTO := mapper.AuthUserFromRequestToUseCase(req)

	user, err := h.userUseCase.AuthenticateUser(r.Context(), *mappedDTO)

	if err != nil {
		return err
	}

	return httperror.WriteJSON(w, http.StatusOK, user)

}
