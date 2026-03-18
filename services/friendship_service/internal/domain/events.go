package domain

const (
	EventFriendRequestSent      = "friendship_request.sent"
	EventFriendRequestRejected  = "friendship_request.rejected"
	EventFriendRequestCancelled = "friendship_request.cancelled"
	EventFriendshipCreated      = "friendship.created"
	EventFriendshipDeleted      = "friendship.deleted"
)

// EventFriendRequestSent
type FriendRequestSentPayload struct {
	RequestID  string `json:"request_id"`
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
}

// EventFriendshipCreated
type FriendshipCreatedPayload struct {
	FriendshipID string `json:"friendship_id"`
	User1ID      string `json:"user1_id"`
	User2ID      string `json:"user2_id"`
}

// EventFriendRequestRejected
type FriendRequestRejectedPayload struct {
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
}

// EventFriendRequestCancelled
type FriendRequestCancelledPayload struct {
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
}

// EventFriendshipDeleted
type FriendshipDeletedPayload struct {
	UserID   string `json:"user_id"`
	FriendID string `json:"friend_id"`
}
