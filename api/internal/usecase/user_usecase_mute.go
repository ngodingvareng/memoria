package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
)

// MuteUser implements [UserUsecase].
func (u *userUsecase) MuteUser(ctx context.Context, muterUserID uuid.UUID, username string) error {
	target, err := u.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return errs.ErrNotFound
		}
		return fmt.Errorf("resolving username to mute: %w", err)
	}
	if err := u.mutes.Mute(ctx, muterUserID, target.ID); err != nil {
		return fmt.Errorf("muting user: %w", err)
	}
	return nil
}

// UnmuteUser implements [UserUsecase].
func (u *userUsecase) UnmuteUser(ctx context.Context, muterUserID uuid.UUID, username string) error {
	target, err := u.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return errs.ErrNotFound
		}
		return fmt.Errorf("resolving username to unmute: %w", err)
	}
	if err := u.mutes.Unmute(ctx, muterUserID, target.ID); err != nil {
		return fmt.Errorf("unmuting user: %w", err)
	}
	return nil
}

// ListMutedUsers implements [UserUsecase].
func (u *userUsecase) ListMutedUsers(ctx context.Context, muterUserID uuid.UUID) ([]*entity.User, error) {
	ids, err := u.mutes.ListMutedUserIDs(ctx, muterUserID)
	if err != nil {
		return nil, fmt.Errorf("listing muted users: %w", err)
	}
	users, err := u.resolveUsers(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("resolving muted users: %w", err)
	}
	return users, nil
}
