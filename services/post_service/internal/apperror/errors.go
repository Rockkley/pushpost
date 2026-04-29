package apperror

import "github.com/rockkley/pushpost/services/common_service/apperror"

func PostNotFound() apperror.AppError {
	return apperror.NotFound(CodePostNotFound, "post not found")
}

func ContentEmpty() apperror.AppError {
	return apperror.Validation(CodeContentEmpty, "content", "content cannot be empty")
}

func ContentTooLong() apperror.AppError {
	return apperror.Validation(CodeContentTooLong, "content", "content exceeds maximum length of 5000 characters")
}

func NotPostAuthor() apperror.AppError {
	return apperror.Forbidden(CodeNotPostAuthor, "you are not the author of this post")
}

func CannotVoteOwnPost() apperror.AppError {
	return apperror.Validation(CodeCannotVoteOwnPost, "like", "you cannot like your own post")
}
