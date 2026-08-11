package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
	"github.com/ngodingvareng/memoria/internal/errs"
)

// LoginWithGoogle implements [AuthUsecase].
func (u *authUsecase) LoginWithGoogle(ctx context.Context, input GoogleLoginInput) (*AuthTokens, error) {
	identity, err := u.googleVerifier.Verify(ctx, input.IDToken)
	if err != nil {
		return nil, err
	}

	var tokens *AuthTokens
	err = u.uow.WithTransaction(ctx, func(repos AuthRepositories) error {
		account, err := repos.UserAccount.GetByProvider(ctx, enum.AuthProviderGoogle, identity.Subject)
		if err == nil {
			user, err := repos.User.GetByID(ctx, account.UserID)
			if err != nil {
				return fmt.Errorf("looking up google-linked user: %w", err)
			}
			t, err := u.issueSession(ctx, repos.RefreshToken, user, uuid.New(), input.IPAddress, input.UserAgent)
			if err != nil {
				return fmt.Errorf("issuing session: %w", err)
			}
			tokens = t
			return nil
		}
		if !errors.Is(err, errs.ErrNotFound) {
			return fmt.Errorf("looking up google account: %w", err)
		}

		// No user_accounts row links this Google identity yet. If the
		// email is already registered some other way, reject rather
		// than silently taking over that account — see LoginWithGoogle's
		// doc comment.
		existingUser, err := repos.User.GetByEmail(ctx, identity.Email)
		if err == nil && existingUser != nil {
			return errs.ErrEmailAlreadyExists
		}
		if err != nil && !errors.Is(err, errs.ErrNotFound) {
			return fmt.Errorf("checking existing email: %w", err)
		}

		user, err := repos.User.Create(ctx, &entity.User{
			Name: identity.Name, Email: identity.Email,
			EmailVerified: identity.EmailVerified, Timezone: "UTC",
		})
		if err != nil {
			return fmt.Errorf("creating user: %w", err)
		}
		if _, err := repos.UserAccount.CreateOAuth(ctx, user.ID, enum.AuthProviderGoogle, identity.Subject); err != nil {
			return fmt.Errorf("creating google account: %w", err)
		}
		t, err := u.issueSession(ctx, repos.RefreshToken, user, uuid.New(), input.IPAddress, input.UserAgent)
		if err != nil {
			return fmt.Errorf("issuing session: %w", err)
		}
		tokens = t
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("google login: %w", err)
	}
	return tokens, nil
}
