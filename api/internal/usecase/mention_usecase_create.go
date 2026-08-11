package usecase

import (
	"context"
	"fmt"

	"github.com/ngodingvareng/memoria/internal/enum"
	"github.com/ngodingvareng/memoria/internal/errs"
)

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

	if _, err := u.notifications.CreateNotification(ctx, CreateNotificationInput{
		UserID:      target.ID,
		Kind:        enum.NotificationKindMentionedInMoment,
		ActorUserID: &input.OwnerUserID,
		MomentID:    &input.MomentID,
	}); err != nil {
		return nil, fmt.Errorf("notifying mentioned user: %w", err)
	}

	sharedCircleIDs, err := u.circleShares.ListSharedCircleIDs(ctx, input.OwnerUserID, target.ID)
	if err != nil {
		return nil, fmt.Errorf("listing shared circles: %w", err)
	}

	return &CreateMentionResult{Mention: mention, SharedCircleIDs: sharedCircleIDs}, nil
}
