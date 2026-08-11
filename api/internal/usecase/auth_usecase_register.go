package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
)

// Register implements [AuthUsecase].
func (u *authUsecase) Register(ctx context.Context, input RegisterInput) (*AuthTokens, error) {
	existing, err := u.users.GetByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		return nil, fmt.Errorf("checking existing email: %w", err)
	}
	if existing != nil {
		return nil, errs.ErrEmailAlreadyExists
	}

	hashedPassword, err := u.hasher.Hash(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	var tokens *AuthTokens
	err = u.uow.WithTransaction(ctx, func(repos AuthRepositories) error {
		user, err := repos.User.Create(ctx, &entity.User{
			Name: input.Name, Email: input.Email, Timezone: "UTC",
		})
		if err != nil {
			return fmt.Errorf("creating user: %w", err)
		}
		if _, err := repos.UserAccount.CreateCredential(ctx, user.ID, user.ID.String(), hashedPassword); err != nil {
			return fmt.Errorf("creating credential account: %w", err)
		}
		// Register doubles as login — the account is unusable for the
		// welcome/username-claim step otherwise, and there's no reason
		// to make the frontend log in a second time right after signup.
		t, err := u.issueSession(ctx, repos.RefreshToken, user, uuid.New(), input.IPAddress, input.UserAgent)
		if err != nil {
			return fmt.Errorf("issuing session: %w", err)
		}
		tokens = t
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("registering user: %w", err)
	}
	return tokens, nil
}
