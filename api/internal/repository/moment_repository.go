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

var _ usecase.MomentRepository = (*momentRepository)(nil)
var _ usecase.MomentAccessChecker = (*momentRepository)(nil)

type momentRepository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewMomentRepository(pool *pgxpool.Pool) *momentRepository {
	return &momentRepository{pool: pool, q: db.New(pool)}
}

// Create implements [usecase.MomentRepository]. When moment.ClientID
// collides with an already-synced offline capture, CreateMoment's own
// ON CONFLICT ... DO NOTHING skips the insert and RETURNING yields no
// row (pgx.ErrNoRows here) — that's re-fetched and returned instead of
// surfacing an error, so a flaky retry gets back the same Moment it
// already recorded rather than a failure.
func (r *momentRepository) Create(ctx context.Context, moment *entity.Moment) (*entity.Moment, error) {
	row, err := r.q.CreateMoment(ctx, toCreateMomentParams(moment))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) && moment.ClientID != nil {
			existing, getErr := r.q.GetMomentByUserAndClientID(ctx, db.GetMomentByUserAndClientIDParams{
				UserID:   moment.UserID,
				ClientID: ptrToPgUUID(moment.ClientID),
			})
			if getErr != nil {
				return nil, fmt.Errorf("re-fetching moment after conflicting insert: %w", getErr)
			}
			return toEntityMoment(existing), nil
		}
		return nil, fmt.Errorf("insert moment: %w", err)
	}
	return toEntityMoment(row), nil
}

// Update implements [usecase.MomentRepository].
func (r *momentRepository) Update(ctx context.Context, moment *entity.Moment) (*entity.Moment, error) {
	row, err := r.q.UpdateMoment(ctx, toUpdateMomentParams(moment))
	if err != nil {
		// Same reasoning as threadRepository.Update: the query's own
		// WHERE (id = ? AND user_id = ? AND deleted_at IS NULL) makes
		// "doesn't exist", "belongs to someone else", and "already
		// soft-deleted" indistinguishable from here, which is exactly
		// what we want to expose as 404.
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("update moment: %w", err)
	}
	return toEntityMoment(row), nil
}

// SoftDelete implements [usecase.MomentRepository]. SoftDeleteMoment is a
// sqlc :exec query (no rows-affected count), so — same as
// ThreadRepository.SoftDelete — a WHERE clause matching zero rows (wrong
// owner/id, or already deleted) is a silent no-op here, not
// errs.ErrNotFound. Callers needing a real 404 should look the Moment up
// via GetByID first.
func (r *momentRepository) SoftDelete(ctx context.Context, id, userID uuid.UUID) error {
	if err := r.q.SoftDeleteMoment(ctx, db.SoftDeleteMomentParams{ID: id, UserID: userID}); err != nil {
		return fmt.Errorf("soft delete moment: %w", err)
	}
	return nil
}

// GetByID implements [usecase.MomentRepository] and
// [usecase.MomentAccessChecker] — both interfaces declare this method
// with an identical name and signature, so this single implementation
// satisfies both (see CODING_STANDARDS.md §5).
func (r *momentRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Moment, error) {
	row, err := r.q.GetMomentByID(ctx, db.GetMomentByIDParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("get moment by id: %w", err)
	}
	return toEntityMoment(row), nil
}

// ListByUserID implements [usecase.MomentRepository].
func (r *momentRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*entity.Moment, error) {
	rows, err := r.q.ListMomentsByUserID(ctx, db.ListMomentsByUserIDParams{UserID: userID, PageLimit: limit, PageOffset: offset})
	if err != nil {
		return nil, fmt.Errorf("list moments by user id: %w", err)
	}
	return toEntityMoments(rows), nil
}

// ListByThreadID implements [usecase.MomentRepository]. No user_id
// filter at the query level — access to the Thread itself is checked
// upstream by the usecase via ThreadAccessChecker.
func (r *momentRepository) ListByThreadID(ctx context.Context, threadID uuid.UUID, limit, offset int32) ([]*entity.Moment, error) {
	rows, err := r.q.ListMomentsByThreadID(ctx, db.ListMomentsByThreadIDParams{
		ThreadID:   ptrToPgUUID(&threadID),
		PageLimit:  limit,
		PageOffset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list moments by thread id: %w", err)
	}
	return toEntityMoments(rows), nil
}

// Search implements [usecase.MomentRepository].
func (r *momentRepository) Search(ctx context.Context, userID uuid.UUID, query string, limit, offset int32) ([]*entity.Moment, error) {
	rows, err := r.q.SearchMoments(ctx, db.SearchMomentsParams{
		UserID:     userID,
		Query:      query,
		PageLimit:  limit,
		PageOffset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("search moments: %w", err)
	}
	return toEntityMoments(rows), nil
}

// WithTransaction implements [usecase.MomentRepository].
func (r *momentRepository) WithTransaction(ctx context.Context, fn func(usecase.MomentRepository) error) error {
	if r.pool == nil {
		return errors.New("momentRepository: cannot start a transaction from within an existing transaction")
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			slog.DebugContext(ctx, "transaction rollback failed", "error", rbErr)
		}
	}()

	txRepo := &momentRepository{q: db.New(tx)} // pool left nil on purpose, see guard above

	if err := fn(txRepo); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
