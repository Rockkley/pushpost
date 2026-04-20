package entity

const EventMessageSent = "message.sent"

type MessageSentEvent struct {
	MessageID  string `json:"message_id"`
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
	CreatedAt  string `json:"created_at"`
}
