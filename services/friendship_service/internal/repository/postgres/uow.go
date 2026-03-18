package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/rockkley/pushpost/services/common_service/outbox"
	outboxpg "github.com/rockkley/pushpost/services/common_service/outbox/postgres"
	"github.com/rockkley/pushpost/services/friendship_service/internal/domain"
	"github.com/rockkley/pushpost/services/friendship_service/internal/repository"
)

var _ domain.UnitOfWork = (*UnitOfWork)(nil)
var _ domain.Tx = (*uowTx)(nil)

type uowTx struct {
	requests    repository.FriendshipRequestRepository
	friendships repository.FriendshipRepository
	outbox      outbox.WriterInterface
}

func (u *uowTx) Requests() repository.FriendshipRequestRepository {
	return u.requests
}

func (u *uowTx) Friendships() repository.FriendshipRepository { return u.friendships }

func (u *uowTx) Outbox() outbox.WriterInterface { return u.outbox }

type UnitOfWork struct {
	db *sql.DB
}

func NewUnitOfWork(db *sql.DB) domain.UnitOfWork {
	return &UnitOfWork{db: db}
}

func (u *UnitOfWork) Do(ctx context.Context, fn func(domain.Tx) error) error {
	sqlTx, err := u.db.BeginTx(ctx, nil)

	if err != nil {

		return fmt.Errorf("begin tx: %w", err)
	}

	defer sqlTx.Rollback()

	t := &uowTx{
		requests:    NewFriendshipRequestRepository(sqlTx),
		friendships: NewFriendshipRepository(sqlTx),
		outbox:      outboxpg.NewWriterRepository(sqlTx),
	}

	if err = fn(t); err != nil {

		return err
	}

	if err = sqlTx.Commit(); err != nil {

		return fmt.Errorf("commit tx: %w", err)
	}

	return nil

}

func (u *UnitOfWork) Requests() repository.FriendshipRequestRepository {
	return NewFriendshipRequestRepository(u.db)
}

func (u *UnitOfWork) Friendships() repository.FriendshipRepository {
	return NewFriendshipRepository(u.db)
}
