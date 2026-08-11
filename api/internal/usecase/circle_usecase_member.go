package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
	"github.com/ngodingvareng/memoria/internal/errs"
)

// LeaveCircle implements [CircleUsecase]: self-service departure. Blocks
// the Circle's sole active admin from leaving — otherwise the Circle
// would be left with no admin able to manage it, invites, or join
// requests ever again.
func (u *circleUsecase) LeaveCircle(ctx context.Context, circleID, userID uuid.UUID) error {
	isSoleAdmin, err := u.repo.IsSoleActiveAdmin(ctx, circleID, userID)
	if err != nil {
		return fmt.Errorf("checking sole admin status: %w", err)
	}
	if isSoleAdmin {
		return errs.ErrLastCircleAdmin
	}

	if err := u.repo.Leave(ctx, circleID, userID); err != nil {
		return fmt.Errorf("leaving circle: %w", err)
	}
	return nil
}

// RemoveMember implements [CircleUsecase]: admin-forced removal.
func (u *circleUsecase) RemoveMember(ctx context.Context, circleID, targetUserID, removedByUserID uuid.UUID) error {
	if err := u.repo.RemoveMember(ctx, circleID, targetUserID, removedByUserID); err != nil {
		return fmt.Errorf("removing circle member: %w", err)
	}
	return nil
}

// UpdateMemberRole implements [CircleUsecase]. Blocks demoting the
// Circle's sole active admin to member — same orphaned-Circle concern as
// LeaveCircle. Promoting someone to admin never needs the check.
func (u *circleUsecase) UpdateMemberRole(ctx context.Context, input UpdateCircleMemberRoleInput) (*entity.CircleMember, error) {
	if input.Role != enum.CircleRoleAdmin {
		isSoleAdmin, err := u.repo.IsSoleActiveAdmin(ctx, input.CircleID, input.UserID)
		if err != nil {
			return nil, fmt.Errorf("checking sole admin status: %w", err)
		}
		if isSoleAdmin {
			return nil, errs.ErrLastCircleAdmin
		}
	}

	member, err := u.repo.UpdateMemberRole(ctx, input.CircleID, input.UserID, input.ChangedByUserID, input.Role)
	if err != nil {
		return nil, fmt.Errorf("updating circle member role: %w", err)
	}
	return member, nil
}

// UpdateMemberPermissions implements [CircleUsecase].
func (u *circleUsecase) UpdateMemberPermissions(ctx context.Context, input UpdateCircleMemberPermissionsInput) (*entity.CircleMember, error) {
	member, err := u.repo.UpdateMemberPermissions(ctx, input.CircleID, input.UserID, input.ChangedByUserID, input.CanInvite, input.CanCapture)
	if err != nil {
		return nil, fmt.Errorf("updating circle member permissions: %w", err)
	}
	return member, nil
}
