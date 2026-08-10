package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
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
		Username: ptrToPgText(user.Username),
		Email:    user.Email,
		Timezone: user.Timezone,
	})
	if err != nil {
		// Catches the race where two concurrent registrations for the
		// same email/username both pass the usecase's own GetByEmail
		// check before either commits — uq_users_email_lower and
		// uq_users_username_lower are the real guards, this just
		// translates whichever one fired into a domain-level error
		// instead of a raw pg error leaking upward.
		if translated, ok := translateUserUniqueViolation(err); ok {
			return nil, translated
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

// GetByUsername implements [usecase.UserRepository].
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	row, err := r.q.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return toEntityUser(row), nil
}

// UpdateImagePath implements [usecase.UserRepository].
func (r *userRepository) UpdateImagePath(ctx context.Context, id uuid.UUID, imagePath *string) (*entity.User, error) {
	row, err := r.q.UpdateUserImagePath(ctx, db.UpdateUserImagePathParams{
		ID:        id,
		ImagePath: ptrToPgText(imagePath),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("update user image path: %w", err)
	}
	return toEntityUser(row), nil
}

// SearchByUsernamePrefix implements [usecase.UserSearcher].
func (r *userRepository) SearchByUsernamePrefix(ctx context.Context, excludeUserID uuid.UUID, query string, limit int32) ([]*entity.User, error) {
	rows, err := r.q.SearchUsersByUsername(ctx, db.SearchUsersByUsernameParams{
		ExcludeUserID: excludeUserID,
		Query:         pgtype.Text{String: escapeLikePattern(query), Valid: true},
		PageLimit:     limit,
	})
	if err != nil {
		return nil, fmt.Errorf("search users by username: %w", err)
	}
	users := make([]*entity.User, len(rows))
	for i, row := range rows {
		users[i] = toEntityUser(row)
	}
	return users, nil
}

// escapeLikePattern escapes ILIKE's own wildcard characters in
// caller-supplied search text, so a query like "50%" or "a_b" is
// matched literally instead of being (mis)interpreted as a pattern.
func escapeLikePattern(s string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(s)
}

// SetUsername implements [usecase.UserRepository].
func (r *userRepository) SetUsername(ctx context.Context, id uuid.UUID, username string) (*entity.User, error) {
	row, err := r.q.SetUsername(ctx, db.SetUsernameParams{
		ID:       id,
		Username: ptrToPgText(&username),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		if translated, ok := translateUserUniqueViolation(err); ok {
			return nil, translated
		}
		return nil, fmt.Errorf("set username: %w", err)
	}
	return toEntityUser(row), nil
}

// translateUserUniqueViolation converts a uq_users_email_lower/
// uq_users_username_lower unique-violation into the matching domain
// error, so callers never see a raw pg error. ok is false for any other
// error, leaving the caller to wrap/return it as-is.
func translateUserUniqueViolation(err error) (error, bool) {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != pgUniqueViolationCode {
		return nil, false
	}
	if pgErr.ConstraintName == "uq_users_username_lower" {
		return errs.ErrUsernameAlreadyExists, true
	}
	return errs.ErrEmailAlreadyExists, true
}
