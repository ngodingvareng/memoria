package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/enum"
	"github.com/ngodingvareng/memoria/internal/errs"
)

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

	// Direct adds had no decision point of their own — too consequential
	// to leave unsaid even though nothing was asked (FEATURES.md, Circle
	// Invite). Pending invites are the "someone needs you" case.
	// Accepting/declining an invite the recipient already saw doesn't
	// get a further notification — they already know.
	for _, member := range result.AddedMembers {
		if _, err := u.notifications.CreateNotification(ctx, CreateNotificationInput{
			UserID:      member.UserID,
			Kind:        enum.NotificationKindAddedToCircle,
			ActorUserID: &input.InvitedByUserID,
			CircleID:    &input.CircleID,
		}); err != nil {
			return nil, fmt.Errorf("notifying added circle member: %w", err)
		}
	}
	for _, invite := range result.PendingInvites {
		if _, err := u.notifications.CreateNotification(ctx, CreateNotificationInput{
			UserID:         *invite.InviteeUserID,
			Kind:           enum.NotificationKindCircleInviteReceived,
			ActorUserID:    &input.InvitedByUserID,
			CircleID:       &input.CircleID,
			CircleInviteID: &invite.ID,
		}); err != nil {
			return nil, fmt.Errorf("notifying invited user: %w", err)
		}
	}

	return result, nil
}
