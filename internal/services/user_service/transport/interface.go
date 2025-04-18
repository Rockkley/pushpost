package transport

import (
	"github.com/gofiber/fiber/v2"
)

type AuthHandler interface {
	RegisterUser(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	VerifyEmailOTP(c *fiber.Ctx) error
	SendNewOTP(c *fiber.Ctx) error
}
type UserHandler interface {
	GetUserByUUID(c *fiber.Ctx) error
	GetUserByEmail(c *fiber.Ctx) error
	GetByToken(c *fiber.Ctx) error
	GetFriends(c *fiber.Ctx) error
	DeleteFriend(c *fiber.Ctx) error
}

type FriendshipHandler interface {
	CreateFriendshipRequest(c *fiber.Ctx) error
	FindFriendshipRequestsByRecipientUUID(c *fiber.Ctx) error
	UpdateFriendshipRequestStatus(c *fiber.Ctx) error
	DeleteFriendshipRequest(c *fiber.Ctx) error
	AcceptFriendshipRequest(c *fiber.Ctx) error
	DeclineFriendshipRequest(c *fiber.Ctx) error
	FindFriendshipRequest(c *fiber.Ctx) error
	FindIncomingFriendshipRequests(c *fiber.Ctx) error
}

type Handler interface {
	Handler()
}
