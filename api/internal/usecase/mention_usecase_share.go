package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
)

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
