package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
)

// MomentReader is the minimal capability CommentUsecase/ReactionUsecase
// need from the Moment domain: reading a Moment for a viewer who is
// frequently not its owner, and resolving which Circle audiences that
// viewer shares with it. Satisfied by moment_repository.go.
type MomentReader interface {
	// GetWithAccess returns errs.ErrNotFound if the Moment doesn't exist,
	// is soft-deleted, or viewerID has no path to it (owner, active
	// mention, Circle share, or native collaborative-Thread membership).
	GetWithAccess(ctx context.Context, id, viewerID uuid.UUID) (*entity.Moment, error)
	// ListVisibleCircleIDs returns which Circle audiences viewerID shares
	// with this Moment (its native Thread's Circle, plus every Circle
	// it's been shared into, intersected with viewerID's active
	// memberships).
	ListVisibleCircleIDs(ctx context.Context, momentID, viewerID uuid.UUID) ([]uuid.UUID, error)
	// IsActivelyMentioned reports whether viewerID has a non-removed
	// mention on this Moment — the non-owner half of the mention
	// context's eligibility.
	IsActivelyMentioned(ctx context.Context, momentID, viewerID uuid.UUID) (bool, error)
}

// ResponseAudience is what ResponseAccessChecker resolves: which of the
// three FEATURES.md Response contexts (see the Response section) a
// viewer may act/read in for a given Moment.
type ResponseAudience struct {
	Moment *entity.Moment
	// MentionAllowed is true when the viewer may act/read in the
	// mention context (circle_id NULL): the Moment's owner, or an
	// active (non-removed) mentioned user.
	MentionAllowed bool
	// CircleIDs are the Circle audiences (circle_id set) the viewer may
	// act/read in for this Moment.
	CircleIDs []uuid.UUID
}

// ResponseAccessChecker resolves Comment/Reaction visibility, shared by
// comment_usecase.go and reaction_usecase.go rather than duplicated
// (CODING_STANDARDS.md §5) since both need the exact same three-context
// resolution and blocking gate.
type ResponseAccessChecker struct {
	moments MomentReader
	blocks  UserBlockChecker
}

func NewResponseAccessChecker(moments MomentReader, blocks UserBlockChecker) *ResponseAccessChecker {
	return &ResponseAccessChecker{moments: moments, blocks: blocks}
}

// ResolveAudience fetches the Moment (scoped to what viewerID can
// reach), then checks blocking between viewerID and the Moment's owner
// — "Blocked users can never view, comment, or react to the other's
// Moments, in any context, in either direction" (FEATURES.md, Response
// Terms) — before resolving which contexts the viewer may act/read in.
func (r *ResponseAccessChecker) ResolveAudience(
	ctx context.Context,
	momentID, viewerID uuid.UUID,
) (*ResponseAudience, error) {
	moment, err := r.moments.GetWithAccess(ctx, momentID, viewerID)
	if err != nil {
		return nil, fmt.Errorf("resolving moment access: %w", err)
	}

	blocked, err := r.blocks.IsBlockedEitherDirection(ctx, viewerID, moment.UserID)
	if err != nil {
		return nil, fmt.Errorf("checking block status: %w", err)
	}
	if blocked {
		return nil, errs.ErrAccessDenied
	}

	circleIDs, err := r.moments.ListVisibleCircleIDs(ctx, momentID, viewerID)
	if err != nil {
		return nil, fmt.Errorf("listing visible circle ids: %w", err)
	}

	// The mention context (circle_id IS NULL) is visible to the owner
	// and to an active mentioned user — never to a Circle member who
	// isn't also one of those two (FEATURES.md, Response Terms:
	// "Visibility stays scoped to context").
	mentionAllowed := moment.UserID == viewerID
	if !mentionAllowed {
		mentionAllowed, err = r.moments.IsActivelyMentioned(ctx, momentID, viewerID)
		if err != nil {
			return nil, fmt.Errorf("checking mention status: %w", err)
		}
	}

	return &ResponseAudience{Moment: moment, MentionAllowed: mentionAllowed, CircleIDs: circleIDs}, nil
}
