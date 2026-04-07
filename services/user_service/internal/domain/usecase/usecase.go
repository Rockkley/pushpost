package usecase

import (
	"context"
	"encoding/json"
	"github.com/rockkley/pushpost/services/common_service/ctxlog"
	"github.com/rockkley/pushpost/services/common_service/outbox"
	"github.com/rockkley/pushpost/services/user_service/internal/domain"
	"github.com/rockkley/pushpost/services/user_service/internal/reserved"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	apperr "github.com/rockkley/pushpost/services/user_service/internal/apperror"
	"github.com/rockkley/pushpost/services/user_service/internal/domain/dto"
	"github.com/rockkley/pushpost/services/user_service/internal/entity"
)

type UserUseCase struct {
	uow domain.UnitOfWorkInterface
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

	if reserved.IsReserved(req.Username) {
		return nil, apperr.UsernameReserved()
	}

	user := &entity.User{
		ID:           uuid.New(),
		Username:     strings.ToLower(strings.TrimSpace(req.Username)),
		Email:        req.Email,
		PasswordHash: req.PasswordHash,
	}

	inner, err := json.Marshal(domain.UserCreatedEvent{
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
	})

	type envelope struct {
		EventType string          `json:"event_type"`
		Payload   json.RawMessage `json:"payload"`
	}
	payload, err := json.Marshal(envelope{
		EventType: "user.created",
		Payload:   inner,
	})

	if err != nil {
		return nil, commonapperr.Internal("marshal user.created event", err)
	}

	err = u.uow.Do(ctx, func(tx domain.Tx) error {
		if err = tx.Users().Create(ctx, user); err != nil {

			return err
		}

		return tx.Outbox().Insert(ctx, &outbox.OutboxEvent{
			ID:            uuid.New(),
			AggregateID:   user.ID.String(),
			AggregateType: "user",
			EventType:     "user",
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

	if email == "" {

		return nil, commonapperr.Validation(
			commonapperr.CodeFieldRequired, "email", "email is required",
		)
	}

	user, err := u.uow.Reader().GetUserByEmail(ctx, email)

	if err != nil {
		log.Debug("failed to get user by email", slog.Any("error", err))

		return nil, err
	}

	return user, nil
}

func (u *UserUseCase) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	log := ctxlog.From(ctx).With(
		slog.String("op", "UserUseCase.GetUserByID"),
		slog.String("user_id", id.String()),
	)

	// FindByID not filtering deleted_at in SQL
	// we check it explicitly in the use case to distinguish "not found" and "deleted" at the business logic level
	user, err := u.uow.Reader().FindByID(ctx, id)

	if err != nil {
		log.Debug("user not found", slog.Any("error", err))

		return nil, err
	}

	if user.IsDeleted() {
		log.Warn("attempt to access deleted account")

		return nil, apperr.UserDeleted()
	}

	return user, nil
}

func (u *UserUseCase) DeleteUser(ctx context.Context, id uuid.UUID) error {
	log := ctxlog.From(ctx).With(
		slog.String("op", "UserUseCase.DeleteUser"),
		slog.String("user_id", id.String()),
	)

	err := u.uow.Do(ctx, func(tx domain.Tx) error {
		user, err := tx.Users().FindByID(ctx, id)

		if err != nil {

			return err
		}

		if user.IsDeleted() {

			return apperr.UserNotFound()
		}

		return tx.Users().SoftDelete(ctx, id)
	})

	if err != nil {
		log.Error("failed to delete user", slog.Any("error", err))

		return err
	}

	log.Info("user deleted")

	return nil
}

func (u *UserUseCase) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	log := ctxlog.From(ctx).With(
		slog.String("op", "UserUseCase.GetUserByUsername"),
		slog.String("username", username),
	)

	if username == "" {

		return nil, commonapperr.Validation(
			commonapperr.CodeFieldRequired, "username", "username is required",
		)
	}

	user, err := u.uow.Reader().FindByUsername(ctx, username)

	if err != nil {

		log.Debug("user not found by username", slog.Any("error", err))
		return nil, err
	}

	if user.IsDeleted() {

		return nil, apperr.UserNotFound()
	}

	return user, nil
}
