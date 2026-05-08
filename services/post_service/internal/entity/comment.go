package entity

import (
	"github.com/google/uuid"
	"time"
)

type Comment struct {
	ID             uuid.UUID  `json:"id"`
	PostID         uuid.UUID  `json:"post_id"`
	AuthorID       uuid.UUID  `json:"author_id"`
	ParentID       *uuid.UUID `json:"parent_id,omitempty"`
	ReplyToUserID  *uuid.UUID `json:"reply_to_user_id,omitempty"`
	Content        string     `json:"content"`
	UpvotesCount   int        `json:"upvotes_count"`
	DownvotesCount int        `json:"downvotes_count"`
	Rating         int        `json:"rating"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
