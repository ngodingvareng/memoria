package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
)

// DeclineInvite implements [CircleInviteUsecase].
func (u *circleInviteUsecase) DeclineInvite(ctx context.Context, inviteID, userID uuid.UUID) (*entity.CircleInvite, error) {
	invite, err := u.repo.DeclineUsernameInvite(ctx, inviteID, userID)
	if err != nil {
		return nil, fmt.Errorf("declining circle invite: %w", err)
	}
	return invite, nil
}
