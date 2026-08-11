package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

var _ usecase.ReactionRepository = (*reactionRepository)(nil)

type reactionRepository struct {
	q *db.Queries
}

func NewReactionRepository(pool *pgxpool.Pool) *reactionRepository {
	return &reactionRepository{q: db.New(pool)}
}

// Upsert implements [usecase.ReactionRepository]. Reacting again with a
// different kind just changes it (FEATURES.md, Response) — there is no
// separate update path.
func (r *reactionRepository) Upsert(ctx context.Context, momentID, userID uuid.UUID, circleID *uuid.UUID, kind enum.ReactionKind) (*entity.Reaction, error) {
	row, err := r.q.UpsertReaction(ctx, db.UpsertReactionParams{
		MomentID: momentID, UserID: ptrToPgUUID(&userID), CircleID: ptrToPgUUID(circleID), Kind: db.ReactionKind(kind),
	})
	if err != nil {
		return nil, fmt.Errorf("upsert reaction: %w", err)
	}
	return toEntityReaction(row), nil
}

// Delete implements [usecase.ReactionRepository]. :exec — a WHERE
// clause matching zero rows (never reacted, already un-reacted) is a
// silent no-op.
func (r *reactionRepository) Delete(ctx context.Context, momentID, userID uuid.UUID, circleID *uuid.UUID) error {
	if err := r.q.DeleteReaction(ctx, db.DeleteReactionParams{
		MomentID: momentID, UserID: ptrToPgUUID(&userID), CircleID: ptrToPgUUID(circleID),
	}); err != nil {
		return fmt.Errorf("delete reaction: %w", err)
	}
	return nil
}

// ListForMoment implements [usecase.ReactionRepository]. Returns the
// rows (who reacted + kind) — counts are never materialized
// (FEATURES.md, Response: "Reaction counts are never displayed as a
// number").
func (r *reactionRepository) ListForMoment(ctx context.Context, momentID, viewerID uuid.UUID, mentionAllowed bool, visibleCircleIDs []uuid.UUID) ([]*entity.Reaction, error) {
	if visibleCircleIDs == nil {
		visibleCircleIDs = []uuid.UUID{}
	}
	rows, err := r.q.ListReactionsForMoment(ctx, db.ListReactionsForMomentParams{
		MomentID: momentID, MentionAllowed: mentionAllowed, VisibleCircleIds: visibleCircleIDs, ViewerID: viewerID,
	})
	if err != nil {
		return nil, fmt.Errorf("list reactions for moment: %w", err)
	}
	return toEntityReactions(rows), nil
}

// AnonymizeByUserID implements [usecase.ReactionRepository].
func (r *reactionRepository) AnonymizeByUserID(ctx context.Context, userID uuid.UUID) error {
	if err := r.q.AnonymizeReactionsByUserID(ctx, ptrToPgUUID(&userID)); err != nil {
		return fmt.Errorf("anonymize reactions by user id: %w", err)
	}
	return nil
}
