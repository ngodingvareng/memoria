package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/usecase"
	"github.com/ngodingvareng/memoria/pkg/errs"
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
