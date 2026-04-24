package friendship

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	friendshipv1 "github.com/rockkley/pushpost/services/friendship_service/gen/friendshipv1"
)

type GRPCClient struct {
	conn   *grpc.ClientConn
	client friendshipv1.FriendshipServiceClient
}

func NewGRPCClient(addr string, useTLS bool) (*GRPCClient, error) {
	if addr == "" {
		return nil, fmt.Errorf("friendship grpc addr cannot be empty")
	}

	var creds credentials.TransportCredentials
	if useTLS {
		creds = credentials.NewClientTLSFromCert(nil, "")
	} else {
		creds = insecure.NewCredentials()
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("dial friendship service: %w", err)
	}
	return &GRPCClient{
		conn:   conn,
		client: friendshipv1.NewFriendshipServiceClient(conn),
	}, nil
}

func (c *GRPCClient) Close() error {
	return c.conn.Close()
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
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}
