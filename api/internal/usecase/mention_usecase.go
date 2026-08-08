package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
	"github.com/ngodingvareng/memoria/internal/errs"
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

// --- Usecase ---

type MentionUsecase interface {
	CreateMention(ctx context.Context, input CreateMentionInput) (*entity.MomentMention, error)
	ListMentions(ctx context.Context, momentID, ownerUserID uuid.UUID) ([]*entity.MomentMention, error)
	// LeaveMention is the mentioned user removing themselves — no
	// notification to the owner (FEATURES.md, Leaving a mention).
	LeaveMention(ctx context.Context, momentID, mentionedUserID uuid.UUID) error
	DeleteMention(ctx context.Context, mentionID, momentID, ownerUserID uuid.UUID) error
	ListMentionedMoments(ctx context.Context, input ListMentionedMomentsInput) (*MomentListResult, error)

	ShareToCircle(ctx context.Context, input ShareMomentToCircleInput) (*entity.MomentCircle, error)
	UnshareFromCircle(ctx context.Context, momentID, circleID, userID uuid.UUID) error
}

type mentionUsecase struct {
	repo    MentionRepository
	moments MomentAccessChecker
	circles CircleAccessChecker
	users   UserPolicyReader
	knowns  UserKnownChecker
	blocks  UserBlockChecker
}

func NewMentionUsecase(repo MentionRepository, moments MomentAccessChecker, circles CircleAccessChecker, users UserPolicyReader, knowns UserKnownChecker, blocks UserBlockChecker) MentionUsecase {
	return &mentionUsecase{repo: repo, moments: moments, circles: circles, users: users, knowns: knowns, blocks: blocks}
}

// CreateMention implements [MentionUsecase]. Enforces
// users.mention_policy (FEATURES.md, Privacy & Control: "Who may
// mention a user at all is controlled by that user") and blocking,
// which overrides any policy in both directions.
func (u *mentionUsecase) CreateMention(ctx context.Context, input CreateMentionInput) (*entity.MomentMention, error) {
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
	return mention, nil
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
