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
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

var _ usecase.UserVerificationRepository = (*userVerificationRepository)(nil)

type userVerificationRepository struct {
	q *db.Queries
}

func NewUserVerificationRepository(pool *pgxpool.Pool) *userVerificationRepository {
	return &userVerificationRepository{q: db.New(pool)}
}

// Create implements [usecase.UserVerificationRepository].
func (r *userVerificationRepository) Create(ctx context.Context, identifier, value string, expiresAt time.Time) (*entity.UserVerification, error) {
	row, err := r.q.CreateUserVerification(ctx, db.CreateUserVerificationParams{
		Identifier: identifier,
		Value:      value,
		ExpiresAt:  pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create user verification: %w", err)
	}
	return toEntityUserVerification(row), nil
}

// GetValid implements [usecase.UserVerificationRepository].
func (r *userVerificationRepository) GetValid(ctx context.Context, identifier, value string) (*entity.UserVerification, error) {
	row, err := r.q.GetValidUserVerification(ctx, db.GetValidUserVerificationParams{
		Identifier: identifier,
		Value:      value,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("get valid user verification: %w", err)
	}
	return toEntityUserVerification(row), nil
}

// Delete implements [usecase.UserVerificationRepository]. A sqlc :exec
// query — a WHERE clause matching zero rows (already deleted) is a
// silent no-op here.
func (r *userVerificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.q.DeleteUserVerification(ctx, id); err != nil {
		return fmt.Errorf("delete user verification: %w", err)
	}
	return nil
}

// DeleteByIdentifier implements [usecase.UserVerificationRepository].
func (r *userVerificationRepository) DeleteByIdentifier(ctx context.Context, identifier string) error {
	if err := r.q.DeleteUserVerificationsByIdentifier(ctx, identifier); err != nil {
		return fmt.Errorf("delete user verifications by identifier: %w", err)
	}
	return nil
}
