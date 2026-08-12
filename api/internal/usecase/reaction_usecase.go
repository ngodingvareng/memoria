package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
)

// --- Repository interface, defined here per the Dependency Rule ---

type ReactionRepository interface {
	Upsert(
		ctx context.Context,
		momentID, userID uuid.UUID,
		circleID *uuid.UUID,
		kind enum.ReactionKind,
	) (*entity.Reaction, error)
	// Delete is a sqlc :exec query — a WHERE clause matching zero rows
	// is a silent no-op.
	Delete(ctx context.Context, momentID, userID uuid.UUID, circleID *uuid.UUID) error
	// ListForMoment returns the rows (who reacted + kind); counts are
	// never materialized (FEATURES.md, Response).
	ListForMoment(
		ctx context.Context,
		momentID, viewerID uuid.UUID,
		mentionAllowed bool,
		visibleCircleIDs []uuid.UUID,
	) ([]*entity.Reaction, error)
	// AnonymizeByUserID is account deletion's counterpart to
	// CommentRepository.AnonymizeByUserID — same reasoning (FEATURES.md,
	// Lifecycle & Deletion).
	AnonymizeByUserID(ctx context.Context, userID uuid.UUID) error
}

// --- Inputs / outputs ---

type SetReactionInput struct {
	MomentID uuid.UUID
	UserID   uuid.UUID
	// CircleID nil means the mention context — see CreateCommentInput.
	CircleID *uuid.UUID
	Kind     enum.ReactionKind
}

type RemoveReactionInput struct {
	MomentID uuid.UUID
	UserID   uuid.UUID
	CircleID *uuid.UUID
}

// --- Usecase ---

type ReactionUsecase interface {
	SetReaction(ctx context.Context, input SetReactionInput) (*entity.Reaction, error)
	RemoveReaction(ctx context.Context, input RemoveReactionInput) error
	ListReactions(ctx context.Context, momentID, viewerID uuid.UUID) ([]*entity.Reaction, error)
}

type reactionUsecase struct {
	repo   ReactionRepository
	access *ResponseAccessChecker
	events ResponseEventRepository
}

func NewReactionUsecase(
	repo ReactionRepository,
	access *ResponseAccessChecker,
	events ResponseEventRepository,
) ReactionUsecase {
	return &reactionUsecase{repo: repo, access: access, events: events}
}

// SetReaction implements [ReactionUsecase]. Same context-authorization
// and response_events shape as CommentUsecase.CreateComment — see its
// comment for why there's no shared transaction helper here.
func (u *reactionUsecase) SetReaction(ctx context.Context, input SetReactionInput) (*entity.Reaction, error) {
	audience, err := requireAudience(ctx, u.access, input.MomentID, input.UserID, input.CircleID)
	if err != nil {
		return nil, fmt.Errorf("checking response access: %w", err)
	}

	reaction, err := u.repo.Upsert(ctx, input.MomentID, input.UserID, input.CircleID, input.Kind)
	if err != nil {
		return nil, fmt.Errorf("setting reaction: %w", err)
	}

	if audience.Moment.UserID != input.UserID {
		if err := u.events.Create(ctx, &entity.ResponseEvent{
			RecipientUserID: audience.Moment.UserID,
			ActorUserID:     &input.UserID,
			MomentID:        input.MomentID,
			Kind:            enum.ResponseEventKindReaction,
			ReactionID:      &reaction.ID,
		}); err != nil {
			return nil, fmt.Errorf("recording response event: %w", err)
		}
	}

	return reaction, nil
}

// RemoveReaction implements [ReactionUsecase]: un-reacting.
func (u *reactionUsecase) RemoveReaction(ctx context.Context, input RemoveReactionInput) error {
	if err := u.repo.Delete(ctx, input.MomentID, input.UserID, input.CircleID); err != nil {
		return fmt.Errorf("removing reaction: %w", err)
	}
	return nil
}

// ListReactions implements [ReactionUsecase].
func (u *reactionUsecase) ListReactions(ctx context.Context, momentID, viewerID uuid.UUID) ([]*entity.Reaction, error) {
	audience, err := u.access.ResolveAudience(ctx, momentID, viewerID)
	if err != nil {
		return nil, fmt.Errorf("resolving response access: %w", err)
	}

	reactions, err := u.repo.ListForMoment(ctx, momentID, viewerID, audience.MentionAllowed, audience.CircleIDs)
	if err != nil {
		return nil, fmt.Errorf("listing reactions: %w", err)
	}
	return reactions, nil
}
