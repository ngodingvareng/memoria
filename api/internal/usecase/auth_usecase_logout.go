package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/ngodingvareng/memoria/internal/errs"
)

// Logout implements [AuthUsecase].
func (u *authUsecase) Logout(ctx context.Context, refreshToken string) error {
	row, err := u.refreshTokens.GetByTokenHash(ctx, u.refreshTokenGen.Hash(refreshToken))
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil // idempotent
		}
		return fmt.Errorf("looking up refresh token: %w", err)
	}
	if row.RevokedAt != nil {
		return nil
	}
	if err := u.refreshTokens.Revoke(ctx, row.ID); err != nil {
		return fmt.Errorf("revoking refresh token: %w", err)
	}
	return nil
}
