package friendship_api

type RelationshipResponse struct {
	AreFriends             bool `json:"are_friends"`
	PendingRequestSent     bool `json:"pending_request_sent"`
	PendingRequestReceived bool `json:"pending_request_received"`
}

func ResolveStatus(rel RelationshipResponse) string {
	switch {
	case rel.AreFriends:
		return "friends"
	case rel.PendingRequestSent:
		return "request_sent"
	case rel.PendingRequestReceived:
		return "request_received"
	default:
		return "not_friends"
	}
}
