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

var _ usecase.ThreadRepository = (*threadRepository)(nil)
var _ usecase.ThreadAccessChecker = (*threadRepository)(nil)

type threadRepository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewThreadRepository(pool *pgxpool.Pool) *threadRepository {
	return &threadRepository{pool: pool, q: db.New(pool)}
}

// Create implements [usecase.ThreadRepository].
func (r *threadRepository) Create(ctx context.Context, thread *entity.Thread) (*entity.Thread, error) {
	row, err := r.q.CreateThread(ctx, toCreateThreadParams(thread))
	if err != nil {
		return nil, fmt.Errorf("insert thread: %w", err)
	}
	return toEntityThread(row), nil
}

// Update implements [usecase.ThreadRepository].
func (r *threadRepository) Update(ctx context.Context, thread *entity.Thread) (*entity.Thread, error) {
	row, err := r.q.UpdateThread(ctx, toUpdateThreadParams(thread))
	if err != nil {
		// Same pattern as userRepository.GetByEmail: UpdateThread's own
		// WHERE (id = ? AND user_id = ? AND deleted_at IS NULL) means
		// pgx.ErrNoRows covers "doesn't exist", "belongs to someone
		// else", and "already soft-deleted" all at once — from here
		// they're indistinguishable, and indistinguishable is exactly
		// what we want to expose as 404 rather than leaking which case
		// it was.
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("update thread: %w", err)
	}
	return toEntityThread(row), nil
}

// SoftDelete implements [usecase.ThreadRepository]. Soft delete only — sets
// deleted_at rather than removing the row, same recovery-grace-period
// reasoning as SoftDeleteUser (see SCHEMA_REVIEW.md #9 for why a hard
// delete here would also orphan any thread_images/moments
// still on disk).
//
// SoftDeleteThread is a sqlc :exec query (no rows-affected count), so
// — same as ThreadImageRepository.Delete, see its test
// TestThreadImageRepository_Delete_WrongThreadID_NoOp — a WHERE
// clause matching zero rows is a silent no-op here, not errs.ErrNotFound.
// If callers need a real 404 on a missing/foreign thread, look it up
// via GetByID first.
func (r *threadRepository) SoftDelete(ctx context.Context, id, userID uuid.UUID) error {
	if err := r.q.SoftDeleteThread(ctx, db.SoftDeleteThreadParams{
		ID:     id,
		UserID: userID,
	}); err != nil {
		return fmt.Errorf("soft delete thread: %w", err)
	}
	return nil
}

// GetByID implements [usecase.ThreadRepository] and [usecase.ThreadAccessChecker] —
// both interfaces declare this method with an identical name and signature,
// so this single implementation satisfies both without a second method.
//
// Backed by GetThreadWithAccess rather than the plain GetThreadByID
// query: a Thread can now be collaborative (circle_id set), and access
// to those is governed by circle_members, not user_id alone — see the
// header comment on the threads table. For a personal Thread (the only
// kind this app currently creates) the two are equivalent.
func (r *threadRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Thread, error) {
	row, err := r.q.GetThreadWithAccess(ctx, db.GetThreadWithAccessParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("get thread by id: %w", err)
	}
	return toEntityThread(row), nil
}

// Search implements [usecase.ThreadRepository].
func (r *threadRepository) Search(ctx context.Context, params usecase.SearchThreadsParams) ([]*entity.Thread, int64, error) {
	rows, err := r.q.SearchThreads(ctx, db.SearchThreadsParams{
		UserID:        params.UserID,
		Name:          ptrToPgText(params.Name),
		Archived:      ptrToPgBool(params.Archived),
		SortByRecency: params.SortByRecency,
		PageOffset:    params.Offset,
		PageLimit:     params.Limit,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("search threads: %w", err)
	}

	total, err := r.q.CountSearchThreads(ctx, db.CountSearchThreadsParams{
		UserID:   params.UserID,
		Name:     ptrToPgText(params.Name),
		Archived: ptrToPgBool(params.Archived),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("count search threads: %w", err)
	}

	threads := make([]*entity.Thread, len(rows))
	for i, row := range rows {
		threads[i] = toEntityThread(row)
	}
	return threads, total, nil
}

// ListByCircle implements [usecase.ThreadRepository].
func (r *threadRepository) ListByCircle(ctx context.Context, circleID uuid.UUID) ([]*entity.Thread, error) {
	rows, err := r.q.ListThreadsByCircleID(ctx, ptrToPgUUID(&circleID))
	if err != nil {
		return nil, fmt.Errorf("list threads by circle: %w", err)
	}

	threads := make([]*entity.Thread, len(rows))
	for i, row := range rows {
		threads[i] = toEntityThread(row)
	}
	return threads, nil
}

// WithTransaction implements [usecase.ThreadRepository].
func (r *threadRepository) WithTransaction(ctx context.Context, fn func(usecase.ThreadRepository) error) error {
	if r.pool == nil {
		return errors.New("threadRepository: cannot start a transaction from within an existing transaction")
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

	txRepo := &threadRepository{q: db.New(tx)} // pool left nil on purpose, see guard above

	if err := fn(txRepo); err != nil {
		// Not wrapped further here — fn (the usecase's callback) already
		// wraps whatever Create returned. Wrapping again would just
		// stutter the error message ("creating thread: creating
		// thread: insert thread: ...").
		return err
	}
	return tx.Commit(ctx)

}
