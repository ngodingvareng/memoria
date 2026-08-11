package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ngodingvareng/memoria/internal/errs"
)

// ResetPassword implements [AuthUsecase].
func (u *authUsecase) ResetPassword(ctx context.Context, input ResetPasswordInput) error {
	identifier := passwordResetIdentifier(input.Email)
	verification, err := u.verifications.GetValid(ctx, identifier, u.refreshTokenGen.Hash(input.Token))
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return errs.ErrInvalidToken
		}
		return fmt.Errorf("validating reset token: %w", err)
	}

	user, err := u.users.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return errs.ErrInvalidToken
		}
		return fmt.Errorf("resolving email: %w", err)
	}

	passwordHash, err := u.hasher.Hash(input.NewPassword)
	if err != nil {
		return fmt.Errorf("hashing new password: %w", err)
	}
	if err := u.userAccounts.UpdatePasswordHash(ctx, user.ID, passwordHash); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	// Best-effort from here: the password change already succeeded,
	// none of this should turn a successful reset into an error
	// response.
	if err := u.verifications.Delete(context.WithoutCancel(ctx), verification.ID); err != nil {
		slog.WarnContext(ctx, "failed to consume password reset token", "error", err)
	}
	if err := u.userAccounts.ResetFailedLoginAttempts(context.WithoutCancel(ctx), user.ID); err != nil {
		slog.WarnContext(ctx, "failed to clear login lockout after password reset", "error", err)
	}
	// Force re-login everywhere — a password reset is exactly the
	// moment a stolen session should stop working.
	if err := u.refreshTokens.RevokeAllByUserID(context.WithoutCancel(ctx), user.ID); err != nil {
		slog.WarnContext(ctx, "failed to revoke sessions after password reset", "error", err)
	}

	return nil
}
