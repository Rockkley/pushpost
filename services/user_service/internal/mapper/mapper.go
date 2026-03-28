package mapper

import (
	"github.com/rockkley/pushpost/services/user_service/internal/domain/dto"
	reqDTO "github.com/rockkley/pushpost/services/user_service/internal/transport/http/dto"
)

func CreateUserFromRequestToUseCase(req reqDTO.CreateUserRequestDTO) *dto.CreateUserDTO {
	return &dto.CreateUserDTO{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: req.PasswordHash,
	}
}
