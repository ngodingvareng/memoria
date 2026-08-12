package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
)

// AcceptInvite implements [CircleInviteUsecase]. Accepting the invite
// row and seating the membership run in the same transaction — an
// accepted invite with no resulting membership would be a stuck state
// no retry can recover from.
func (u *circleInviteUsecase) AcceptInvite(
	ctx context.Context,
	inviteID, userID uuid.UUID,
) (*entity.CircleMember, error) {
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
