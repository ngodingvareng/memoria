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

var _ usecase.CircleJoinRequestRepository = (*circleJoinRequestRepository)(nil)

type circleJoinRequestRepository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewCircleJoinRequestRepository(pool *pgxpool.Pool) *circleJoinRequestRepository {
	return &circleJoinRequestRepository{pool: pool, q: db.New(pool)}
}

// Create implements [usecase.CircleJoinRequestRepository].
func (r *circleJoinRequestRepository) Create(ctx context.Context, circleID, userID uuid.UUID, inviteID *uuid.UUID) (*entity.CircleJoinRequest, error) {
	row, err := r.q.CreateCircleJoinRequest(ctx, db.CreateCircleJoinRequestParams{
		CircleID: circleID, UserID: userID, InviteID: ptrToPgUUID(inviteID),
	})
	if err != nil {
		return nil, fmt.Errorf("create circle join request: %w", err)
	}
	return toEntityCircleJoinRequest(row), nil
}

// GetPendingByUser implements [usecase.CircleJoinRequestRepository].
func (r *circleJoinRequestRepository) GetPendingByUser(ctx context.Context, circleID, userID uuid.UUID) (*entity.CircleJoinRequest, error) {
	row, err := r.q.GetPendingCircleJoinRequestByUser(ctx, db.GetPendingCircleJoinRequestByUserParams{CircleID: circleID, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("get pending circle join request: %w", err)
	}
	return toEntityCircleJoinRequest(row), nil
}

// ListPendingByCircleID implements [usecase.CircleJoinRequestRepository].
func (r *circleJoinRequestRepository) ListPendingByCircleID(ctx context.Context, circleID uuid.UUID, limit, offset int32) ([]*entity.CircleJoinRequest, error) {
	rows, err := r.q.ListPendingCircleJoinRequestsByCircleID(ctx, db.ListPendingCircleJoinRequestsByCircleIDParams{
		CircleID: circleID, PageLimit: limit, PageOffset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list pending circle join requests: %w", err)
	}
	return toEntityCircleJoinRequests(rows), nil
}

// CountPendingByCircleID implements [usecase.CircleJoinRequestRepository].
func (r *circleJoinRequestRepository) CountPendingByCircleID(ctx context.Context, circleID uuid.UUID) (int64, error) {
	count, err := r.q.CountPendingCircleJoinRequestsByCircleID(ctx, circleID)
	if err != nil {
		return 0, fmt.Errorf("count pending circle join requests: %w", err)
	}
	return count, nil
}

// Approve implements [usecase.CircleJoinRequestRepository]. Returns
// errs.ErrNotFound if someone else already decided this request —
// exactly how two approvers acting at once resolve to one decision.
func (r *circleJoinRequestRepository) Approve(ctx context.Context, id, circleID, decidedByUserID uuid.UUID) (*entity.CircleJoinRequest, error) {
	row, err := r.q.ApproveCircleJoinRequest(ctx, db.ApproveCircleJoinRequestParams{
		ID: id, CircleID: circleID, DecidedByUserID: ptrToPgUUID(&decidedByUserID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("approve circle join request: %w", err)
	}
	return toEntityCircleJoinRequest(row), nil
}

// Reject implements [usecase.CircleJoinRequestRepository].
func (r *circleJoinRequestRepository) Reject(ctx context.Context, id, circleID, decidedByUserID uuid.UUID) (*entity.CircleJoinRequest, error) {
	row, err := r.q.RejectCircleJoinRequest(ctx, db.RejectCircleJoinRequestParams{
		ID: id, CircleID: circleID, DecidedByUserID: ptrToPgUUID(&decidedByUserID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("reject circle join request: %w", err)
	}
	return toEntityCircleJoinRequest(row), nil
}

// Cancel implements [usecase.CircleJoinRequestRepository].
func (r *circleJoinRequestRepository) Cancel(ctx context.Context, id, userID uuid.UUID) (*entity.CircleJoinRequest, error) {
	row, err := r.q.CancelCircleJoinRequest(ctx, db.CancelCircleJoinRequestParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("cancel circle join request: %w", err)
	}
	return toEntityCircleJoinRequest(row), nil
}

// AddMembers implements [usecase.CircleJoinRequestRepository]. Same
// underlying write as CircleInviteRepository.AddMembers — both flows
// funnel into the one admit path (FEATURES.md, Circle Invite).
func (r *circleJoinRequestRepository) AddMembers(ctx context.Context, circleID uuid.UUID, userIDs []uuid.UUID) ([]*entity.CircleMember, error) {
	rows, err := r.q.AddCircleMembers(ctx, db.AddCircleMembersParams{CircleID: circleID, UserIds: userIDs})
	if err != nil {
		return nil, fmt.Errorf("add circle members: %w", err)
	}
	return toEntityCircleMembers(rows), nil
}

// WithTransaction implements [usecase.CircleJoinRequestRepository].
func (r *circleJoinRequestRepository) WithTransaction(ctx context.Context, fn func(usecase.CircleJoinRequestRepository) error) error {
	if r.pool == nil {
		return errors.New("circleJoinRequestRepository: cannot start a transaction from within an existing transaction")
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

	txRepo := &circleJoinRequestRepository{q: db.New(tx)} // pool left nil on purpose, see guard above

	if err := fn(txRepo); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
