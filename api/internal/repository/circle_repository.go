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
	"github.com/ngodingvareng/memoria/internal/enum"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

var _ usecase.CircleRepository = (*circleRepository)(nil)
var _ usecase.CircleAccessChecker = (*circleRepository)(nil)
var _ usecase.CircleShareChecker = (*circleRepository)(nil)

type circleRepository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewCircleRepository(pool *pgxpool.Pool) *circleRepository {
	return &circleRepository{pool: pool, q: db.New(pool)}
}

// Create implements [usecase.CircleRepository].
func (r *circleRepository) Create(ctx context.Context, circle *entity.Circle) (*entity.Circle, error) {
	row, err := r.q.CreateCircle(ctx, toCreateCircleParams(circle))
	if err != nil {
		return nil, fmt.Errorf("insert circle: %w", err)
	}
	return toEntityCircle(row), nil
}

// AddCreatorAsAdmin implements [usecase.CircleRepository].
func (r *circleRepository) AddCreatorAsAdmin(
	ctx context.Context,
	circleID, userID uuid.UUID,
) (*entity.CircleMember, error) {
	row, err := r.q.AddCircleCreatorAsAdmin(ctx, db.AddCircleCreatorAsAdminParams{CircleID: circleID, UserID: userID})
	if err != nil {
		return nil, fmt.Errorf("seat circle creator as admin: %w", err)
	}
	return toEntityCircleMember(row), nil
}

// Update implements [usecase.CircleRepository].
func (r *circleRepository) Update(
	ctx context.Context,
	circle *entity.Circle,
	userID uuid.UUID,
) (*entity.Circle, error) {
	row, err := r.q.UpdateCircle(ctx, toUpdateCircleParams(circle, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("update circle: %w", err)
	}
	return toEntityCircle(row), nil
}

// UpdateImagePath implements [usecase.CircleRepository]. Same admin-only
// gate as Update, enforced in the WHERE clause — a non-admin's attempt
// and a wrong/foreign id are indistinguishable, both errs.ErrNotFound.
func (r *circleRepository) UpdateImagePath(
	ctx context.Context,
	id, userID uuid.UUID,
	imagePath *string,
) (*entity.Circle, error) {
	row, err := r.q.UpdateCircleImagePath(ctx, db.UpdateCircleImagePathParams{
		ID:        id,
		UserID:    userID,
		ImagePath: ptrToPgText(imagePath),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("update circle image path: %w", err)
	}
	return toEntityCircle(row), nil
}

// Dissolve implements [usecase.CircleRepository].
func (r *circleRepository) Dissolve(ctx context.Context, id, userID uuid.UUID) error {
	if err := r.q.DissolveCircle(ctx, db.DissolveCircleParams{ID: id, UserID: userID}); err != nil {
		return fmt.Errorf("dissolve circle: %w", err)
	}
	return nil
}

// GetByID implements [usecase.CircleRepository].
func (r *circleRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Circle, error) {
	row, err := r.q.GetCircleWithAccess(ctx, db.GetCircleWithAccessParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("get circle by id: %w", err)
	}
	return toEntityCircle(row), nil
}

// ListByUserID implements [usecase.CircleRepository].
func (r *circleRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Circle, error) {
	rows, err := r.q.ListCirclesByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list circles by user id: %w", err)
	}
	return toEntityCircles(rows), nil
}

// ListMembers implements [usecase.CircleRepository].
func (r *circleRepository) ListMembers(ctx context.Context, circleID uuid.UUID) ([]*entity.CircleMember, error) {
	rows, err := r.q.ListActiveCircleMembers(ctx, circleID)
	if err != nil {
		return nil, fmt.Errorf("list active circle members: %w", err)
	}
	return toEntityCircleMembers(rows), nil
}

// GetActiveMember implements [usecase.CircleRepository] and
// [usecase.CircleAccessChecker] — both interfaces declare this method
// with an identical name and signature (CODING_STANDARDS.md §5).
func (r *circleRepository) GetActiveMember(
	ctx context.Context,
	circleID, userID uuid.UUID,
) (*entity.CircleMember, error) {
	row, err := r.q.GetActiveCircleMember(ctx, db.GetActiveCircleMemberParams{CircleID: circleID, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("get active circle member: %w", err)
	}
	return toEntityCircleMember(row), nil
}

// IsSoleActiveAdmin implements [usecase.CircleRepository].
func (r *circleRepository) IsSoleActiveAdmin(ctx context.Context, circleID, userID uuid.UUID) (bool, error) {
	isSole, err := r.q.IsSoleActiveCircleAdmin(
		ctx,
		db.IsSoleActiveCircleAdminParams{CircleID: circleID, UserID: userID},
	)
	if err != nil {
		return false, fmt.Errorf("checking sole active circle admin: %w", err)
	}
	return isSole.Bool, nil
}

// Leave implements [usecase.CircleRepository]. LeaveCircle is a sqlc
// :exec query, so a WHERE clause matching zero rows is a silent no-op.
func (r *circleRepository) Leave(ctx context.Context, circleID, userID uuid.UUID) error {
	if err := r.q.LeaveCircle(ctx, db.LeaveCircleParams{CircleID: circleID, UserID: userID}); err != nil {
		return fmt.Errorf("leave circle: %w", err)
	}
	return nil
}

// RemoveMember implements [usecase.CircleRepository]. Silent no-op if
// removedByUserID isn't an active admin or userID isn't an active
// member — same reasoning as Leave.
func (r *circleRepository) RemoveMember(ctx context.Context, circleID, userID, removedByUserID uuid.UUID) error {
	if err := r.q.RemoveCircleMember(ctx, db.RemoveCircleMemberParams{
		CircleID: circleID, UserID: userID, RemovedByUserID: removedByUserID,
	}); err != nil {
		return fmt.Errorf("remove circle member: %w", err)
	}
	return nil
}

// UpdateMemberRole implements [usecase.CircleRepository].
func (r *circleRepository) UpdateMemberRole(
	ctx context.Context,
	circleID, userID, changedByUserID uuid.UUID,
	role enum.CircleRole,
) (*entity.CircleMember, error) {
	row, err := r.q.UpdateCircleMemberRole(ctx, db.UpdateCircleMemberRoleParams{
		Role: db.CircleRole(role), CircleID: circleID, UserID: userID, ChangedByUserID: changedByUserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("update circle member role: %w", err)
	}
	return toEntityCircleMember(row), nil
}

// UpdateMemberPermissions implements [usecase.CircleRepository].
func (r *circleRepository) UpdateMemberPermissions(
	ctx context.Context,
	circleID, userID, changedByUserID uuid.UUID,
	canInvite, canCapture bool,
) (*entity.CircleMember, error) {
	row, err := r.q.UpdateCircleMemberPermissions(ctx, db.UpdateCircleMemberPermissionsParams{
		CanInvite: canInvite, CanCapture: canCapture,
		CircleID: circleID, UserID: userID, ChangedByUserID: changedByUserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("update circle member permissions: %w", err)
	}
	return toEntityCircleMember(row), nil
}

// ListSharedCircleIDs implements [usecase.CircleShareChecker].
func (r *circleRepository) ListSharedCircleIDs(ctx context.Context, userA, userB uuid.UUID) ([]uuid.UUID, error) {
	ids, err := r.q.ListCircleIDsSharedBetweenUsers(
		ctx,
		db.ListCircleIDsSharedBetweenUsersParams{UserA: userA, UserB: userB},
	)
	if err != nil {
		return nil, fmt.Errorf("listing shared circles: %w", err)
	}
	return ids, nil
}

// WithTransaction implements [usecase.CircleRepository].
func (r *circleRepository) WithTransaction(ctx context.Context, fn func(usecase.CircleRepository) error) error {
	if r.pool == nil {
		return errors.New("circleRepository: cannot start a transaction from within an existing transaction")
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

	txRepo := &circleRepository{q: db.New(tx)} // pool left nil on purpose, see guard above

	if err := fn(txRepo); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
