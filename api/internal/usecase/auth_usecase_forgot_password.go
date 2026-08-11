package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ngodingvareng/memoria/internal/errs"
)

// ForgotPassword implements [AuthUsecase]. Never reveals whether email
// belongs to an account — the "user not found" path and the "sent"
// path both return nil, on purpose.
func (u *authUsecase) ForgotPassword(ctx context.Context, email string) error {
	user, err := u.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("resolving email: %w", err)
	}

	identifier := passwordResetIdentifier(email)
	// Invalidate any earlier pending reset before issuing a new one —
	// same "resend OTP" reasoning as DeleteUserVerificationsByIdentifier's
	// own doc comment.
	if err := u.verifications.DeleteByIdentifier(ctx, identifier); err != nil {
		return fmt.Errorf("clearing previous reset tokens: %w", err)
	}

	rawToken, err := u.refreshTokenGen.Generate()
	if err != nil {
		return fmt.Errorf("generating reset token: %w", err)
	}
	if _, err := u.verifications.Create(ctx, identifier, u.refreshTokenGen.Hash(rawToken), time.Now().Add(passwordResetTokenTTL)); err != nil {
		return fmt.Errorf("storing reset token: %w", err)
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s&email=%s",
		strings.TrimSuffix(u.webBaseURL, "/"), url.QueryEscape(rawToken), url.QueryEscape(email))
	body := fmt.Sprintf(
		"Hi %s,\n\nSomeone requested a password reset for your Memoria account. If this was you, use the link below within 30 minutes to set a new password:\n\n%s\n\nIf you didn't request this, you can safely ignore this email.",
		user.Name, resetLink,
	)
	if err := u.mailer.Send(ctx, email, "Reset your Memoria password", body); err != nil {
		return fmt.Errorf("sending reset email: %w", err)
	}
	return nil
}
