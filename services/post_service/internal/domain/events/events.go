package events

const (
	EventPostCreated = "post.created"
	EventPostUpdated = "post.updated"
	EventPostDeleted = "post.deleted"
)

type PostCreatedEvent struct {
	PostID    string `json:"post_id"`
	AuthorID  string `json:"author_id"`
	CreatedAt string `json:"created_at"` // RFC3339Nano - inserted_at in feeds
}

type PostUpdatedEvent struct {
	PostID  string `json:"post_id"`
	Version int    `json:"version"`
}

type PostDeletedEvent struct {
	PostID   string `json:"post_id"`
	AuthorID string `json:"author_id"`
}
