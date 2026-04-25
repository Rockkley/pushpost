package email

import "context"

type Sender interface {
	SendOTP(ctx context.Context, to, code string) error
}
