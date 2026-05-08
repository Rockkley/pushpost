package kafka

const (
	TopicFriendRequestSent     = "friendship_request.sent"
	TopicFriendshipCreated     = "friendship.created"
	TopicFriendRequestRejected = "friendship_request.rejected"
	TopicMessageSent           = "message.sent"
	TopicCommentReplied        = "comment.replied"
	TopicCommentMentioned      = "comment.mentioned"
)

var ConsumedTopics = []string{
	TopicFriendRequestSent,
	TopicFriendshipCreated,
	TopicFriendRequestRejected,
	TopicMessageSent,
	TopicCommentReplied,
	TopicCommentMentioned,
}
