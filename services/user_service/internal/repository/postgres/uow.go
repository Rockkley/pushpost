package postgres

import (
	"context"
	"database/sql"
	"fmt"
	outbox2 "github.com/rockkley/pushpost/services/common_service/outbox"
	"github.com/rockkley/pushpost/services/user_service/internal/domain"

	outboxpg "github.com/rockkley/pushpost/services/common_service/outbox/postgres"
	"github.com/rockkley/pushpost/services/user_service/internal/repository"
)

type uowTx struct {
	users  repository.UserRepositoryInterface
	outbox outbox2.WriterInterface
}

func (t *uowTx) Users() repository.UserRepositoryInterface { return t.users }
func (t *uowTx) Outbox() outbox2.WriterInterface           { return t.outbox }

type UnitOfWork struct {
	db *sql.DB
}

func NewUnitOfWork(db *sql.DB) *UnitOfWork {
	return &UnitOfWork{db: db}
}

func (u *UnitOfWork) Reader() repository.UserRepositoryInterface {
	return NewUserRepository(u.db)
}

func (u *UnitOfWork) Do(ctx context.Context, fn func(domain.Tx) error) error {
	sqlTx, err := u.db.BeginTx(ctx, nil)

	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer sqlTx.Rollback()

	t := &uowTx{
		users:  NewUserRepository(sqlTx),
		outbox: outboxpg.NewWriterRepository(sqlTx),
	}

	if err = fn(t); err != nil {
		return err
	}

	if err = sqlTx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
