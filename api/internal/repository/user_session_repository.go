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

var _ usecase.UserSessionRepository = (*userSessionRepository)(nil)

type userSessionRepository struct {
	q *db.Queries
}

func NewUserSessionRepository(pool *pgxpool.Pool) *userSessionRepository {
	return &userSessionRepository{q: db.New(pool)}
}

// Create implements [usecase.UserSessionRepository].
func (r *userSessionRepository) Create(ctx context.Context, session *entity.UserSession) (*entity.UserSession, error) {
	row, err := r.q.CreateUserSession(ctx, db.CreateUserSessionParams{
		UserID:    session.UserID,
		TokenHash: session.TokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: session.ExpiresAt, Valid: true},
		IpAddress: ptrToPgText(session.IPAddress),
		UserAgent: ptrToPgText(session.UserAgent),
	})
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return toEntitySession(row), nil
}

// GetByTokenHash implements [usecase.UserSessionRepository].
func (r *userSessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*entity.UserSession, error) {
	row, err := r.q.GetUserSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("get session by token hash: %w", err)
	}
	return toEntitySession(row), nil
}

// DeleteByTokenHash implements [usecase.UserSessionRepository].
func (r *userSessionRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	if err := r.q.DeleteUserSessionByTokenHash(ctx, tokenHash); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// DeleteAllByUserID implements [usecase.UserSessionRepository].
func (r *userSessionRepository) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	if err := r.q.DeleteAllUserSessionsByUserID(ctx, userID); err != nil {
		return fmt.Errorf("delete all sessions: %w", err)
	}
	return nil
}
