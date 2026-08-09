package usecase

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
	"github.com/ngodingvareng/memoria/internal/errs"
)

// --- Repository interfaces, defined here per the Dependency Rule ---

type CommentRepository interface {
	Create(ctx context.Context, momentID, userID uuid.UUID, circleID *uuid.UUID, body string) (*entity.Comment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Comment, error)
	ListForMoment(ctx context.Context, momentID, viewerID uuid.UUID, mentionAllowed bool, visibleCircleIDs []uuid.UUID) ([]*entity.Comment, error)
	// Update/SoftDelete return errs.ErrNotFound / are a silent no-op —
	// see comment_repository.go for the exact split.
	Update(ctx context.Context, id, userID uuid.UUID, body string) (*entity.Comment, error)
	SoftDelete(ctx context.Context, id, userID uuid.UUID) error
}

// ResponseEventRepository writes to the batched daily digest queue
// (FEATURES.md, Notification). Declared here (Comment is the first
// consumer) and reused as-is by reaction_usecase.go — same signature,
// CODING_STANDARDS.md §5.
type ResponseEventRepository interface {
	Create(ctx context.Context, event *entity.ResponseEvent) error
}

// --- Inputs / outputs ---

type CreateCommentInput struct {
	MomentID uuid.UUID
	UserID   uuid.UUID
	// CircleID nil means the mention context; set, it names which
	// Circle audience the comment is posted in (FEATURES.md, Response).
	CircleID *uuid.UUID
	Body     string
}

type UpdateCommentInput struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Body   string
}

// --- Usecase ---

type CommentUsecase interface {
	CreateComment(ctx context.Context, input CreateCommentInput) (*entity.Comment, error)
	ListComments(ctx context.Context, momentID, viewerID uuid.UUID) ([]*entity.Comment, error)
	UpdateComment(ctx context.Context, input UpdateCommentInput) (*entity.Comment, error)
	SoftDeleteComment(ctx context.Context, id, userID uuid.UUID) error
	// GetAudience resolves which contexts viewerID may act/read in for
	// momentID (FEATURES.md, Response) — which circle_id(s) a comment or
	// reaction composer should offer, and whether the mention context is
	// available. A thin passthrough to ResponseAccessChecker, exposed
	// here since Comment is its first consumer (see response_access.go).
	GetAudience(ctx context.Context, momentID, viewerID uuid.UUID) (*ResponseAudience, error)
}

type commentUsecase struct {
	repo   CommentRepository
	access *ResponseAccessChecker
	events ResponseEventRepository
}

func NewCommentUsecase(repo CommentRepository, access *ResponseAccessChecker, events ResponseEventRepository) CommentUsecase {
	return &commentUsecase{repo: repo, access: access, events: events}
}

// requireAudience resolves the viewer's audience for momentID and
// checks circleID (nil = mention context) is one they may act/read in
// — the same rule CreateComment and ListComments both need.
func requireAudience(ctx context.Context, access *ResponseAccessChecker, momentID, viewerID uuid.UUID, circleID *uuid.UUID) (*ResponseAudience, error) {
	audience, err := access.ResolveAudience(ctx, momentID, viewerID)
	if err != nil {
		return nil, err
	}
	if circleID == nil {
		if !audience.MentionAllowed {
			return nil, errs.ErrForbidden
		}
		return audience, nil
	}
	if !slices.Contains(audience.CircleIDs, *circleID) {
		return nil, errs.ErrForbidden
	}
	return audience, nil
}

// CreateComment implements [CommentUsecase]. Writes the comment and a
// response_events row (skipped when the actor is the Moment's own
// owner — reacting/commenting on your own Moment is not a response
// event) as two statements; there's no cross-repository transaction
// helper here (unlike Thread/Moment, Comment has no multi-repo
// WithTransaction — a lost response_events row only delays a digest
// notification, not a correctness issue, so the extra machinery isn't
// worth it).
func (u *commentUsecase) CreateComment(ctx context.Context, input CreateCommentInput) (*entity.Comment, error) {
	audience, err := requireAudience(ctx, u.access, input.MomentID, input.UserID, input.CircleID)
	if err != nil {
		return nil, fmt.Errorf("checking response access: %w", err)
	}

	comment, err := u.repo.Create(ctx, input.MomentID, input.UserID, input.CircleID, input.Body)
	if err != nil {
		return nil, fmt.Errorf("creating comment: %w", err)
	}

	if audience.Moment.UserID != input.UserID {
		if err := u.events.Create(ctx, &entity.ResponseEvent{
			RecipientUserID: audience.Moment.UserID,
			ActorUserID:     &input.UserID,
			MomentID:        input.MomentID,
			Kind:            enum.ResponseEventKindComment,
			CommentID:       &comment.ID,
		}); err != nil {
			return nil, fmt.Errorf("recording response event: %w", err)
		}
	}

	return comment, nil
}

// ListComments implements [CommentUsecase]: the full set of comments
// visible to viewerID across every context they have access to.
func (u *commentUsecase) ListComments(ctx context.Context, momentID, viewerID uuid.UUID) ([]*entity.Comment, error) {
	audience, err := u.access.ResolveAudience(ctx, momentID, viewerID)
	if err != nil {
		return nil, fmt.Errorf("resolving response access: %w", err)
	}

	comments, err := u.repo.ListForMoment(ctx, momentID, viewerID, audience.MentionAllowed, audience.CircleIDs)
	if err != nil {
		return nil, fmt.Errorf("listing comments: %w", err)
	}
	return comments, nil
}

// UpdateComment implements [CommentUsecase].
func (u *commentUsecase) UpdateComment(ctx context.Context, input UpdateCommentInput) (*entity.Comment, error) {
	comment, err := u.repo.Update(ctx, input.ID, input.UserID, input.Body)
	if err != nil {
		return nil, fmt.Errorf("updating comment: %w", err)
	}
	return comment, nil
}

// SoftDeleteComment implements [CommentUsecase].
func (u *commentUsecase) SoftDeleteComment(ctx context.Context, id, userID uuid.UUID) error {
	if err := u.repo.SoftDelete(ctx, id, userID); err != nil {
		return fmt.Errorf("soft-deleting comment: %w", err)
	}
	return nil
}

// GetAudience implements [CommentUsecase].
func (u *commentUsecase) GetAudience(ctx context.Context, momentID, viewerID uuid.UUID) (*ResponseAudience, error) {
	audience, err := u.access.ResolveAudience(ctx, momentID, viewerID)
	if err != nil {
		return nil, fmt.Errorf("resolving audience: %w", err)
	}
	return audience, nil
}
