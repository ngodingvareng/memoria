package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

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
