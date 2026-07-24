package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

var _ usecase.ActivityRepository = (*activityRepository)(nil)

type activityRepository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewActivityRepository(pool *pgxpool.Pool) *activityRepository {
	return &activityRepository{pool: pool, q: db.New(pool)}
}

// Create implements [usecase.ActivityRepository].
func (r *activityRepository) Create(ctx context.Context, activity *entity.Activity) (*entity.Activity, error) {
	row, err := r.q.CreateActivity(ctx, toCreateActivityParams(activity))
	if err != nil {
		return nil, fmt.Errorf("insert activity: %w", err)
	}
	return toEntityActivity(row), nil
}

// WithTransaction implements [usecase.ActivityRepository].
func (r *activityRepository) WithTransaction(ctx context.Context, fn func(usecase.ActivityRepository) error) error {
	if r.pool == nil {
		return errors.New("activityRepository: cannot start a transaction from within an existing transaction")
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		// Rollback after a successful Commit always "fails" with
		// pgx.ErrTxClosed — that's expected and fine to discard. Any
		// other rollback error is unusual enough to be worth a trace,
		// but not worth escalating to Error: the real outcome (commit
		// succeeded, or the original err from fn) is already what gets
		// returned/logged through the normal path below.
		if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			slog.DebugContext(ctx, "transaction rollback failed", "error", rbErr)
		}
	}()

	txRepo := &activityRepository{q: db.New(tx)} // pool left nil on purpose, see guard above

	if err := fn(txRepo); err != nil {
		// Not wrapped further here — fn (the usecase's callback) already
		// wraps whatever Create returned. Wrapping again would just
		// stutter the error message ("creating activity: creating
		// activity: insert activity: ...").
		return err
	}
	return tx.Commit(ctx)

}
