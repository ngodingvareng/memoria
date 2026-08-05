package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

var _ usecase.RefreshTokenRepository = (*refreshTokenRepository)(nil)

type refreshTokenRepository struct {
	q *db.Queries
}

// NewRefreshTokenRepository takes a db.DBTX (not *pgxpool.Pool directly)
// so it can be used both outside a transaction (pool) and bound to a
// transaction (pgx.Tx) by AuthUnitOfWork.
func NewRefreshTokenRepository(dbtx db.DBTX) *refreshTokenRepository {
	return &refreshTokenRepository{q: db.New(dbtx)}
}

func (r *refreshTokenRepository) Create(ctx context.Context, rt *entity.RefreshToken) (*entity.RefreshToken, error) {
	row, err := r.q.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    rt.UserID,
		FamilyID:  rt.FamilyID,
		TokenHash: rt.TokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: rt.ExpiresAt, Valid: true},
		IpAddress: ptrToPgText(rt.IPAddress),
		UserAgent: ptrToPgText(rt.UserAgent),
	})
	if err != nil {
		return nil, fmt.Errorf("create refresh token: %w", err)
	}
	return toEntityRefreshToken(row), nil
}

func (r *refreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	row, err := r.q.GetRefreshTokenByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("get refresh token by hash: %w", err)
	}
	return toEntityRefreshToken(row), nil
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	if err := r.q.RevokeRefreshToken(ctx, id); err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

func (r *refreshTokenRepository) Rotate(ctx context.Context, id, replacedByID uuid.UUID) error {
	affected, err := r.q.RotateRefreshToken(ctx, db.RotateRefreshTokenParams{
		ID:           id,
		ReplacedByID: pgtype.UUID{Bytes: replacedByID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("rotate refresh token: %w", err)
	}
	if affected == 0 {
		return errs.ErrConflict
	}
	return nil
}

func (r *refreshTokenRepository) RevokeFamily(ctx context.Context, familyID uuid.UUID) error {
	if err := r.q.RevokeRefreshTokenFamily(ctx, familyID); err != nil {
		return fmt.Errorf("revoke refresh token family: %w", err)
	}
	return nil
}

func (r *refreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	if err := r.q.RevokeAllRefreshTokensByUserID(ctx, userID); err != nil {
		return fmt.Errorf("revoke all refresh tokens: %w", err)
	}
	return nil
}
