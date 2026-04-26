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
