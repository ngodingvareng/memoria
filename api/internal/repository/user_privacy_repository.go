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
var _ usecase.UserKnownChecker = (*userPrivacyRepository)(nil)
var _ usecase.UserKnownRepository = (*userPrivacyRepository)(nil)
var _ usecase.UserMuteChecker = (*userPrivacyRepository)(nil)

// userPrivacyRepository wraps the read-only Is* checks from
// user_blocks.sql/user_knows.sql/user_mutes.sql, plus the one write
// path built so far — MarkKnown. It still doesn't expose Delete/List
// for any of the three tables — the rest of Known/Block/Mute management
// is a separate, not-yet-built feature (see user_privacy.go).
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

// MarkKnown implements [usecase.UserKnownRepository].
func (r *userPrivacyRepository) MarkKnown(ctx context.Context, knowerUserID, knownUserID uuid.UUID) error {
	if err := r.q.CreateUserKnown(ctx, db.CreateUserKnownParams{
		KnowerUserID: knowerUserID,
		KnownUserID:  knownUserID,
	}); err != nil {
		return fmt.Errorf("marking user known: %w", err)
	}
	return nil
}

// IsMuted implements [usecase.UserMuteChecker].
func (r *userPrivacyRepository) IsMuted(ctx context.Context, muterUserID, mutedUserID uuid.UUID) (bool, error) {
	muted, err := r.q.IsUserMuted(ctx, db.IsUserMutedParams{MuterUserID: muterUserID, MutedUserID: mutedUserID})
	if err != nil {
		return false, fmt.Errorf("checking mute status: %w", err)
	}
	return muted, nil
}
