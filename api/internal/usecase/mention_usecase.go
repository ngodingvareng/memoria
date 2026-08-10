package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
	"github.com/ngodingvareng/memoria/internal/errs"
)

// mentionSearchOverfetchLimit is how many discoverable candidates
// SearchMentionableUsers pulls before narrowing by mention_policy/
// blocking; mentionSearchResultLimit is what actually reaches the
// caller once that filtering is applied.
const (
	mentionSearchOverfetchLimit = 20
	mentionSearchResultLimit    = 8
)

// --- Repository interface, defined here per the Dependency Rule ---

type MentionRepository interface {
	Create(ctx context.Context, momentID, mentionedUserID uuid.UUID, displayName string) (*entity.MomentMention, error)
	ListByMomentID(ctx context.Context, momentID uuid.UUID) ([]*entity.MomentMention, error)
	ListMentionedMoments(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*entity.Moment, error)
	// Remove is a sqlc :exec query — a WHERE clause matching zero rows
	// (already removed, never mentioned) is a silent no-op.
	Remove(ctx context.Context, momentID, mentionedUserID uuid.UUID) error
	Delete(ctx context.Context, id, momentID uuid.UUID) error

	// ShareToCircle returns nil, nil when already shared — idempotent,
	// not an error (FEATURES.md, Mention).
	ShareToCircle(ctx context.Context, momentID, circleID, sharedByUserID uuid.UUID) (*entity.MomentCircle, error)
	ListSharedCircleIDs(ctx context.Context, momentID uuid.UUID) ([]uuid.UUID, error)
	UnshareFromCircle(ctx context.Context, momentID, circleID uuid.UUID) error
}

// UserSearcher backs the mention-search typeahead (FEATURES.md, Mention)
// — a candidate list of discoverable usernames matching a prefix, which
// SearchMentionableUsers then narrows by mention_policy/blocking, the
// same way CreateMention gates a single resolved username. Satisfied
// structurally by the existing UserRepository (see user_repository.go's
// SearchByUsernamePrefix) — no new repository needed.
type UserSearcher interface {
	SearchByUsernamePrefix(ctx context.Context, excludeUserID uuid.UUID, query string, limit int32) ([]*entity.User, error)
}

// --- Inputs / outputs ---

type CreateMentionInput struct {
	MomentID uuid.UUID
	// OwnerUserID must own the Moment — only the owner adds mentions to
	// it (FEATURES.md, Mention: naming another person present).
	OwnerUserID uuid.UUID
	Username    string
}

type ShareMomentToCircleInput struct {
	MomentID uuid.UUID
	CircleID uuid.UUID
	// UserID must own the Moment and be an active member of the target
	// Circle.
	UserID uuid.UUID
}

type ListMentionedMomentsInput struct {
	UserID   uuid.UUID
	Page     int32
	PageSize int32
}

// CreateMentionResult is CreateMention's return: the new mention, plus
// which Circles the owner and the mentioned user both actively belong
// to — the candidate set for the mention flow's optional "Share to
// circle too?" step (FEATURES.md, Mention), computed in the same call
// so the frontend can offer it immediately with no extra request.
type CreateMentionResult struct {
	Mention         *entity.MomentMention
	SharedCircleIDs []uuid.UUID
}

// --- Usecase ---

type MentionUsecase interface {
	CreateMention(ctx context.Context, input CreateMentionInput) (*CreateMentionResult, error)
	ListMentions(ctx context.Context, momentID, ownerUserID uuid.UUID) ([]*entity.MomentMention, error)
	// LeaveMention is the mentioned user removing themselves — no
	// notification to the owner (FEATURES.md, Leaving a mention).
	LeaveMention(ctx context.Context, momentID, mentionedUserID uuid.UUID) error
	DeleteMention(ctx context.Context, mentionID, momentID, ownerUserID uuid.UUID) error
	ListMentionedMoments(ctx context.Context, input ListMentionedMomentsInput) (*MomentListResult, error)
	// SearchMentionableUsers backs the mention-search typeahead: a
	// prefix match over discoverable usernames the requester is
	// currently allowed to mention (FEATURES.md, Mention + Privacy &
	// Control). Empty query returns an empty slice — no "browse
	// everyone" mode.
	SearchMentionableUsers(ctx context.Context, requestingUserID uuid.UUID, query string) ([]*entity.User, error)

	ShareToCircle(ctx context.Context, input ShareMomentToCircleInput) (*entity.MomentCircle, error)
	UnshareFromCircle(ctx context.Context, momentID, circleID, userID uuid.UUID) error
	// ListSharedCircles is which Circles a personal Moment is currently
	// shared to — owner-only, backs the "manage sharing" surface.
	ListSharedCircles(ctx context.Context, momentID, ownerUserID uuid.UUID) ([]uuid.UUID, error)
}

type mentionUsecase struct {
	repo         MentionRepository
	moments      MomentAccessChecker
	circles      CircleAccessChecker
	circleShares CircleShareChecker
	users        UserPolicyReader
	knowns       UserKnownChecker
	blocks       UserBlockChecker
	searcher     UserSearcher
}

func NewMentionUsecase(repo MentionRepository, moments MomentAccessChecker, circles CircleAccessChecker, circleShares CircleShareChecker, users UserPolicyReader, knowns UserKnownChecker, blocks UserBlockChecker, searcher UserSearcher) MentionUsecase {
	return &mentionUsecase{repo: repo, moments: moments, circles: circles, circleShares: circleShares, users: users, knowns: knowns, blocks: blocks, searcher: searcher}
}

// CreateMention implements [MentionUsecase]. Enforces
// users.mention_policy (FEATURES.md, Privacy & Control: "Who may
// mention a user at all is controlled by that user") and blocking,
// which overrides any policy in both directions.
func (u *mentionUsecase) CreateMention(ctx context.Context, input CreateMentionInput) (*CreateMentionResult, error) {
	if _, err := u.moments.GetByID(ctx, input.MomentID, input.OwnerUserID); err != nil {
		return nil, fmt.Errorf("checking moment ownership: %w", err)
	}

	target, err := u.users.GetByUsername(ctx, input.Username)
	if err != nil {
		return nil, fmt.Errorf("resolving mentioned username: %w", err)
	}

	blocked, err := u.blocks.IsBlockedEitherDirection(ctx, input.OwnerUserID, target.ID)
	if err != nil {
		return nil, fmt.Errorf("checking block status: %w", err)
	}
	if blocked {
		return nil, errs.ErrAccessDenied
	}

	switch target.MentionPolicy {
	case enum.AudiencePolicyAnyone:
		// allowed
	case enum.AudiencePolicyKnown:
		known, err := u.knowns.IsKnownTo(ctx, target.ID, input.OwnerUserID)
		if err != nil {
			return nil, fmt.Errorf("checking known status: %w", err)
		}
		if !known {
			return nil, errs.ErrAccessDenied
		}
	default: // enum.AudiencePolicyNobody
		return nil, errs.ErrAccessDenied
	}

	mention, err := u.repo.Create(ctx, input.MomentID, target.ID, target.Name)
	if err != nil {
		return nil, fmt.Errorf("creating mention: %w", err)
	}

	sharedCircleIDs, err := u.circleShares.ListSharedCircleIDs(ctx, input.OwnerUserID, target.ID)
	if err != nil {
		return nil, fmt.Errorf("listing shared circles: %w", err)
	}

	return &CreateMentionResult{Mention: mention, SharedCircleIDs: sharedCircleIDs}, nil
}

// ListMentions implements [MentionUsecase].
func (u *mentionUsecase) ListMentions(ctx context.Context, momentID, ownerUserID uuid.UUID) ([]*entity.MomentMention, error) {
	if _, err := u.moments.GetByID(ctx, momentID, ownerUserID); err != nil {
		return nil, fmt.Errorf("checking moment ownership: %w", err)
	}
	mentions, err := u.repo.ListByMomentID(ctx, momentID)
	if err != nil {
		return nil, fmt.Errorf("listing mentions: %w", err)
	}
	return mentions, nil
}

// LeaveMention implements [MentionUsecase].
func (u *mentionUsecase) LeaveMention(ctx context.Context, momentID, mentionedUserID uuid.UUID) error {
	if err := u.repo.Remove(ctx, momentID, mentionedUserID); err != nil {
		return fmt.Errorf("leaving mention: %w", err)
	}
	return nil
}

// DeleteMention implements [MentionUsecase].
func (u *mentionUsecase) DeleteMention(ctx context.Context, mentionID, momentID, ownerUserID uuid.UUID) error {
	if _, err := u.moments.GetByID(ctx, momentID, ownerUserID); err != nil {
		return fmt.Errorf("checking moment ownership: %w", err)
	}
	if err := u.repo.Delete(ctx, mentionID, momentID); err != nil {
		return fmt.Errorf("deleting mention: %w", err)
	}
	return nil
}

// ListMentionedMoments implements [MentionUsecase]: "Mentioned Moments"
// (FEATURES.md, Looking Back).
func (u *mentionUsecase) ListMentionedMoments(ctx context.Context, input ListMentionedMomentsInput) (*MomentListResult, error) {
	page, pageSize := normalizedPage(input.Page, input.PageSize)
	moments, err := u.repo.ListMentionedMoments(ctx, input.UserID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("listing mentioned moments: %w", err)
	}
	return &MomentListResult{Moments: moments, Page: page, PageSize: pageSize}, nil
}

// SearchMentionableUsers implements [MentionUsecase]. Applies the same
// per-candidate checks CreateMention applies to a single resolved
// username — blocking (either direction) and mention_policy, including
// resolving audience_policy = "known" via IsKnownTo — just stopping
// once mentionSearchResultLimit matches are found instead of failing on
// the first disallowed one.
func (u *mentionUsecase) SearchMentionableUsers(ctx context.Context, requestingUserID uuid.UUID, query string) ([]*entity.User, error) {
	if strings.TrimSpace(query) == "" {
		return []*entity.User{}, nil
	}

	candidates, err := u.searcher.SearchByUsernamePrefix(ctx, requestingUserID, query, mentionSearchOverfetchLimit)
	if err != nil {
		return nil, fmt.Errorf("searching users by username: %w", err)
	}

	results := make([]*entity.User, 0, mentionSearchResultLimit)
	for _, candidate := range candidates {
		if len(results) >= mentionSearchResultLimit {
			break
		}

		blocked, err := u.blocks.IsBlockedEitherDirection(ctx, requestingUserID, candidate.ID)
		if err != nil {
			return nil, fmt.Errorf("checking block status: %w", err)
		}
		if blocked {
			continue
		}

		switch candidate.MentionPolicy {
		case enum.AudiencePolicyAnyone:
			results = append(results, candidate)
		case enum.AudiencePolicyKnown:
			known, err := u.knowns.IsKnownTo(ctx, candidate.ID, requestingUserID)
			if err != nil {
				return nil, fmt.Errorf("checking known status: %w", err)
			}
			if known {
				results = append(results, candidate)
			}
		default: // enum.AudiencePolicyNobody
		}
	}

	return results, nil
}

// ShareToCircle implements [MentionUsecase]: the mention flow's
// optional "Share to circle too?" step (FEATURES.md, Mention) — never a
// standalone "share" action, and only offered when the owner already
// shares a Circle with someone they mentioned.
func (u *mentionUsecase) ShareToCircle(ctx context.Context, input ShareMomentToCircleInput) (*entity.MomentCircle, error) {
	if _, err := u.moments.GetByID(ctx, input.MomentID, input.UserID); err != nil {
		return nil, fmt.Errorf("checking moment ownership: %w", err)
	}
	if _, err := u.circles.GetActiveMember(ctx, input.CircleID, input.UserID); err != nil {
		return nil, fmt.Errorf("checking circle membership: %w", err)
	}

	share, err := u.repo.ShareToCircle(ctx, input.MomentID, input.CircleID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("sharing moment to circle: %w", err)
	}
	return share, nil
}

// UnshareFromCircle implements [MentionUsecase].
func (u *mentionUsecase) UnshareFromCircle(ctx context.Context, momentID, circleID, userID uuid.UUID) error {
	if _, err := u.moments.GetByID(ctx, momentID, userID); err != nil {
		return fmt.Errorf("checking moment ownership: %w", err)
	}
	if err := u.repo.UnshareFromCircle(ctx, momentID, circleID); err != nil {
		return fmt.Errorf("unsharing moment from circle: %w", err)
	}
	return nil
}

// ListSharedCircles implements [MentionUsecase].
func (u *mentionUsecase) ListSharedCircles(ctx context.Context, momentID, ownerUserID uuid.UUID) ([]uuid.UUID, error) {
	if _, err := u.moments.GetByID(ctx, momentID, ownerUserID); err != nil {
		return nil, fmt.Errorf("checking moment ownership: %w", err)
	}
	ids, err := u.repo.ListSharedCircleIDs(ctx, momentID)
	if err != nil {
		return nil, fmt.Errorf("listing shared circles: %w", err)
	}
	return ids, nil
}
