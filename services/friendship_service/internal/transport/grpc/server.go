package grpc

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	friendshipv1 "github.com/rockkley/pushpost/services/friendship_service/gen/friendship/v1"
	"github.com/rockkley/pushpost/services/friendship_service/internal/domain"
)

type FriendshipServer struct {
	friendshipv1.UnimplementedFriendshipServiceServer
	uc  domain.FriendshipUseCase
	log *slog.Logger
}

func NewFriendshipServer(uc domain.FriendshipUseCase, log *slog.Logger) *FriendshipServer {
	return &FriendshipServer{uc: uc, log: log}
}

func (s *FriendshipServer) AreFriends(
	ctx context.Context,
	req *friendshipv1.AreFriendsRequest,
) (*friendshipv1.AreFriendsResponse, error) {
	user1, err := uuid.Parse(req.User1Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user1_id: %v", err)
	}
	user2, err := uuid.Parse(req.User2Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user2_id: %v", err)
	}

	ok, err := s.uc.AreFriends(ctx, user1, user2)
	if err != nil {
		s.log.Error("AreFriends failed", slog.Any("error", err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &friendshipv1.AreFriendsResponse{AreFriends: ok}, nil
}

func (s *FriendshipServer) GetFriendIDs(
	ctx context.Context,
	req *friendshipv1.GetFriendIDsRequest,
) (*friendshipv1.GetFriendIDsResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	ids, err := s.uc.GetFriendIDs(ctx, userID)
	if err != nil {
		s.log.Error("GetFriendIDs failed", slog.Any("error", err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	strIDs := make([]string, len(ids))
	for i, id := range ids {
		strIDs[i] = id.String()
	}

	return &friendshipv1.GetFriendIDsResponse{FriendIds: strIDs}, nil
}
