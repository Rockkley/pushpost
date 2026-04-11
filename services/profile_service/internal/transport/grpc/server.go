package grpc

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	profilev1 "github.com/rockkley/pushpost/services/profile_service/gen/profile/v1"
	"github.com/rockkley/pushpost/services/profile_service/internal/domain"
)

type ProfileServer struct {
	profilev1.UnimplementedProfileServiceServer
	uc  domain.ProfileUseCaseInterface
	log *slog.Logger
}

func NewProfileServer(uc domain.ProfileUseCaseInterface, log *slog.Logger) *ProfileServer {
	return &ProfileServer{uc: uc, log: log}
}

func (s *ProfileServer) GetProfileByUsername(
	ctx context.Context,
	req *profilev1.GetProfileByUsernameRequest,
) (*profilev1.GetProfileByUsernameResponse, error) {
	if req.Username == "" {

		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	profile, err := s.uc.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, domain.ErrProfileNotFound) {

			return nil, status.Error(codes.NotFound, "profile not found")
		}

		s.log.Error("GetProfileByUsername failed",
			slog.String("username", req.Username),
			slog.Any("error", err),
		)

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &profilev1.GetProfileByUsernameResponse{
		UserId:       profile.UserID.String(),
		Username:     profile.Username,
		CreatedAt:    profile.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		DisplayName:  derefString(profile.DisplayName),
		FirstName:    derefString(profile.FirstName),
		LastName:     derefString(profile.LastName),
		BirthDate:    formatDate(profile.BirthDate),
		AvatarUrl:    derefString(profile.AvatarURL),
		Bio:          derefString(profile.Bio),
		TelegramLink: derefString(profile.TelegramLink),
		IsPrivate:    profile.IsPrivate,
	}, nil
}

func derefString(s *string) string {
	if s == nil {

		return ""
	}

	return *s
}

func formatDate(t *time.Time) string {
	if t == nil {

		return ""
	}

	return t.Format("2006-01-02")
}
