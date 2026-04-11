package grpc

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	profilev1 "github.com/rockkley/pushpost/services/profile_service/gen/profile/v1"
)

var ErrNotFound = errors.New("profile not found")

type ProfileResponse struct {
	UserID       string
	Username     string
	CreatedAt    string
	DisplayName  string
	FirstName    string
	LastName     string
	BirthDate    string
	AvatarURL    string
	Bio          string
	TelegramLink string
	IsPrivate    bool
}

type Client struct {
	grpc profilev1.ProfileServiceClient
}

func NewClient(addr string) (*Client, error) {
	if addr == "" {
		return nil, fmt.Errorf("profile grpc addr cannot be empty")
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial profile service: %w", err)
	}
	return &Client{grpc: profilev1.NewProfileServiceClient(conn)}, nil
}

func (c *Client) GetByUsername(ctx context.Context, username string) (ProfileResponse, error) {
	resp, err := c.grpc.GetProfileByUsername(ctx, &profilev1.GetProfileByUsernameRequest{
		Username: username,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return ProfileResponse{}, ErrNotFound
		}
		return ProfileResponse{}, fmt.Errorf("profile grpc: %w", err)
	}
	return ProfileResponse{
		UserID:       resp.UserId,
		Username:     resp.Username,
		CreatedAt:    resp.CreatedAt,
		DisplayName:  resp.DisplayName,
		FirstName:    resp.FirstName,
		LastName:     resp.LastName,
		BirthDate:    resp.BirthDate,
		AvatarURL:    resp.AvatarUrl,
		Bio:          resp.Bio,
		TelegramLink: resp.TelegramLink,
		IsPrivate:    resp.IsPrivate,
	}, nil
}
