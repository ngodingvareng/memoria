package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
)

// pendingUsernameInviteTTL is how long a held-back direct invite stays
// answerable before it lapses on its own (FEATURES.md, Circle Invite:
// "Left unanswered, that invite expires on its own rather than hanging
// in their notifications indefinitely"). Not specified precisely by
// FEATURES.md; 30 days mirrors the general shape of an invite people
// are expected to notice within a normal usage cadence.
const pendingUsernameInviteTTL = 30 * 24 * time.Hour

// --- Repository interface, defined here per the Dependency Rule ---

type CircleInviteRepository interface {
	CreateUsernameInvite(ctx context.Context, circleID, invitedByUserID, inviteeUserID uuid.UUID, expiresAt time.Time) (*entity.CircleInvite, error)
	GetPendingUsernameInvite(ctx context.Context, circleID, inviteeUserID uuid.UUID) (*entity.CircleInvite, error)
	ListPendingByInviteeID(ctx context.Context, inviteeUserID uuid.UUID) ([]*entity.CircleInvite, error)
	// AcceptUsernameInvite/DeclineUsernameInvite return errs.ErrNotFound
	// if the invite was already answered, revoked, or has lapsed —
	// indistinguishable from a wrong id, same "no row = 404" reasoning
	// used throughout.
	AcceptUsernameInvite(ctx context.Context, id, inviteeUserID uuid.UUID) (*entity.CircleInvite, error)
	DeclineUsernameInvite(ctx context.Context, id, inviteeUserID uuid.UUID) (*entity.CircleInvite, error)
	RevokeUsernameInvite(ctx context.Context, id, circleID uuid.UUID) (*entity.CircleInvite, error)

	// RevokeActiveLink returns nil, nil when the Circle has no live link
	// yet — the ordinary first-generation case, not an error.
	RevokeActiveLink(ctx context.Context, circleID uuid.UUID) (*entity.CircleInvite, error)
	CreateLink(ctx context.Context, circleID, invitedByUserID uuid.UUID, tokenHash string, requiresApproval bool) (*entity.CircleInvite, error)
	GetActiveLinkByCircleID(ctx context.Context, circleID uuid.UUID) (*entity.CircleInvite, error)
	SetLinkRequiresApproval(ctx context.Context, circleID uuid.UUID, requiresApproval bool) (*entity.CircleInvite, error)

	// AddMembers admits people directly — used both for the "policy
	// clears" branch of InviteDirect and for AcceptInvite.
	AddMembers(ctx context.Context, circleID uuid.UUID, userIDs []uuid.UUID) ([]*entity.CircleMember, error)

	WithTransaction(ctx context.Context, fn func(CircleInviteRepository) error) error
}

// CircleInviteLinkResolver is the minimal capability
// CircleJoinRequestUsecase needs: resolving a followed link's raw token
// to its Circle. Declared here (the owner domain) per CODING_STANDARDS.md
// §5 — same shape as ThreadAccessChecker/MomentAccessChecker.
type CircleInviteLinkResolver interface {
	GetActiveLinkByTokenHash(ctx context.Context, tokenHash string) (*entity.CircleInvite, error)
}

// --- Inputs / outputs ---

type InviteDirectInput struct {
	CircleID        uuid.UUID
	InvitedByUserID uuid.UUID
	Usernames       []string
}

// InviteDirectResult separates the two outcomes FEATURES.md describes:
// AddedMembers joined immediately (circle_invite_policy = "anyone"),
// PendingInvites were held back for the recipient's own acceptance
// (any other policy). SkippedUsernames covers everything that could not
// even be attempted (unknown username, not discoverable, blocked).
type InviteDirectResult struct {
	AddedMembers     []*entity.CircleMember
	PendingInvites   []*entity.CircleInvite
	SkippedUsernames []string
}

type CreateOrRotateInviteLinkInput struct {
	CircleID         uuid.UUID
	UserID           uuid.UUID
	RequiresApproval bool
}

// CreateOrRotateInviteLinkResult carries the raw token alongside the
// entity — the raw value is returned to the caller exactly once and
// never recoverable afterward (only its hash is stored).
type CreateOrRotateInviteLinkResult struct {
	Invite   *entity.CircleInvite
	RawToken string
}

// --- Usecase ---

type CircleInviteUsecase interface {
	InviteDirect(ctx context.Context, input InviteDirectInput) (*InviteDirectResult, error)
	ListMyPendingInvites(ctx context.Context, userID uuid.UUID) ([]*entity.CircleInvite, error)
	AcceptInvite(ctx context.Context, inviteID, userID uuid.UUID) (*entity.CircleMember, error)
	DeclineInvite(ctx context.Context, inviteID, userID uuid.UUID) (*entity.CircleInvite, error)
	RevokeInvite(ctx context.Context, inviteID, circleID, revokedByUserID uuid.UUID) (*entity.CircleInvite, error)

	GetInviteLink(ctx context.Context, circleID, userID uuid.UUID) (*entity.CircleInvite, error)
	CreateOrRotateInviteLink(ctx context.Context, input CreateOrRotateInviteLinkInput) (*CreateOrRotateInviteLinkResult, error)
	SetInviteLinkRequiresApproval(ctx context.Context, circleID, userID uuid.UUID, requiresApproval bool) (*entity.CircleInvite, error)
}

type circleInviteUsecase struct {
	repo          CircleInviteRepository
	circles       CircleAccessChecker
	users         UserPolicyReader
	blocks        UserBlockChecker
	tokens        RefreshTokenGenerator
	notifications NotificationCreator
}

func NewCircleInviteUsecase(repo CircleInviteRepository, circles CircleAccessChecker, users UserPolicyReader, blocks UserBlockChecker, tokens RefreshTokenGenerator, notifications NotificationCreator) CircleInviteUsecase {
	return &circleInviteUsecase{repo: repo, circles: circles, users: users, blocks: blocks, tokens: tokens, notifications: notifications}
}

// requireCanInvite gates every entry point in the Circle Invite flow
// (FEATURES.md, Circle Invite): adding users directly, generating or
// rotating the link, and revoking either.
func (u *circleInviteUsecase) requireCanInvite(ctx context.Context, circleID, userID uuid.UUID) error {
	member, err := u.circles.GetActiveMember(ctx, circleID, userID)
	if err != nil {
		return fmt.Errorf("checking circle membership: %w", err)
	}
	if !member.CanInvite {
		return errs.ErrInsufficientPermission
	}
	return nil
}
