package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rockkley/pushpost/services/common_service/outbox"
	outboxpg "github.com/rockkley/pushpost/services/common_service/outbox/postgres"
	"github.com/rockkley/pushpost/services/message_service/internal/domain"
	"github.com/rockkley/pushpost/services/message_service/internal/repository"
)

// compile-time interface checks
var _ domain.UnitOfWork = (*UnitOfWork)(nil)
var _ domain.Tx = (*uowTx)(nil)

type uowTx struct {
	messages repository.MessageRepository
	outbox   outbox.WriterInterface
}

func (t *uowTx) Messages() repository.MessageRepository { return t.messages }
func (t *uowTx) Outbox() outbox.WriterInterface         { return t.outbox }

type UnitOfWork struct {
	db *sql.DB
}

func NewUnitOfWork(db *sql.DB) *UnitOfWork {
	return &UnitOfWork{db: db}
}

func (u *UnitOfWork) Reader() repository.MessageRepository {
	return NewMessageRepository(u.db)
}

func (u *UnitOfWork) Do(ctx context.Context, fn func(domain.Tx) error) error {
	sqlTx, err := u.db.BeginTx(ctx, nil)

	if err != nil {

		return fmt.Errorf("begin tx: %w", err)
	}
	defer sqlTx.Rollback()

	t := &uowTx{
		messages: NewMessageRepository(sqlTx),
		outbox:   outboxpg.NewWriterRepository(sqlTx),
	}

	if err = fn(t); err != nil {

		return err
	}

	if err = sqlTx.Commit(); err != nil {

		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
