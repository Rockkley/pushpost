package transport

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"pushpost/internal/services/user_service/domain/dto"
	"pushpost/internal/services/user_service/domain/usecase"
	"pushpost/internal/services/user_service/entity"
)

type UserHandler struct {
	useCase *usecase.UserUseCase
}

func RegisterUserHandler(useCase usecase.UserUseCase) *UserHandler {
	return &UserHandler{useCase: &useCase}
}

func (h *UserHandler) RegisterUser(c *fiber.Ctx) error {
	var body entity.User

	if err := c.BodyParser(&body); err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	params := dto.RegisterUserDTO{
		Name:     body.Name,
		Email:    body.Email,
		Password: body.Password,
		Age:      body.Age,
	}

	if err := params.Validate(); err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	err := h.useCase.RegisterUser(&params)

	if err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "User created successfully"})
}

func (h *UserHandler) GetUserByUUID(c *fiber.Ctx) error {
	var body entity.User

	if err := c.BodyParser(&body); err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	user, err := h.useCase.GetByUUID(body.UUID)

	if err != nil {

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusFound).JSON(user)
}

func (h *UserHandler) GetUserByEmail(c *fiber.Ctx) error {
	var body entity.User

	if err := c.BodyParser(&body); err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	user, err := h.useCase.GetByEmail(body.Email)

	if err != nil {

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(200).JSON(user)
}

func (h *UserHandler) GetByToken(c *fiber.Ctx) error {
	userUUID := c.Locals("userUUID").(uuid.UUID)

	user, err := h.useCase.GetByUUID(userUUID)

	if err != nil {

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	userData := dto.UserDataByUUID{Name: user.Name, Age: user.Age}

	return c.Status(fiber.StatusOK).JSON(userData)
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	var loginRequest dto.UserLoginDTO

	if err := c.BodyParser(&loginRequest); err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request format",
		})
	}

	token, err := h.useCase.Login(loginRequest)

	if err != nil {

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"type":  "Bearer",
	})
}

func (h *UserHandler) AddFriend(c *fiber.Ctx) error {
	var friendshipRequest struct {
		userUUID    uuid.UUID
		friendEmail string
	}

	if err := c.BodyParser(&friendshipRequest); err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request format",
		})
	}

	err := h.useCase.AddFriend(email)

	if err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Friendship created successfully"})

}
