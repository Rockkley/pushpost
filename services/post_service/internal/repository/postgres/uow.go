package postgres

import (
	"context"
	"database/sql"
	"fmt"

	outboxpg "github.com/rockkley/pushpost/services/common_service/outbox/postgres"
	"github.com/rockkley/pushpost/services/post_service/internal/domain"
	"github.com/rockkley/pushpost/services/post_service/internal/repository"
)

type uowTx struct {
	posts  repository.PostRepositoryInterface
	outbox domain.OutboxWriterInterface
}

func (t *uowTx) Posts() repository.PostRepositoryInterface { return t.posts }
func (t *uowTx) Outbox() domain.OutboxWriterInterface      { return t.outbox }

type UnitOfWork struct{ db *sql.DB }

func NewUnitOfWork(db *sql.DB) *UnitOfWork { return &UnitOfWork{db: db} }

func (u *UnitOfWork) Reader() repository.PostRepositoryInterface {
	return NewPostRepository(u.db)
}

func (u *UnitOfWork) Do(ctx context.Context, fn func(domain.Tx) error) error {
	tx, err := u.db.BeginTx(ctx, nil)

	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if err = fn(&uowTx{
		posts:  NewPostRepository(tx),
		outbox: outboxpg.NewWriterRepository(tx),
	}); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
