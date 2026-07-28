//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/repository"
)

func TestUserRepository_Create_Success(t *testing.T) {
	pool := setupTestDB(t)
	repo := repository.NewUserRepository(pool)

	user, err := repo.Create(context.Background(), &entity.User{
		Name:     "Test User",
		Email:    "unique-create@example.com",
		Timezone: "UTC",
	})

	require.NoError(t, err)
	require.NotEmpty(t, user.ID)
	require.Equal(t, "unique-create@example.com", user.Email)
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	pool := setupTestDB(t)
	repo := repository.NewUserRepository(pool)

	email := "duplicate@example.com"
	_, err := repo.Create(context.Background(), &entity.User{Name: "First", Email: email, Timezone: "UTC"})
	require.NoError(t, err)

	_, err = repo.Create(context.Background(), &entity.User{Name: "Second", Email: email, Timezone: "UTC"})

	require.ErrorIs(t, err, errs.ErrEmailAlreadyExists)
}

func TestUserRepository_Create_DuplicateEmail_CaseInsensitive(t *testing.T) {
	// Matches uq_users_email_lower — case shouldn't matter for uniqueness.
	pool := setupTestDB(t)
	repo := repository.NewUserRepository(pool)

	_, err := repo.Create(context.Background(), &entity.User{Name: "First", Email: "CaseTest@example.com", Timezone: "UTC"})
	require.NoError(t, err)

	_, err = repo.Create(context.Background(), &entity.User{Name: "Second", Email: "casetest@example.com", Timezone: "UTC"})

	require.ErrorIs(t, err, errs.ErrEmailAlreadyExists)
}

func TestUserRepository_GetByEmail_Found(t *testing.T) {
	pool := setupTestDB(t)
	repo := repository.NewUserRepository(pool)

	created, err := repo.Create(context.Background(), &entity.User{Name: "Findable", Email: "findable@example.com", Timezone: "UTC"})
	require.NoError(t, err)

	found, err := repo.GetByEmail(context.Background(), "findable@example.com")

	require.NoError(t, err)
	require.Equal(t, created.ID, found.ID)
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	repo := repository.NewUserRepository(pool)

	_, err := repo.GetByEmail(context.Background(), "nobody@example.com")

	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestUserRepository_GetByID_Found(t *testing.T) {
	pool := setupTestDB(t)
	repo := repository.NewUserRepository(pool)

	created, err := repo.Create(context.Background(), &entity.User{Name: "ByID", Email: "byid@example.com", Timezone: "UTC"})
	require.NoError(t, err)

	found, err := repo.GetByID(context.Background(), created.ID)

	require.NoError(t, err)
	require.Equal(t, created.Email, found.Email)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	repo := repository.NewUserRepository(pool)

	_, err := repo.GetByID(context.Background(), uuid.New())

	require.ErrorIs(t, err, errs.ErrNotFound)
}
