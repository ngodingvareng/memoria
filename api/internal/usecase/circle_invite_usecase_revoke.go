package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
)

// RevokeInvite implements [CircleInviteUsecase]: the inviting side
// withdrawing an invite the recipient hasn't answered yet.
func (u *circleInviteUsecase) RevokeInvite(
	ctx context.Context,
	inviteID, circleID, revokedByUserID uuid.UUID,
) (*entity.CircleInvite, error) {
	if err := u.requireCanInvite(ctx, circleID, revokedByUserID); err != nil {
		return nil, err
	}
	invite, err := u.repo.RevokeUsernameInvite(ctx, inviteID, circleID)
	if err != nil {
		return nil, fmt.Errorf("revoking circle invite: %w", err)
	}
	return invite, nil
}
