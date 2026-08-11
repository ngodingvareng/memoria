package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
)

// ListMyPendingInvites implements [CircleInviteUsecase].
func (u *circleInviteUsecase) ListMyPendingInvites(ctx context.Context, userID uuid.UUID) ([]*entity.CircleInvite, error) {
	invites, err := u.repo.ListPendingByInviteeID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing pending invites: %w", err)
	}
	return invites, nil
}
