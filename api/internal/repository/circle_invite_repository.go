package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

var _ usecase.CircleInviteRepository = (*circleInviteRepository)(nil)
var _ usecase.CircleInviteLinkResolver = (*circleInviteRepository)(nil)

type circleInviteRepository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewCircleInviteRepository(pool *pgxpool.Pool) *circleInviteRepository {
	return &circleInviteRepository{pool: pool, q: db.New(pool)}
}

// CreateUsernameInvite implements [usecase.CircleInviteRepository].
func (r *circleInviteRepository) CreateUsernameInvite(
	ctx context.Context,
	circleID, invitedByUserID, inviteeUserID uuid.UUID,
	expiresAt time.Time,
) (*entity.CircleInvite, error) {
	row, err := r.q.CreateCircleUsernameInvite(ctx, db.CreateCircleUsernameInviteParams{
		CircleID: circleID, InvitedByUserID: ptrToPgUUID(&invitedByUserID),
		InviteeUserID: ptrToPgUUID(&inviteeUserID), ExpiresAt: timeToPgTimestamptz(expiresAt),
	})
	if err != nil {
		return nil, fmt.Errorf("create circle username invite: %w", err)
	}
	return toEntityCircleInvite(row), nil
}

// GetPendingUsernameInvite implements [usecase.CircleInviteRepository].
func (r *circleInviteRepository) GetPendingUsernameInvite(
	ctx context.Context,
	circleID, inviteeUserID uuid.UUID,
) (*entity.CircleInvite, error) {
	row, err := r.q.GetPendingCircleUsernameInvite(ctx, db.GetPendingCircleUsernameInviteParams{
		CircleID: circleID, InviteeUserID: ptrToPgUUID(&inviteeUserID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("get pending circle username invite: %w", err)
	}
	return toEntityCircleInvite(row), nil
}

// ListPendingByInviteeID implements [usecase.CircleInviteRepository].
func (r *circleInviteRepository) ListPendingByInviteeID(
	ctx context.Context,
	inviteeUserID uuid.UUID,
) ([]*entity.CircleInvite, error) {
	rows, err := r.q.ListPendingCircleInvitesByInviteeID(ctx, ptrToPgUUID(&inviteeUserID))
	if err != nil {
		return nil, fmt.Errorf("list pending circle invites: %w", err)
	}
	return toEntityCircleInvites(rows), nil
}

// AcceptUsernameInvite implements [usecase.CircleInviteRepository].
func (r *circleInviteRepository) AcceptUsernameInvite(
	ctx context.Context,
	id, inviteeUserID uuid.UUID,
) (*entity.CircleInvite, error) {
	row, err := r.q.AcceptCircleUsernameInvite(
		ctx,
		db.AcceptCircleUsernameInviteParams{ID: id, InviteeUserID: ptrToPgUUID(&inviteeUserID)},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("accept circle username invite: %w", err)
	}
	return toEntityCircleInvite(row), nil
}

// DeclineUsernameInvite implements [usecase.CircleInviteRepository].
func (r *circleInviteRepository) DeclineUsernameInvite(
	ctx context.Context,
	id, inviteeUserID uuid.UUID,
) (*entity.CircleInvite, error) {
	row, err := r.q.DeclineCircleUsernameInvite(
		ctx,
		db.DeclineCircleUsernameInviteParams{ID: id, InviteeUserID: ptrToPgUUID(&inviteeUserID)},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("decline circle username invite: %w", err)
	}
	return toEntityCircleInvite(row), nil
}

// RevokeUsernameInvite implements [usecase.CircleInviteRepository].
func (r *circleInviteRepository) RevokeUsernameInvite(
	ctx context.Context,
	id, circleID uuid.UUID,
) (*entity.CircleInvite, error) {
	row, err := r.q.RevokeCircleUsernameInvite(ctx, db.RevokeCircleUsernameInviteParams{ID: id, CircleID: circleID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("revoke circle username invite: %w", err)
	}
	return toEntityCircleInvite(row), nil
}

// RevokeActiveLink implements [usecase.CircleInviteRepository]. Returns
// nil, nil when the Circle has no live link yet — the ordinary
// first-generation case, not an error (see the query's own comment).
func (r *circleInviteRepository) RevokeActiveLink(
	ctx context.Context,
	circleID uuid.UUID,
) (*entity.CircleInvite, error) {
	row, err := r.q.RevokeActiveCircleInviteLink(ctx, circleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("revoke active circle invite link: %w", err)
	}
	return toEntityCircleInvite(row), nil
}

// CreateLink implements [usecase.CircleInviteRepository].
func (r *circleInviteRepository) CreateLink(
	ctx context.Context,
	circleID, invitedByUserID uuid.UUID,
	tokenHash string,
	requiresApproval bool,
) (*entity.CircleInvite, error) {
	row, err := r.q.CreateCircleInviteLink(ctx, db.CreateCircleInviteLinkParams{
		CircleID: circleID, InvitedByUserID: ptrToPgUUID(&invitedByUserID),
		TokenHash: ptrToPgText(&tokenHash), RequiresApproval: ptrToPgBool(&requiresApproval),
	})
	if err != nil {
		return nil, fmt.Errorf("create circle invite link: %w", err)
	}
	return toEntityCircleInvite(row), nil
}

// GetActiveLinkByCircleID implements [usecase.CircleInviteRepository].
func (r *circleInviteRepository) GetActiveLinkByCircleID(
	ctx context.Context,
	circleID uuid.UUID,
) (*entity.CircleInvite, error) {
	row, err := r.q.GetActiveCircleInviteLinkByCircleID(ctx, circleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("get active circle invite link: %w", err)
	}
	return toEntityCircleInvite(row), nil
}

// GetActiveLinkByTokenHash implements [usecase.CircleInviteRepository]
// and [usecase.CircleInviteLinkResolver].
func (r *circleInviteRepository) GetActiveLinkByTokenHash(
	ctx context.Context,
	tokenHash string,
) (*entity.CircleInvite, error) {
	row, err := r.q.GetActiveCircleInviteLinkByTokenHash(ctx, ptrToPgText(&tokenHash))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("get circle invite link by token: %w", err)
	}
	return toEntityCircleInvite(row), nil
}

// SetLinkRequiresApproval implements [usecase.CircleInviteRepository].
func (r *circleInviteRepository) SetLinkRequiresApproval(
	ctx context.Context,
	circleID uuid.UUID,
	requiresApproval bool,
) (*entity.CircleInvite, error) {
	row, err := r.q.SetCircleInviteLinkRequiresApproval(ctx, db.SetCircleInviteLinkRequiresApprovalParams{
		CircleID: circleID, RequiresApproval: ptrToPgBool(&requiresApproval),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("set circle invite link approval requirement: %w", err)
	}
	return toEntityCircleInvite(row), nil
}

// AddMembers implements [usecase.CircleInviteRepository].
func (r *circleInviteRepository) AddMembers(
	ctx context.Context,
	circleID uuid.UUID,
	userIDs []uuid.UUID,
) ([]*entity.CircleMember, error) {
	rows, err := r.q.AddCircleMembers(ctx, db.AddCircleMembersParams{CircleID: circleID, UserIds: userIDs})
	if err != nil {
		return nil, fmt.Errorf("add circle members: %w", err)
	}
	return toEntityCircleMembers(rows), nil
}

// WithTransaction implements [usecase.CircleInviteRepository].
func (r *circleInviteRepository) WithTransaction(
	ctx context.Context,
	fn func(usecase.CircleInviteRepository) error,
) error {
	if r.pool == nil {
		return errors.New("circleInviteRepository: cannot start a transaction from within an existing transaction")
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

	txRepo := &circleInviteRepository{q: db.New(tx)} // pool left nil on purpose, see guard above

	if err := fn(txRepo); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
