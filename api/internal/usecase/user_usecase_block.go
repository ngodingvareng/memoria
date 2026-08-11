package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
)

// BlockUser implements [UserUsecase].
func (u *userUsecase) BlockUser(ctx context.Context, blockerUserID uuid.UUID, username string) error {
	target, err := u.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return errs.ErrNotFound
		}
		return fmt.Errorf("resolving username to block: %w", err)
	}
	if err := u.blocks.Block(ctx, blockerUserID, target.ID); err != nil {
		return fmt.Errorf("blocking user: %w", err)
	}
	return nil
}

// UnblockUser implements [UserUsecase].
func (u *userUsecase) UnblockUser(ctx context.Context, blockerUserID uuid.UUID, username string) error {
	target, err := u.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return errs.ErrNotFound
		}
		return fmt.Errorf("resolving username to unblock: %w", err)
	}
	if err := u.blocks.Unblock(ctx, blockerUserID, target.ID); err != nil {
		return fmt.Errorf("unblocking user: %w", err)
	}
	return nil
}

// ListBlockedUsers implements [UserUsecase].
func (u *userUsecase) ListBlockedUsers(ctx context.Context, blockerUserID uuid.UUID) ([]*entity.User, error) {
	ids, err := u.blocks.ListBlockedUserIDs(ctx, blockerUserID)
	if err != nil {
		return nil, fmt.Errorf("listing blocked users: %w", err)
	}
	users, err := u.resolveUsers(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("resolving blocked users: %w", err)
	}
	return users, nil
}
