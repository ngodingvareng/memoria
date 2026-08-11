package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

var _ usecase.UserBlockChecker = (*userPrivacyRepository)(nil)
var _ usecase.UserBlockRepository = (*userPrivacyRepository)(nil)
var _ usecase.UserKnownChecker = (*userPrivacyRepository)(nil)
var _ usecase.UserKnownRepository = (*userPrivacyRepository)(nil)
var _ usecase.UserMuteChecker = (*userPrivacyRepository)(nil)
var _ usecase.UserMuteRepository = (*userPrivacyRepository)(nil)

// userPrivacyRepository wraps user_blocks.sql/user_knows.sql/
// user_mutes.sql: the read-only Is*/gating checks, and the Block/Mute/
// Known mark + unmark + list CRUD (see user_privacy.go).
type userPrivacyRepository struct {
	q *db.Queries
}

func NewUserPrivacyRepository(pool *pgxpool.Pool) *userPrivacyRepository {
	return &userPrivacyRepository{q: db.New(pool)}
}

// IsBlockedEitherDirection implements [usecase.UserBlockChecker].
func (r *userPrivacyRepository) IsBlockedEitherDirection(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	blocked, err := r.q.IsBlockedEitherDirection(ctx, db.IsBlockedEitherDirectionParams{UserA: userA, UserB: userB})
	if err != nil {
		return false, fmt.Errorf("checking block status: %w", err)
	}
	return blocked, nil
}

// IsKnownTo implements [usecase.UserKnownChecker].
func (r *userPrivacyRepository) IsKnownTo(ctx context.Context, ownerUserID, otherUserID uuid.UUID) (bool, error) {
	known, err := r.q.IsUserKnownTo(ctx, db.IsUserKnownToParams{OwnerUserID: ownerUserID, OtherUserID: otherUserID})
	if err != nil {
		return false, fmt.Errorf("checking known status: %w", err)
	}
	return known.Bool, nil
}

// MarkKnown implements [usecase.UserKnownRepository]. Idempotent (see
// user_knows.sql's CreateUserKnown) — marking someone already known is
// a no-op, not an error.
func (r *userPrivacyRepository) MarkKnown(ctx context.Context, knowerUserID, knownUserID uuid.UUID) error {
	if err := r.q.CreateUserKnown(ctx, db.CreateUserKnownParams{
		KnowerUserID: knowerUserID,
		KnownUserID:  knownUserID,
	}); err != nil {
		return fmt.Errorf("marking user known: %w", err)
	}
	return nil
}

// Unmark implements [usecase.UserKnownRepository]. Silent no-op if the
// pair isn't currently marked known.
func (r *userPrivacyRepository) Unmark(ctx context.Context, knowerUserID, knownUserID uuid.UUID) error {
	if err := r.q.DeleteUserKnown(ctx, db.DeleteUserKnownParams{
		KnowerUserID: knowerUserID,
		KnownUserID:  knownUserID,
	}); err != nil {
		return fmt.Errorf("unmarking user known: %w", err)
	}
	return nil
}

// ListKnownUserIDs implements [usecase.UserKnownRepository].
func (r *userPrivacyRepository) ListKnownUserIDs(ctx context.Context, knowerUserID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.q.ListKnownUsersByKnowerID(ctx, knowerUserID)
	if err != nil {
		return nil, fmt.Errorf("listing known users: %w", err)
	}
	ids := make([]uuid.UUID, len(rows))
	for i, row := range rows {
		ids[i] = row.KnownUserID
	}
	return ids, nil
}

// IsMuted implements [usecase.UserMuteChecker].
func (r *userPrivacyRepository) IsMuted(ctx context.Context, muterUserID, mutedUserID uuid.UUID) (bool, error) {
	muted, err := r.q.IsUserMuted(ctx, db.IsUserMutedParams{MuterUserID: muterUserID, MutedUserID: mutedUserID})
	if err != nil {
		return false, fmt.Errorf("checking mute status: %w", err)
	}
	return muted, nil
}

// Block implements [usecase.UserBlockRepository]. Idempotent (see
// user_blocks.sql's CreateUserBlock) — blocking someone already blocked
// is a no-op, not an error.
func (r *userPrivacyRepository) Block(ctx context.Context, blockerUserID, blockedUserID uuid.UUID) error {
	if err := r.q.CreateUserBlock(ctx, db.CreateUserBlockParams{BlockerUserID: blockerUserID, BlockedUserID: blockedUserID}); err != nil {
		return fmt.Errorf("blocking user: %w", err)
	}
	return nil
}

// Unblock implements [usecase.UserBlockRepository]. Silent no-op if the
// pair isn't currently blocked.
func (r *userPrivacyRepository) Unblock(ctx context.Context, blockerUserID, blockedUserID uuid.UUID) error {
	if err := r.q.DeleteUserBlock(ctx, db.DeleteUserBlockParams{BlockerUserID: blockerUserID, BlockedUserID: blockedUserID}); err != nil {
		return fmt.Errorf("unblocking user: %w", err)
	}
	return nil
}

// ListBlockedUserIDs implements [usecase.UserBlockRepository].
func (r *userPrivacyRepository) ListBlockedUserIDs(ctx context.Context, blockerUserID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.q.ListBlockedUsersByBlockerID(ctx, blockerUserID)
	if err != nil {
		return nil, fmt.Errorf("listing blocked users: %w", err)
	}
	ids := make([]uuid.UUID, len(rows))
	for i, row := range rows {
		ids[i] = row.BlockedUserID
	}
	return ids, nil
}

// Mute implements [usecase.UserMuteRepository]. Idempotent (see
// user_mutes.sql's CreateUserMute) — muting someone already muted is a
// no-op, not an error.
func (r *userPrivacyRepository) Mute(ctx context.Context, muterUserID, mutedUserID uuid.UUID) error {
	if err := r.q.CreateUserMute(ctx, db.CreateUserMuteParams{MuterUserID: muterUserID, MutedUserID: mutedUserID}); err != nil {
		return fmt.Errorf("muting user: %w", err)
	}
	return nil
}

// Unmute implements [usecase.UserMuteRepository]. Silent no-op if the
// pair isn't currently muted.
func (r *userPrivacyRepository) Unmute(ctx context.Context, muterUserID, mutedUserID uuid.UUID) error {
	if err := r.q.DeleteUserMute(ctx, db.DeleteUserMuteParams{MuterUserID: muterUserID, MutedUserID: mutedUserID}); err != nil {
		return fmt.Errorf("unmuting user: %w", err)
	}
	return nil
}

// ListMutedUserIDs implements [usecase.UserMuteRepository].
func (r *userPrivacyRepository) ListMutedUserIDs(ctx context.Context, muterUserID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.q.ListMutedUsersByMuterID(ctx, muterUserID)
	if err != nil {
		return nil, fmt.Errorf("listing muted users: %w", err)
	}
	ids := make([]uuid.UUID, len(rows))
	for i, row := range rows {
		ids[i] = row.MutedUserID
	}
	return ids, nil
}
