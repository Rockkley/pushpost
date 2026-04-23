package friendship

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/rockkley/pushpost/services/friendship_service/gen/friendshipv1"
)

type GRPCClient struct {
	client friendshipv1.FriendshipServiceClient
}

func NewGRPCClient(addr string) (*GRPCClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial friendship service: %w", err)
	}
	return &GRPCClient{client: friendshipv1.NewFriendshipServiceClient(conn)}, nil
}

func (c *GRPCClient) GetFriendIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	resp, err := c.client.GetFriendIDs(ctx, &friendshipv1.GetFriendIDsRequest{
		UserId: userID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("grpc get friend ids: %w", err)
	}

	ids := make([]uuid.UUID, 0, len(resp.FriendIds))
	for _, s := range resp.FriendIds {
		id, err := uuid.Parse(s)
		if err != nil {
			continue // не ломаем ленту из-за одного битого UUID
		}
		ids = append(ids, id)
	}
	return ids, nil
}
