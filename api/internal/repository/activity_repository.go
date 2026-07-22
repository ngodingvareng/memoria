package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moez-rd/memoria/internal/db"
	"github.com/moez-rd/memoria/internal/entity"
	"github.com/moez-rd/memoria/internal/usecase"
)

var _ usecase.ActivityRepository = (*activityRepository)(nil)

type activityRepository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

// Create implements [usecase.ActivityRepository].

func NewActivityRepository(pool *pgxpool.Pool) *activityRepository {
	return &activityRepository{pool: pool, q: db.New(pool)}
}

func (r *activityRepository) Create(ctx context.Context, activity *entity.Activity) (*entity.Activity, error) {
	row, err := r.q.CreateActivity(ctx, toCreateActivityParams(activity))
	if err != nil {
		return nil, err
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
		return err
	}
	defer tx.Rollback(ctx)

	txRepo := &activityRepository{q: db.New(tx)}
	if err := fn(txRepo); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
