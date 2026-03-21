package profile_api

import "context"

type Client interface {
	GetByUsername(ctx context.Context, username string) (ProfileResponse, error)
}
