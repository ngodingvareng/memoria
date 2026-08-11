package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
)

// ListMyCircles implements [CircleUsecase].
func (u *circleUsecase) ListMyCircles(ctx context.Context, userID uuid.UUID) ([]*entity.Circle, error) {
	circles, err := u.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing circles: %w", err)
	}
	return circles, nil
}

// ListMembers implements [CircleUsecase]. Requires the caller to
// already be an active member — same access rule as GetCircle.
func (u *circleUsecase) ListMembers(ctx context.Context, circleID, userID uuid.UUID) ([]*entity.CircleMember, error) {
	if _, err := u.repo.GetByID(ctx, circleID, userID); err != nil {
		return nil, fmt.Errorf("checking circle access: %w", err)
	}
	members, err := u.repo.ListMembers(ctx, circleID)
	if err != nil {
		return nil, fmt.Errorf("listing circle members: %w", err)
	}
	return members, nil
}
