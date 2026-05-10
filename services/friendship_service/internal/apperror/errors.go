package apperror

import (
	"fmt"
	"github.com/rockkley/pushpost/services/common_service/apperror"
	"time"
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

func RequestCooldown(time time.Time) apperror.AppError {
	return apperror.BadRequest(CodeRequestCooldown, fmt.Sprintf("you must wait %v before sending another request to this user", time))
}

func CannotBlockSelf() apperror.AppError {
	return apperror.BadRequest(CodeCannotBlockSelf, "cannot block yourself")
}

func AlreadyBlocked() apperror.AppError {
	return apperror.BadRequest(CodeAlreadyBlocked, "user is already blocked")
}

func BlockNotFound() apperror.AppError {
	return apperror.NotFound(CodeBlockNotFound, "block not found")
}

func UserBlocked() apperror.AppError {
	return apperror.BadRequest(CodeUserBlocked, "you cannot perform this action because the user has blocked you")
}

// -- Postgres constraint mapper

func MapConstraint(constraintName string) apperror.AppError {
	switch constraintName {
	case "friendship_requests_no_self":
		return CannotBefriendSelf()
	case "friendships_unique_pair":
		return AlreadyFriends()
	case "friendship_requests_unique_pending":
		return FriendRequestExists()
	case "friendships_ordered_users":
		return apperror.Internal("friendship ordering constraint violated", nil)
	case "chk_block_not_self":
		return CannotBlockSelf()
	case "blocks_pkey":
		return AlreadyBlocked()

	default:
		return nil // return nil = pass control to generic mapper
	}
}
