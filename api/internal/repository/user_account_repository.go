package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

var _ usecase.UserAccountRepository = (*userAccountRepository)(nil)

type userAccountRepository struct {
	q *db.Queries
}

func NewUserAccountRepository(pool *pgxpool.Pool) *userAccountRepository {
	return &userAccountRepository{q: db.New(pool)}
}

// CreateCredential implements [usecase.UserAccountRepository].
func (r *userAccountRepository) CreateCredential(ctx context.Context, userID uuid.UUID, accountID string, passwordHash string) (*entity.UserAccount, error) {
	row, err := r.q.CreateCredentialUserAccount(ctx, db.CreateCredentialUserAccountParams{
		UserID:       userID,
		AccountID:    accountID,
		PasswordHash: pgtype.Text{String: passwordHash, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create credential account: %w", err)
	}
	return toEntityUserAccount(row), nil

}

// GetCredentialByUserID implements [usecase.UserAccountRepository].
func (r *userAccountRepository) GetCredentialByUserID(ctx context.Context, userID uuid.UUID) (*entity.UserAccount, error) {
	row, err := r.q.GetCredentialUserAccountByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("get credential account: %w", err)
	}
	return toEntityUserAccount(row), nil

}

// IncrementFailedLoginAttempts implements [usecase.UserAccountRepository].
func (r *userAccountRepository) IncrementFailedLoginAttempts(ctx context.Context, userID uuid.UUID) (*entity.UserAccount, error) {
	row, err := r.q.IncrementFailedLoginAttempts(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("increment failed login attempts: %w", err)
	}
	return toEntityUserAccount(row), nil
}

// LockCredentialAccount implements [usecase.UserAccountRepository].
func (r *userAccountRepository) LockCredentialAccount(ctx context.Context, userID uuid.UUID, until time.Time) error {
	if err := r.q.LockCredentialUserAccount(ctx, db.LockCredentialUserAccountParams{
		UserID:      userID,
		LockedUntil: timeToPgTimestamptz(until),
	}); err != nil {
		return fmt.Errorf("lock credential account: %w", err)
	}
	return nil
}

// ResetFailedLoginAttempts implements [usecase.UserAccountRepository].
func (r *userAccountRepository) ResetFailedLoginAttempts(ctx context.Context, userID uuid.UUID) error {
	if err := r.q.ResetFailedLoginAttempts(ctx, userID); err != nil {
		return fmt.Errorf("reset failed login attempts: %w", err)
	}
	return nil
}

// UpdatePasswordHash implements [usecase.UserAccountRepository].
func (r *userAccountRepository) UpdatePasswordHash(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	if err := r.q.UpdateUserAccountPasswordHash(ctx, db.UpdateUserAccountPasswordHashParams{
		UserID:       userID,
		PasswordHash: pgtype.Text{String: passwordHash, Valid: true},
	}); err != nil {
		return fmt.Errorf("update password hash: %w", err)
	}
	return nil
}

// GetByProvider implements [usecase.UserAccountRepository].
func (r *userAccountRepository) GetByProvider(ctx context.Context, provider enum.AuthProvider, accountID string) (*entity.UserAccount, error) {
	row, err := r.q.GetUserAccountByProvider(ctx, db.GetUserAccountByProviderParams{
		ProviderID: db.AuthProviderID(provider),
		AccountID:  accountID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("get user account by provider: %w", err)
	}
	return toEntityUserAccount(row), nil
}

// CreateOAuth implements [usecase.UserAccountRepository]. Token/scope
// columns are left unset (Valid: false) — see the interface doc comment.
func (r *userAccountRepository) CreateOAuth(ctx context.Context, userID uuid.UUID, provider enum.AuthProvider, accountID string) (*entity.UserAccount, error) {
	row, err := r.q.CreateOAuthUserAccount(ctx, db.CreateOAuthUserAccountParams{
		UserID:     userID,
		AccountID:  accountID,
		ProviderID: db.AuthProviderID(provider),
	})
	if err != nil {
		return nil, fmt.Errorf("create oauth account: %w", err)
	}
	return toEntityUserAccount(row), nil
}
