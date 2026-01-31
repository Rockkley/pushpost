package domain

type Session struct {
	SessionID string
	UserID    string
	DeviceID  string
	Expires   int64
}
