package dto

import (
	"errors"
	"github.com/google/uuid"
)

const (
	MinContentLength = 1
	MaxContentLength = 5000
)

type CreatePostDTO struct {
	AuthorID uuid.UUID
	Content  string
}

func (d *CreatePostDTO) Validate() error {
	if d.AuthorID == uuid.Nil {

		return errors.New("author_id is required")
	}

	l := len([]rune(d.Content))

	if l < MinContentLength {

		return errors.New("content cannot be empty")
	}

	if l > MaxContentLength {

		return errors.New("content is too long")
	}

	return nil
}

type GetFeedDTO struct {
	UserID uuid.UUID
	Limit  int
	// Cursor: "2024-01-15T10:00:00Z,<uuid>" - ISO timestamp + post UUID
	// Пустой = первая страница
	Cursor string
}
