package events

const (
	EventPostCreated    = "post.created"
	EventPostUpdated    = "post.updated"
	EventPostDeleted    = "post.deleted"
	EventCommentReplied = "comment.replied"
	EventCommentMention = "comment.mentioned"
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

type CommentRepliedEvent struct {
	PostID           string `json:"post_id"`
	CommentID        string `json:"comment_id"`
	ParentCommentID  string `json:"parent_comment_id"`
	ReplyAuthorID    string `json:"reply_author_id"`
	OriginalAuthorID string `json:"original_author_id"`
}

type CommentMentionedEvent struct {
	PostID        string   `json:"post_id"`
	CommentID     string   `json:"comment_id"`
	AuthorID      string   `json:"author_id"`
	MentionedList []string `json:"mentioned_list"`
}
