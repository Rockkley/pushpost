package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rockkley/pushpost/services/common_service/ctxlog"
	"github.com/rockkley/pushpost/services/common_service/outbox"
	"github.com/rockkley/pushpost/services/user_service/internal/domain"
	"log/slog"

	"github.com/google/uuid"

	"github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/user_service/internal/domain/dto"
	"github.com/rockkley/pushpost/services/user_service/internal/entity"
)

type UserUseCase struct {
	uow domain.UnitOfWorkInterface
	//userRepo repository.UserRepositoryInterface
}

func NewUserUseCase(uow domain.UnitOfWorkInterface) *UserUseCase {
	return &UserUseCase{uow: uow}
}

func (u *UserUseCase) CreateUser(ctx context.Context, req dto.CreateUserDTO) (*entity.User, error) {
	log := ctxlog.From(ctx).With(slog.String("op", "UserUseCase.CreateUser"))

	if err := req.Validate(); err != nil {
		log.Debug("dto validation failed", slog.Any("error", err))
		return nil, err
	}

	user := &entity.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: req.PasswordHash,
	}

	payload, err := json.Marshal(domain.UserCreatedEvent{
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
	})

	if err != nil {
		return nil, fmt.Errorf("marshal user.created event: %w", err)
	}

	err = u.uow.Do(ctx, func(tx domain.Tx) error {
		if err = tx.Users().Create(ctx, user); err != nil {
			return err
		}
		return tx.Outbox().Insert(ctx, &outbox.OutboxEvent{
			ID:            uuid.New(),
			AggregateID:   user.ID.String(),
			AggregateType: "user",
			EventType:     "user.created",
			Payload:       payload,
		})
	})
	if err != nil {
		log.Error("failed to create user", slog.Any("error", err))
		return nil, err
	}

	log.Info("user created", slog.String("user_id", user.ID.String()))
	return user, nil
}

func (u *UserUseCase) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	log := ctxlog.From(ctx).With(slog.String("op", "UserUseCase.GetUserByEmail"))

	user, err := u.uow.Reader().GetUserByEmail(ctx, email)
	if err != nil {
		log.Debug("user not found by email")
		return nil, apperror.NotFound(apperror.CodeUserNotFound, "user not found")
	}

	if user.IsDeleted() {
		log.Warn("attempt to find deleted account by email", slog.String("user_id", user.ID.String()))
		return nil, apperror.NotFound(apperror.CodeUserDeleted, "user is deleted")
	}

	return user, nil
}
