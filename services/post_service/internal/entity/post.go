package entity

import (
	"github.com/google/uuid"
	"time"
)

type Post struct {
	ID            uuid.UUID  `json:"id"`
	AuthorID      uuid.UUID  `json:"author_id"`
	Content       string     `json:"content"`
	Version       int        `json:"version"`
	LikesCount    int        `json:"likes_count"`
	DislikesCount int        `json:"dislikes_count"`
	Rating        int        `json:"rating"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
	InsertedAt    time.Time  `json:"-"`
}

func (p *Post) IsDeleted() bool { return p.DeletedAt != nil }
