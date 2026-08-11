package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

var _ usecase.AccountDeletionUnitOfWork = (*accountDeletionUnitOfWork)(nil)

// accountDeletionUnitOfWork coordinates a transaction across every
// repository UserUsecase.DeleteAccount needs — same shape as
// authUnitOfWork, just a larger repository set. Each sub-repository is
// constructed with pool left nil: none of them need to start their own
// nested transaction from inside this one (see e.g. circleRepository's
// own WithTransaction guard against exactly that).
type accountDeletionUnitOfWork struct {
	pool *pgxpool.Pool
}

func NewAccountDeletionUnitOfWork(pool *pgxpool.Pool) *accountDeletionUnitOfWork {
	return &accountDeletionUnitOfWork{pool: pool}
}

// WithTransaction implements [usecase.AccountDeletionUnitOfWork].
func (u *accountDeletionUnitOfWork) WithTransaction(ctx context.Context, fn func(usecase.AccountDeletionRepositories) error) error {
	tx, err := u.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			slog.DebugContext(ctx, "account deletion transaction rollback failed", "error", rbErr)
		}
	}()

	q := db.New(tx)
	repos := usecase.AccountDeletionRepositories{
		User:         &userRepository{q: q},
		RefreshToken: &refreshTokenRepository{q: q},
		Moment:       &momentRepository{q: q},
		MomentImage:  &momentImageRepository{q: q},
		Thread:       &threadRepository{q: q},
		ThreadImage:  &threadImageRepository{q: q},
		Comment:      &commentRepository{q: q},
		Reaction:     &reactionRepository{q: q},
		Circle:       &circleRepository{q: q},
	}

	if err := fn(repos); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
