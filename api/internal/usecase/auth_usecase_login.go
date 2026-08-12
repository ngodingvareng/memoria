package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/errs"
)

// Login implements [AuthUsecase].
func (u *authUsecase) Login(ctx context.Context, input LoginInput) (*AuthTokens, error) {
	user, err := u.users.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, errs.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("looking up user: %w", err)
	}

	account, err := u.userAccounts.GetCredentialByUserID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, errs.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("looking up credential account: %w", err)
	}
	if account.PasswordHash == nil {
		return nil, errs.ErrInvalidCredentials
	}

	if account.LockedUntil != nil && account.LockedUntil.After(time.Now()) {
		return nil, errs.ErrAccountLocked
	}

	if err := u.hasher.Compare(*account.PasswordHash, input.Password); err != nil {
		if regErr := u.registerFailedLogin(ctx, user.ID); regErr != nil {
			return nil, fmt.Errorf("registering failed login: %w", regErr)
		}
		return nil, errs.ErrInvalidCredentials
	}

	if account.FailedLoginAttempts > 0 || account.LockedUntil != nil {
		if err := u.userAccounts.ResetFailedLoginAttempts(ctx, user.ID); err != nil {
			return nil, fmt.Errorf("resetting failed login attempts: %w", err)
		}
	}

	// Login always starts a fresh chain.
	return u.issueSession(ctx, u.refreshTokens, user, uuid.New(), input.IPAddress, input.UserAgent)
}

// registerFailedLogin records one more failed password attempt and, once
// the count reaches maxFailedLoginAttempts, locks the account for
// loginLockoutDuration. Locking re-triggers on every failed attempt made
// after the count is already at/above the threshold (including right
// after a previous lockout expires), which is deliberately more
// aggressive than resetting the count on expiry alone.
func (u *authUsecase) registerFailedLogin(ctx context.Context, userID uuid.UUID) error {
	updated, err := u.userAccounts.IncrementFailedLoginAttempts(ctx, userID)
	if err != nil {
		return fmt.Errorf("incrementing failed login attempts: %w", err)
	}
	if int(updated.FailedLoginAttempts) >= u.maxFailedLoginAttempts {
		if err := u.userAccounts.LockCredentialAccount(
			ctx,
			userID,
			time.Now().Add(u.loginLockoutDuration),
		); err != nil {
			return fmt.Errorf("locking account: %w", err)
		}
	}
	return nil
}
