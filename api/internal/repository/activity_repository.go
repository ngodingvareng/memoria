package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
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

// Update implements [usecase.ActivityRepository].
func (r *activityRepository) Update(ctx context.Context, activity *entity.Activity) (*entity.Activity, error) {
	row, err := r.q.UpdateActivity(ctx, toUpdateActivityParams(activity))
	if err != nil {
		// Same pattern as userRepository.GetByEmail: UpdateActivity's own
		// WHERE (id = ? AND user_id = ? AND deleted_at IS NULL) means
		// pgx.ErrNoRows covers "doesn't exist", "belongs to someone
		// else", and "already soft-deleted" all at once — from here
		// they're indistinguishable, and indistinguishable is exactly
		// what we want to expose as 404 rather than leaking which case
		// it was.
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("update activity: %w", err)
	}
	return toEntityActivity(row), nil
}

// SoftDelete implements [usecase.ActivityRepository]. Soft delete only — sets
// deleted_at rather than removing the row, same recovery-grace-period
// reasoning as SoftDeleteUser (see SCHEMA_REVIEW.md #9 for why a hard
// delete here would also orphan any activity_images/activity_captures
// still on disk).
//
// SoftDeleteActivity is a sqlc :exec query (no rows-affected count), so
// — same as ActivityImageRepository.Delete, see its test
// TestActivityImageRepository_Delete_WrongActivityID_NoOp — a WHERE
// clause matching zero rows is a silent no-op here, not errs.ErrNotFound.
// If callers need a real 404 on a missing/foreign activity, look it up
// via GetActivityByID first.
func (r *activityRepository) SoftDelete(ctx context.Context, id, userID uuid.UUID) error {
	if err := r.q.SoftDeleteActivity(ctx, db.SoftDeleteActivityParams{
		ID:     id,
		UserID: userID,
	}); err != nil {
		return fmt.Errorf("soft delete activity: %w", err)
	}
	return nil
}

// GetByID implements [usecase.ActivityRepository].
func (r *activityRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Activity, error) {
	row, err := r.q.GetActivityByID(ctx, db.GetActivityByIDParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("get activity by id: %w", err)
	}
	return toEntityActivity(row), nil
}

// Search implements [usecase.ActivityRepository].
func (r *activityRepository) Search(ctx context.Context, params usecase.SearchActivitiesParams) ([]*entity.Activity, int64, error) {
	rows, err := r.q.SearchActivities(ctx, db.SearchActivitiesParams{
		UserID:          params.UserID,
		Name:            ptrToPgText(params.Name),
		IsFixedSchedule: ptrToPgBool(params.IsFixedSchedule),
		PageOffset:      params.Offset,
		PageLimit:       params.Limit,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("search activities: %w", err)
	}

	total, err := r.q.CountSearchActivities(ctx, db.CountSearchActivitiesParams{
		UserID:          params.UserID,
		Name:            ptrToPgText(params.Name),
		IsFixedSchedule: ptrToPgBool(params.IsFixedSchedule),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("count search activities: %w", err)
	}

	activities := make([]*entity.Activity, len(rows))
	for i, row := range rows {
		activities[i] = toEntityActivity(row)
	}
	return activities, total, nil
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

func (r *activityRepository) GetActivityByID(ctx context.Context, id, userID uuid.UUID) (*entity.Activity, error) {
	row, err := r.q.GetActivityByID(ctx, db.GetActivityByIDParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("get activity by id: %w", err)
	}
	return toEntityActivity(row), nil
}
