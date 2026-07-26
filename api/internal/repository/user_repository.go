package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/usecase"
	"github.com/ngodingvareng/memoria/pkg/errs"
)

const pgUniqueViolationCode = "23505"

var _ usecase.UserRepository = (*userRepository)(nil)

type userRepository struct {
	q *db.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *userRepository {
	return &userRepository{q: db.New(pool)}
}

// Create implements [usecase.UserRepository].
func (r *userRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	row, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Name:     user.Name,
		Email:    user.Email,
		Timezone: user.Timezone,
	})
	if err != nil {
		// Catches the race where two concurrent registrations for the
		// same email both pass the usecase's own GetByEmail check before
		// either commits — uq_users_email_lower is the real guard, this
		// just translates its violation into a domain-level error
		// instead of a raw pg error leaking upward.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode {
			return nil, errs.ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return toEntityUser(row), nil

}

// GetByEmail implements [usecase.UserRepository].
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return toEntityUser(row), nil

}

// GetByID implements [usecase.UserRepository].
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return toEntityUser(row), nil

}
