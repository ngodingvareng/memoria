package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
	"github.com/ngodingvareng/memoria/internal/errs"
)

// --- Repository interface, defined here per the Dependency Rule ---

type CircleJoinRequestRepository interface {
	Create(ctx context.Context, circleID, userID uuid.UUID, inviteID *uuid.UUID) (*entity.CircleJoinRequest, error)
	// GetPendingByUser backs the "you already asked" check before
	// writing a new request.
	GetPendingByUser(ctx context.Context, circleID, userID uuid.UUID) (*entity.CircleJoinRequest, error)
	ListPendingByCircleID(
		ctx context.Context,
		circleID uuid.UUID,
		limit, offset int32,
	) ([]*entity.CircleJoinRequest, error)
	CountPendingByCircleID(ctx context.Context, circleID uuid.UUID) (int64, error)
	// Approve/Reject/Cancel return errs.ErrNotFound if the request was
	// already decided/withdrawn — same "no row = 404" reasoning used
	// throughout.
	Approve(ctx context.Context, id, circleID, decidedByUserID uuid.UUID) (*entity.CircleJoinRequest, error)
	Reject(ctx context.Context, id, circleID, decidedByUserID uuid.UUID) (*entity.CircleJoinRequest, error)
	Cancel(ctx context.Context, id, userID uuid.UUID) (*entity.CircleJoinRequest, error)

	AddMembers(ctx context.Context, circleID uuid.UUID, userIDs []uuid.UUID) ([]*entity.CircleMember, error)

	WithTransaction(ctx context.Context, fn func(CircleJoinRequestRepository) error) error
}

// CircleMemberLister is the minimal capability this usecase needs to
// find who to notify of a new join request: every active member,
// filtered down to the invite-privileged ones (mirrors ListPending's own
// CanInvite gate). Satisfied structurally by the existing
// CircleRepository (CODING_STANDARDS.md §5) — no new repository needed.
type CircleMemberLister interface {
	ListMembers(ctx context.Context, circleID uuid.UUID) ([]*entity.CircleMember, error)
}

// --- Inputs / outputs ---

// FollowInviteLinkResult reports which of the link's two outcomes
// happened (FEATURES.md, Circle Invite: approval-not-required joins
// immediately, approval-required opens a request).
type FollowInviteLinkResult struct {
	Joined      *entity.CircleMember
	JoinRequest *entity.CircleJoinRequest
}

// --- Usecase ---

type CircleJoinRequestUsecase interface {
	// FollowInviteLink resolves rawToken to its Circle and either joins
	// immediately or opens a join request, depending on the link's
	// requires_approval setting.
	FollowInviteLink(ctx context.Context, rawToken string, userID uuid.UUID) (*FollowInviteLinkResult, error)
	ListPending(
		ctx context.Context,
		circleID, userID uuid.UUID,
		page, pageSize int32,
	) ([]*entity.CircleJoinRequest, error)
	Approve(ctx context.Context, id, circleID, decidedByUserID uuid.UUID) (*entity.CircleMember, error)
	Reject(ctx context.Context, id, circleID, decidedByUserID uuid.UUID) (*entity.CircleJoinRequest, error)
	Cancel(ctx context.Context, id, userID uuid.UUID) (*entity.CircleJoinRequest, error)
}

type circleJoinRequestUsecase struct {
	repo          CircleJoinRequestRepository
	circles       CircleAccessChecker
	members       CircleMemberLister
	links         CircleInviteLinkResolver
	tokens        RefreshTokenGenerator
	notifications NotificationCreator
}

func NewCircleJoinRequestUsecase(
	repo CircleJoinRequestRepository,
	circles CircleAccessChecker,
	members CircleMemberLister,
	links CircleInviteLinkResolver,
	tokens RefreshTokenGenerator,
	notifications NotificationCreator,
) CircleJoinRequestUsecase {
	return &circleJoinRequestUsecase{
		repo:          repo,
		circles:       circles,
		members:       members,
		links:         links,
		tokens:        tokens,
		notifications: notifications,
	}
}

// FollowInviteLink implements [CircleJoinRequestUsecase].
func (u *circleJoinRequestUsecase) FollowInviteLink(
	ctx context.Context,
	rawToken string,
	userID uuid.UUID,
) (*FollowInviteLinkResult, error) {
	invite, err := u.links.GetActiveLinkByTokenHash(ctx, u.tokens.Hash(rawToken))
	if err != nil {
		return nil, fmt.Errorf("resolving invite link: %w", err)
	}

	requiresApproval := invite.RequiresApproval != nil && *invite.RequiresApproval
	if !requiresApproval {
		members, err := u.repo.AddMembers(ctx, invite.CircleID, []uuid.UUID{userID})
		if err != nil {
			return nil, fmt.Errorf("joining circle via link: %w", err)
		}
		var joined *entity.CircleMember
		if len(members) > 0 {
			joined = members[0]
		} else {
			// Already an active member — AddMembers is idempotent and
			// silently skips them (see its own comment); reflect that
			// back to the caller as "already joined" rather than nil.
			joined, err = u.circles.GetActiveMember(ctx, invite.CircleID, userID)
			if err != nil {
				return nil, fmt.Errorf("confirming existing circle membership: %w", err)
			}
		}
		return &FollowInviteLinkResult{Joined: joined}, nil
	}

	// Approval required: idempotent — following the link twice reports
	// the same outstanding request rather than erroring.
	existing, err := u.repo.GetPendingByUser(ctx, invite.CircleID, userID)
	if err == nil {
		return &FollowInviteLinkResult{JoinRequest: existing}, nil
	}
	if !errors.Is(err, errs.ErrNotFound) {
		return nil, fmt.Errorf("checking existing join request: %w", err)
	}

	joinRequest, err := u.repo.Create(ctx, invite.CircleID, userID, &invite.ID)
	if err != nil {
		return nil, fmt.Errorf("creating circle join request: %w", err)
	}

	members, err := u.members.ListMembers(ctx, invite.CircleID)
	if err != nil {
		return nil, fmt.Errorf("listing circle members to notify: %w", err)
	}
	for _, member := range members {
		if !member.CanInvite {
			continue
		}
		if _, err := u.notifications.CreateNotification(ctx, CreateNotificationInput{
			UserID:              member.UserID,
			Kind:                enum.NotificationKindCircleJoinRequestReceived,
			ActorUserID:         &userID,
			CircleID:            &invite.CircleID,
			CircleJoinRequestID: &joinRequest.ID,
		}); err != nil {
			return nil, fmt.Errorf("notifying invite-privileged member: %w", err)
		}
	}

	return &FollowInviteLinkResult{JoinRequest: joinRequest}, nil
}

// ListPending implements [CircleJoinRequestUsecase]: the approval
// queue, gated to members with invite privileges (FEATURES.md, Circle
// Invite).
func (u *circleJoinRequestUsecase) ListPending(
	ctx context.Context,
	circleID, userID uuid.UUID,
	page, pageSize int32,
) ([]*entity.CircleJoinRequest, error) {
	member, err := u.circles.GetActiveMember(ctx, circleID, userID)
	if err != nil {
		return nil, fmt.Errorf("checking circle membership: %w", err)
	}
	if !member.CanInvite {
		return nil, errs.ErrInsufficientPermission
	}

	page, pageSize = normalizedPage(page, pageSize)
	requests, err := u.repo.ListPendingByCircleID(ctx, circleID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("listing pending circle join requests: %w", err)
	}
	return requests, nil
}

// Approve implements [CircleJoinRequestUsecase]. Deciding the request
// and seating the membership run in the same transaction.
func (u *circleJoinRequestUsecase) Approve(
	ctx context.Context,
	id, circleID, decidedByUserID uuid.UUID,
) (*entity.CircleMember, error) {
	member, err := u.circles.GetActiveMember(ctx, circleID, decidedByUserID)
	if err != nil {
		return nil, fmt.Errorf("checking circle membership: %w", err)
	}
	if !member.CanInvite {
		return nil, errs.ErrInsufficientPermission
	}

	var seated *entity.CircleMember
	var approvedRequest *entity.CircleJoinRequest
	err = u.repo.WithTransaction(ctx, func(tx CircleJoinRequestRepository) error {
		request, err := tx.Approve(ctx, id, circleID, decidedByUserID)
		if err != nil {
			return err
		}
		approvedRequest = request
		members, err := tx.AddMembers(ctx, circleID, []uuid.UUID{request.UserID})
		if err != nil {
			return err
		}
		if len(members) > 0 {
			seated = members[0]
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("approving circle join request: %w", err)
	}

	// Tells the requester both the decision and the resulting membership
	// in one notification — no separate added_to_circle (FEATURES.md
	// never manufactures a second alert for the same outcome).
	if _, err := u.notifications.CreateNotification(ctx, CreateNotificationInput{
		UserID:              approvedRequest.UserID,
		Kind:                enum.NotificationKindCircleJoinRequestApproved,
		ActorUserID:         &decidedByUserID,
		CircleID:            &circleID,
		CircleJoinRequestID: &approvedRequest.ID,
	}); err != nil {
		return nil, fmt.Errorf("notifying approved requester: %w", err)
	}

	return seated, nil
}

// Reject implements [CircleJoinRequestUsecase].
func (u *circleJoinRequestUsecase) Reject(
	ctx context.Context,
	id, circleID, decidedByUserID uuid.UUID,
) (*entity.CircleJoinRequest, error) {
	member, err := u.circles.GetActiveMember(ctx, circleID, decidedByUserID)
	if err != nil {
		return nil, fmt.Errorf("checking circle membership: %w", err)
	}
	if !member.CanInvite {
		return nil, errs.ErrInsufficientPermission
	}

	request, err := u.repo.Reject(ctx, id, circleID, decidedByUserID)
	if err != nil {
		return nil, fmt.Errorf("rejecting circle join request: %w", err)
	}
	return request, nil
}

// Cancel implements [CircleJoinRequestUsecase]: the requester
// withdrawing before anyone answered.
func (u *circleJoinRequestUsecase) Cancel(
	ctx context.Context,
	id, userID uuid.UUID,
) (*entity.CircleJoinRequest, error) {
	request, err := u.repo.Cancel(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("cancelling circle join request: %w", err)
	}
	return request, nil
}
