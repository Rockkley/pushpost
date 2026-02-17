package dto

import "github.com/google/uuid"

type AcceptRequestDTO struct {
	SenderID   uuid.UUID `json:"senderID"`
	ReceiverID uuid.UUID `json:"receiverID"`
}

type RejectRequestDTO struct {
	SenderID   uuid.UUID `json:"senderID"`
	ReceiverID uuid.UUID `json:"receiverID"`
}

type CancelRequestDTO struct {
	SenderID   uuid.UUID `json:"senderID"`
	ReceiverID uuid.UUID `json:"receiverID"`
}

type DeleteFriendshipDTO struct {
	UserID   uuid.UUID `json:"userID"`
	FriendID uuid.UUID `json:"friendID"`
}

type SendRequestDTO struct {
	SenderID   uuid.UUID `json:"senderID"`
	ReceiverID uuid.UUID `json:"receiverID"`
}
