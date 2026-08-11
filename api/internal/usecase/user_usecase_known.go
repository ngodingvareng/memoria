package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
)

// MarkUserKnown implements [UserUsecase].
func (u *userUsecase) MarkUserKnown(ctx context.Context, knowerUserID uuid.UUID, username string) error {
	target, err := u.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return errs.ErrNotFound
		}
		return fmt.Errorf("resolving username: %w", err)
	}
	if err := u.knowns.MarkKnown(ctx, knowerUserID, target.ID); err != nil {
		return fmt.Errorf("marking user known: %w", err)
	}
	return nil
}

// UnmarkUserKnown implements [UserUsecase].
func (u *userUsecase) UnmarkUserKnown(ctx context.Context, knowerUserID uuid.UUID, username string) error {
	target, err := u.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return errs.ErrNotFound
		}
		return fmt.Errorf("resolving username to unmark: %w", err)
	}
	if err := u.knowns.Unmark(ctx, knowerUserID, target.ID); err != nil {
		return fmt.Errorf("unmarking user known: %w", err)
	}
	return nil
}

// ListKnownUsers implements [UserUsecase].
func (u *userUsecase) ListKnownUsers(ctx context.Context, knowerUserID uuid.UUID) ([]*entity.User, error) {
	ids, err := u.knowns.ListKnownUserIDs(ctx, knowerUserID)
	if err != nil {
		return nil, fmt.Errorf("listing known users: %w", err)
	}
	users, err := u.resolveUsers(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("resolving known users: %w", err)
	}
	return users, nil
}
