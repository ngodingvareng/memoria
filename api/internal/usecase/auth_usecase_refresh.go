package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
)

// Refresh implements [AuthUsecase].
func (u *authUsecase) Refresh(ctx context.Context, input RefreshInput) (*AuthTokens, error) {
	tokenHash := u.refreshTokenGen.Hash(input.RefreshToken)

	existing, err := u.refreshTokens.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, errs.ErrUnauthorized
		}
		return nil, fmt.Errorf("looking up refresh token: %w", err)
	}

	// revoked_at already set means this token is no longer the tip
	// of its chain, either because it was already rotated, or
	// because it's being reused. The row alone can't tell those two
	// cases apart, so treat it as potential theft: revoke the whole
	// chain and force a fresh login.
	if existing.RevokedAt != nil {
		if err := u.refreshTokens.RevokeFamily(ctx, existing.FamilyID); err != nil {
			return nil, fmt.Errorf("revoking compromised token family: %w", err)
		}
		return nil, errs.ErrUnauthorized
	}
	if existing.ExpiresAt.Before(time.Now()) {
		return nil, errs.ErrTokenExpired
	}

	user, err := u.users.GetByID(ctx, existing.UserID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, errs.ErrUnauthorized
		}
		return nil, fmt.Errorf("looking up refresh token user: %w", err)
	}

	var tokens *AuthTokens
	err = u.uow.WithTransaction(ctx, func(repos AuthRepositories) error {
		accessToken, accessExpiresAt, err := u.accessTokens.Generate(user.ID)
		if err != nil {
			return fmt.Errorf("generating access token: %w", err)
		}
		rawRefresh, err := u.refreshTokenGen.Generate()
		if err != nil {
			return fmt.Errorf("generating refresh token: %w", err)
		}
		refreshExpiresAt := time.Now().Add(u.refreshTokenTTL)

		// Insert the new row first, then rotate the old row to point at
		// this new one — the order matters because replaced_by_id is a
		// FK into refresh_tokens(id).
		newToken, err := repos.RefreshToken.Create(ctx, &entity.RefreshToken{
			UserID: user.ID, FamilyID: existing.FamilyID,
			TokenHash: u.refreshTokenGen.Hash(rawRefresh),
			ExpiresAt: refreshExpiresAt, IPAddress: input.IPAddress, UserAgent: input.UserAgent,
		})
		if err != nil {
			return fmt.Errorf("creating rotated refresh token: %w", err)
		}

		if err := repos.RefreshToken.Rotate(ctx, existing.ID, newToken.ID); err != nil {
			// errs.ErrConflict here means another (concurrent) Refresh
			// request already rotated this token first, we lost the
			// race, and the whole transaction (including newToken
			// above) gets rolled back.
			return fmt.Errorf("rotating refresh token: %w", err)
		}

		tokens = &AuthTokens{
			User: user, AccessToken: accessToken, AccessTokenExpiresAt: accessExpiresAt,
			RefreshToken: rawRefresh, RefreshTokenExpiresAt: refreshExpiresAt,
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, errs.ErrConflict) {
			return nil, errs.ErrUnauthorized
		}
		return nil, fmt.Errorf("refreshing session: %w", err)
	}
	return tokens, nil
}
