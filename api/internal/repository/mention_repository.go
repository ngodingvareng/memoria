package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

var _ usecase.MentionRepository = (*mentionRepository)(nil)

// mentionRepository covers both moment_mentions and moment_circles —
// FEATURES.md frames the Circle-share step as part of the mention flow
// itself, not a separate entity ("the mention flow's optional 'Share to
// circle too?' step... never a standalone 'share' action"). Neither
// side needs a WithTransaction of its own: each write here is already a
// single-row, single-statement operation.
type mentionRepository struct {
	q *db.Queries
}

func NewMentionRepository(pool *pgxpool.Pool) *mentionRepository {
	return &mentionRepository{q: db.New(pool)}
}

// Create implements [usecase.MentionRepository].
func (r *mentionRepository) Create(ctx context.Context, momentID, mentionedUserID uuid.UUID, displayName string) (*entity.MomentMention, error) {
	row, err := r.q.CreateMomentMention(ctx, db.CreateMomentMentionParams{
		MomentID: momentID, MentionedUserID: ptrToPgUUID(&mentionedUserID), DisplayName: displayName,
	})
	if err != nil {
		return nil, fmt.Errorf("create moment mention: %w", err)
	}
	return toEntityMomentMention(row), nil
}

// ListByMomentID implements [usecase.MentionRepository].
func (r *mentionRepository) ListByMomentID(ctx context.Context, momentID uuid.UUID) ([]*entity.MomentMention, error) {
	rows, err := r.q.ListMomentMentionsByMomentID(ctx, momentID)
	if err != nil {
		return nil, fmt.Errorf("list moment mentions: %w", err)
	}
	return toEntityMomentMentions(rows), nil
}

// ListMentionedMoments implements [usecase.MentionRepository]:
// "Mentioned Moments" (FEATURES.md, Looking Back).
func (r *mentionRepository) ListMentionedMoments(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*entity.Moment, error) {
	rows, err := r.q.ListMentionedMomentsByUserID(ctx, db.ListMentionedMomentsByUserIDParams{
		MentionedUserID: ptrToPgUUID(&userID), PageLimit: limit, PageOffset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list mentioned moments: %w", err)
	}
	return toEntityMoments(rows), nil
}

// Remove implements [usecase.MentionRepository]: "Leaving a mention" —
// the mentioned user removing themselves. :exec, silent no-op if
// already removed or nonexistent (nothing further to signal to a caller
// who is already acting on their own membership).
func (r *mentionRepository) Remove(ctx context.Context, momentID, mentionedUserID uuid.UUID) error {
	if err := r.q.RemoveMomentMention(ctx, db.RemoveMomentMentionParams{
		MomentID: momentID, MentionedUserID: ptrToPgUUID(&mentionedUserID),
	}); err != nil {
		return fmt.Errorf("remove moment mention: %w", err)
	}
	return nil
}

// Delete implements [usecase.MentionRepository]: the owner deleting a
// mention they added by mistake.
func (r *mentionRepository) Delete(ctx context.Context, id, momentID uuid.UUID) error {
	if err := r.q.DeleteMomentMention(ctx, db.DeleteMomentMentionParams{ID: id, MomentID: momentID}); err != nil {
		return fmt.Errorf("delete moment mention: %w", err)
	}
	return nil
}

// ShareToCircle implements [usecase.MentionRepository]. Returns nil,
// nil when the Moment is already shared into circleID — idempotent, per
// the query's own comment, not an error.
func (r *mentionRepository) ShareToCircle(ctx context.Context, momentID, circleID, sharedByUserID uuid.UUID) (*entity.MomentCircle, error) {
	row, err := r.q.ShareMomentToCircle(ctx, db.ShareMomentToCircleParams{
		MomentID: momentID, CircleID: circleID, SharedByUserID: ptrToPgUUID(&sharedByUserID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("share moment to circle: %w", err)
	}
	return toEntityMomentCircle(row), nil
}

// ListSharedCircleIDs implements [usecase.MentionRepository].
func (r *mentionRepository) ListSharedCircleIDs(ctx context.Context, momentID uuid.UUID) ([]uuid.UUID, error) {
	ids, err := r.q.ListCircleIDsByMomentID(ctx, momentID)
	if err != nil {
		return nil, fmt.Errorf("list moment shared circle ids: %w", err)
	}
	return ids, nil
}

// UnshareFromCircle implements [usecase.MentionRepository].
func (r *mentionRepository) UnshareFromCircle(ctx context.Context, momentID, circleID uuid.UUID) error {
	if err := r.q.UnshareMomentFromCircle(ctx, db.UnshareMomentFromCircleParams{MomentID: momentID, CircleID: circleID}); err != nil {
		return fmt.Errorf("unshare moment from circle: %w", err)
	}
	return nil
}
