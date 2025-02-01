package transport

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"pushpost/internal/services/user_service/domain"
	"pushpost/internal/services/user_service/domain/dto"
	"pushpost/internal/services/user_service/entity"
)

type UserHandler struct {
	UserUseCase domain.UserUseCase `bind:"*usecase.UserUseCase"`
}

func NewUserHandler(useCase domain.UserUseCase) *UserHandler {
	return &UserHandler{UserUseCase: useCase}
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

	err := h.UserUseCase.RegisterUser(&params)

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

	user, err := h.UserUseCase.GetByUUID(body.UUID)

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

	user, err := h.UserUseCase.GetByEmail(body.Email)

	if err != nil {

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(200).JSON(user)
}

func (h *UserHandler) GetByToken(c *fiber.Ctx) error {
	userUUID := c.Locals("userUUID").(uuid.UUID)

	user, err := h.UserUseCase.GetByUUID(userUUID)

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
			"error": "invalid request format:" + err.Error(),
		})
	}

	token, err := h.UserUseCase.Login(loginRequest)

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

func (h *UserHandler) GetFriends(c *fiber.Ctx) error {
	//var userUUID struct {
	//	Uuid string `json:"uuid"`
	//}
	//
	//if err := c.BodyParser(&userUUID); err != nil {
	//
	//	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	//		"error": "invalid request format",
	//	})
	//}
	//token := strings.Split(c.GetReqHeaders()["Authorization"][0], " ")[1]
	//fmt.Println(token)
	//userUuid, err := jwt.VerifyToken(token, "shenanigans")
	//fmt.Println(userUuid)
	//os.Exit(1)
	userUUID := c.Locals("userUUID").(uuid.UUID)
	friends, err := h.UserUseCase.GetFriends(userUUID)

	if err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(friends)

}

func (h *UserHandler) DeleteFriend(c *fiber.Ctx) error {
	var data struct {
		FriendEmail string `json:"friendEmail"`
	}

	if err := c.BodyParser(&data); err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request format",
		})
	}
	userUUID := c.Locals("userUUID").(uuid.UUID)

	prop := dto.DeleteFriendDTO{UserUUID: userUUID, FriendEmail: data.FriendEmail}
	err := h.UserUseCase.DeleteFriend(&prop)

	if err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Friendship destroyed successfully"})
}

func (h *UserHandler) AddFriend(c *fiber.Ctx) error {
	var friendshipRequest struct {
		FriendEmail string `json:"friendEmail"`
	}

	if err := c.BodyParser(&friendshipRequest); err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request format",
		})
	}

	userUUID := c.Locals("userUUID").(uuid.UUID)
	err := h.UserUseCase.AddFriend(userUUID, friendshipRequest.FriendEmail)

	if err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Friendship created successfully"})

}

func (h *UserHandler) Handler() {}
