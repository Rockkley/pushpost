package domain

type FriendRequestSentPayload struct {
	RequestID  string `json:"request_id"`
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
}

type FriendshipCreatedPayload struct {
	FriendshipID string `json:"friendship_id"`
	User1ID      string `json:"user1_id"`
	User2ID      string `json:"user2_id"`
}

type FriendRequestRejectedPayload struct {
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
}

type MessageSentPayload struct {
	MessageID  string `json:"message_id"`
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
	CreatedAt  string `json:"created_at"`
}

type CommentRepliedPayload struct {
	PostID           string `json:"post_id"`
	CommentID        string `json:"comment_id"`
	ParentCommentID  string `json:"parent_comment_id"`
	ReplyAuthorID    string `json:"reply_author_id"`
	OriginalAuthorID string `json:"original_author_id"`
}

type CommentMentionedPayload struct {
	PostID        string   `json:"post_id"`
	CommentID     string   `json:"comment_id"`
	AuthorID      string   `json:"author_id"`
	MentionedList []string `json:"mentioned_list"`
}
