package kafka

const (
	TopicFriendRequestSent     = "friendship_request.sent"
	TopicFriendshipCreated     = "friendship.created"
	TopicFriendRequestRejected = "friendship_request.rejected"
	TopicMessageSent           = "message.sent"
)

var ConsumedTopics = []string{
	TopicFriendRequestSent,
	TopicFriendshipCreated,
	TopicFriendRequestRejected,
	TopicMessageSent,
}
