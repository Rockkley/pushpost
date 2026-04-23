package events

const EventPostCreated = "post.created"
const EventPostDeleted = "post.deleted"

type PostCreatedEvent struct {
	PostID   string `json:"post_id"`
	AuthorID string `json:"author_id"`
}

type PostDeletedEvent struct {
	PostID   string `json:"post_id"`
	AuthorID string `json:"author_id"`
}
