package apperror

import (
	"github.com/rockkley/pushpost/services/common_service/apperror"
)

func AlreadyFriends() apperror.AppError {
	return apperror.BadRequest(CodeAlreadyFriends, "users are already friends")
}

func NotFriends() apperror.AppError {
	return apperror.BadRequest(CodeNotFriends, "users are not friends")
}

func CannotBefriendSelf() apperror.AppError {
	return apperror.BadRequest(CodeCannotBefriendSelf, "cannot send friend request to yourself")
}

func FriendRequestExists() apperror.AppError {
	return apperror.BadRequest(CodeFriendRequestExists, "friend request already exists")
}

func FriendRequestNotFound() apperror.AppError {
	return apperror.NotFound(CodeFriendRequestNotFound, "friend request not found")
}

func FriendRequestNotPending() apperror.AppError {
	return apperror.BadRequest(CodeFriendRequestNotPending, "friend request is no longer pending")
}

func RequestCooldown() apperror.AppError {
	return apperror.BadRequest(CodeRequestCooldown, "you must wait before sending another request to this user")
}

// -- Postgres constraint mapper

func MapConstraint(constraintName string) apperror.AppError {
	switch constraintName {
	case "friendships_no_self":
		return CannotBefriendSelf()
	case "friendships_unique":
		return AlreadyFriends()
	case "friendship_requests_unique_pending":
		return FriendRequestExists()
	case "friendships_ordered_users":
		return apperror.Internal("friendship ordering constraint violated", nil)
	default:
		return nil // return nil = pass control to generic mapper
	}
}
