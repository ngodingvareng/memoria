package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
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
	repo    CircleInviteRepository
	circles CircleAccessChecker
	users   UserPolicyReader
	blocks  UserBlockChecker
	tokens  RefreshTokenGenerator
}

func NewCircleInviteUsecase(repo CircleInviteRepository, circles CircleAccessChecker, users UserPolicyReader, blocks UserBlockChecker, tokens RefreshTokenGenerator) CircleInviteUsecase {
	return &circleInviteUsecase{repo: repo, circles: circles, users: users, blocks: blocks, tokens: tokens}
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

// InviteDirect implements [CircleInviteUsecase]. Per username: resolve
// the user, skip anyone unreachable (not found, opted out of
// discoverability, blocked either direction), then split the rest by
// circle_invite_policy — "anyone" joins immediately, anything else
// becomes a held-back pending invite (FEATURES.md, Circle Invite).
func (u *circleInviteUsecase) InviteDirect(ctx context.Context, input InviteDirectInput) (*InviteDirectResult, error) {
	if err := u.requireCanInvite(ctx, input.CircleID, input.InvitedByUserID); err != nil {
		return nil, err
	}

	var toAdd []uuid.UUID
	var toInvite []uuid.UUID
	result := &InviteDirectResult{}

	for _, username := range input.Usernames {
		target, err := u.users.GetByUsername(ctx, username)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				result.SkippedUsernames = append(result.SkippedUsernames, username)
				continue
			}
			return nil, fmt.Errorf("resolving invitee username: %w", err)
		}
		if !target.DiscoverableByUsername {
			result.SkippedUsernames = append(result.SkippedUsernames, username)
			continue
		}
		blocked, err := u.blocks.IsBlockedEitherDirection(ctx, input.InvitedByUserID, target.ID)
		if err != nil {
			return nil, fmt.Errorf("checking block status: %w", err)
		}
		if blocked {
			result.SkippedUsernames = append(result.SkippedUsernames, username)
			continue
		}

		if target.CircleInvitePolicy == enum.AudiencePolicyAnyone {
			toAdd = append(toAdd, target.ID)
		} else {
			toInvite = append(toInvite, target.ID)
		}
	}

	err := u.repo.WithTransaction(ctx, func(tx CircleInviteRepository) error {
		if len(toAdd) > 0 {
			members, err := tx.AddMembers(ctx, input.CircleID, toAdd)
			if err != nil {
				return err
			}
			result.AddedMembers = members
		}
		expiresAt := time.Now().Add(pendingUsernameInviteTTL)
		for _, inviteeID := range toInvite {
			invite, err := tx.CreateUsernameInvite(ctx, input.CircleID, input.InvitedByUserID, inviteeID, expiresAt)
			if err != nil {
				return err
			}
			result.PendingInvites = append(result.PendingInvites, invite)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("inviting members: %w", err)
	}
	return result, nil
}

// ListMyPendingInvites implements [CircleInviteUsecase].
func (u *circleInviteUsecase) ListMyPendingInvites(ctx context.Context, userID uuid.UUID) ([]*entity.CircleInvite, error) {
	invites, err := u.repo.ListPendingByInviteeID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing pending invites: %w", err)
	}
	return invites, nil
}

// AcceptInvite implements [CircleInviteUsecase]. Accepting the invite
// row and seating the membership run in the same transaction — an
// accepted invite with no resulting membership would be a stuck state
// no retry can recover from.
func (u *circleInviteUsecase) AcceptInvite(ctx context.Context, inviteID, userID uuid.UUID) (*entity.CircleMember, error) {
	var member *entity.CircleMember
	err := u.repo.WithTransaction(ctx, func(tx CircleInviteRepository) error {
		invite, err := tx.AcceptUsernameInvite(ctx, inviteID, userID)
		if err != nil {
			return err
		}
		members, err := tx.AddMembers(ctx, invite.CircleID, []uuid.UUID{userID})
		if err != nil {
			return err
		}
		if len(members) > 0 {
			member = members[0]
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("accepting circle invite: %w", err)
	}
	return member, nil
}

// DeclineInvite implements [CircleInviteUsecase].
func (u *circleInviteUsecase) DeclineInvite(ctx context.Context, inviteID, userID uuid.UUID) (*entity.CircleInvite, error) {
	invite, err := u.repo.DeclineUsernameInvite(ctx, inviteID, userID)
	if err != nil {
		return nil, fmt.Errorf("declining circle invite: %w", err)
	}
	return invite, nil
}

// RevokeInvite implements [CircleInviteUsecase]: the inviting side
// withdrawing an invite the recipient hasn't answered yet.
func (u *circleInviteUsecase) RevokeInvite(ctx context.Context, inviteID, circleID, revokedByUserID uuid.UUID) (*entity.CircleInvite, error) {
	if err := u.requireCanInvite(ctx, circleID, revokedByUserID); err != nil {
		return nil, err
	}
	invite, err := u.repo.RevokeUsernameInvite(ctx, inviteID, circleID)
	if err != nil {
		return nil, fmt.Errorf("revoking circle invite: %w", err)
	}
	return invite, nil
}

// GetInviteLink implements [CircleInviteUsecase]. Viewing the existing
// link is available to any active member, not just invite-privileged
// ones — generating/rotating/toggling approval is the privileged action.
func (u *circleInviteUsecase) GetInviteLink(ctx context.Context, circleID, userID uuid.UUID) (*entity.CircleInvite, error) {
	if _, err := u.circles.GetActiveMember(ctx, circleID, userID); err != nil {
		return nil, fmt.Errorf("checking circle membership: %w", err)
	}
	invite, err := u.repo.GetActiveLinkByCircleID(ctx, circleID)
	if err != nil {
		return nil, fmt.Errorf("getting circle invite link: %w", err)
	}
	return invite, nil
}

// CreateOrRotateInviteLink implements [CircleInviteUsecase]. Rotation is
// the only revocation mechanism a link has (FEATURES.md, Circle
// Invite): revoking the old one and creating the new one run in the
// same transaction, since uq_circle_invites_active_link permits only
// one live link per Circle at a time.
func (u *circleInviteUsecase) CreateOrRotateInviteLink(ctx context.Context, input CreateOrRotateInviteLinkInput) (*CreateOrRotateInviteLinkResult, error) {
	if err := u.requireCanInvite(ctx, input.CircleID, input.UserID); err != nil {
		return nil, err
	}

	rawToken, err := u.tokens.Generate()
	if err != nil {
		return nil, fmt.Errorf("generating invite link token: %w", err)
	}
	tokenHash := u.tokens.Hash(rawToken)

	var invite *entity.CircleInvite
	err = u.repo.WithTransaction(ctx, func(tx CircleInviteRepository) error {
		if _, err := tx.RevokeActiveLink(ctx, input.CircleID); err != nil {
			return err
		}
		var createErr error
		invite, createErr = tx.CreateLink(ctx, input.CircleID, input.UserID, tokenHash, input.RequiresApproval)
		return createErr
	})
	if err != nil {
		return nil, fmt.Errorf("creating circle invite link: %w", err)
	}
	return &CreateOrRotateInviteLinkResult{Invite: invite, RawToken: rawToken}, nil
}

// SetInviteLinkRequiresApproval implements [CircleInviteUsecase].
// Requests already open when this toggles stay open either way —
// turning approval off does not admit them retroactively (FEATURES.md,
// Circle Invite).
func (u *circleInviteUsecase) SetInviteLinkRequiresApproval(ctx context.Context, circleID, userID uuid.UUID, requiresApproval bool) (*entity.CircleInvite, error) {
	if err := u.requireCanInvite(ctx, circleID, userID); err != nil {
		return nil, err
	}
	invite, err := u.repo.SetLinkRequiresApproval(ctx, circleID, requiresApproval)
	if err != nil {
		return nil, fmt.Errorf("setting circle invite link approval requirement: %w", err)
	}
	return invite, nil
}
