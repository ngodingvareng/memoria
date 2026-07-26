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

var _ usecase.AuthUnitOfWork = (*authUnitOfWork)(nil)

// authUnitOfWork coordinates a transaction across MORE than one
// repository — the multi-repository counterpart to activityRepository's
// own single-repository WithTransaction. Register needs to insert both a
// users row and a user_accounts row atomically; neither repository alone
// can guarantee that, so this exists specifically to own that
// transaction boundary.
type authUnitOfWork struct {
	pool *pgxpool.Pool
}

func NewAuthUnitOfWork(pool *pgxpool.Pool) *authUnitOfWork {
	return &authUnitOfWork{pool: pool}
}

// WithTransaction implements [usecase.AuthUnitOfWork].
func (u *authUnitOfWork) WithTransaction(ctx context.Context, fn func(usecase.AuthRepositories) error) error {
	tx, err := u.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			slog.DebugContext(ctx, "auth transaction rollback failed", "error", rbErr)
		}
	}()

	q := db.New(tx)
	repos := usecase.AuthRepositories{
		User:        &userRepository{q: q},
		UserAccount: &userAccountRepository{q: q},
	}

	if err := fn(repos); err != nil {
		return err
	}
	return tx.Commit(ctx)

}
