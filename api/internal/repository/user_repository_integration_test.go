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
		Username: strPtr(testUsername()),
		Email:    "unique-create@example.com",
		Timezone: "UTC",
	})

	require.NoError(t, err)
	require.NotEmpty(t, user.ID)
	require.Equal(t, "unique-create@example.com", user.Email)
}

func TestUserRepository_Create_WithoutUsername(t *testing.T) {
	// Registration creates the account before the welcome/onboarding
	// step claims a username — Username must be settable to nil.
	pool := setupTestDB(t)
	repo := repository.NewUserRepository(pool)

	user, err := repo.Create(context.Background(), &entity.User{
		Name:     "No Username Yet",
		Email:    "no-username@example.com",
		Timezone: "UTC",
	})

	require.NoError(t, err)
	require.Nil(t, user.Username)
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	pool := setupTestDB(t)
	repo := repository.NewUserRepository(pool)

	email := "duplicate@example.com"
	_, err := repo.Create(context.Background(), &entity.User{Name: "First", Username: strPtr(testUsername()), Email: email, Timezone: "UTC"})
	require.NoError(t, err)

	_, err = repo.Create(context.Background(), &entity.User{Name: "Second", Username: strPtr(testUsername()), Email: email, Timezone: "UTC"})

	require.ErrorIs(t, err, errs.ErrEmailAlreadyExists)
}

func TestUserRepository_Create_DuplicateEmail_CaseInsensitive(t *testing.T) {
	// Matches uq_users_email_lower — case shouldn't matter for uniqueness.
	pool := setupTestDB(t)
	repo := repository.NewUserRepository(pool)

	_, err := repo.Create(context.Background(), &entity.User{Name: "First", Username: strPtr(testUsername()), Email: "CaseTest@example.com", Timezone: "UTC"})
	require.NoError(t, err)

	_, err = repo.Create(context.Background(), &entity.User{Name: "Second", Username: strPtr(testUsername()), Email: "casetest@example.com", Timezone: "UTC"})

	require.ErrorIs(t, err, errs.ErrEmailAlreadyExists)
}

func TestUserRepository_Create_DuplicateUsername(t *testing.T) {
	// Matches uq_users_username_lower — distinct from the email
	// collision above via pgErr.ConstraintName.
	pool := setupTestDB(t)
	repo := repository.NewUserRepository(pool)

	username := testUsername()
	_, err := repo.Create(context.Background(), &entity.User{Name: "First", Username: strPtr(username), Email: "first-username-dup@example.com", Timezone: "UTC"})
	require.NoError(t, err)

	_, err = repo.Create(context.Background(), &entity.User{Name: "Second", Username: strPtr(username), Email: "second-username-dup@example.com", Timezone: "UTC"})

	require.ErrorIs(t, err, errs.ErrUsernameAlreadyExists)
}

func TestUserRepository_GetByEmail_Found(t *testing.T) {
	pool := setupTestDB(t)
	repo := repository.NewUserRepository(pool)

	created, err := repo.Create(context.Background(), &entity.User{Name: "Findable", Username: strPtr(testUsername()), Email: "findable@example.com", Timezone: "UTC"})
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

	created, err := repo.Create(context.Background(), &entity.User{Name: "ByID", Username: strPtr(testUsername()), Email: "byid@example.com", Timezone: "UTC"})
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

func TestUserRepository_SetUsername_Success(t *testing.T) {
	pool := setupTestDB(t)
	repo := repository.NewUserRepository(pool)

	created, err := repo.Create(context.Background(), &entity.User{Name: "Claims Username", Email: "claims-username@example.com", Timezone: "UTC"})
	require.NoError(t, err)
	require.Nil(t, created.Username)

	username := testUsername()
	updated, err := repo.SetUsername(context.Background(), created.ID, username)

	require.NoError(t, err)
	require.NotNil(t, updated.Username)
	require.Equal(t, username, *updated.Username)
}

func TestUserRepository_SetUsername_DuplicateUsername(t *testing.T) {
	pool := setupTestDB(t)
	repo := repository.NewUserRepository(pool)

	taken := testUsername()
	_, err := repo.Create(context.Background(), &entity.User{Name: "First", Username: strPtr(taken), Email: "set-username-first@example.com", Timezone: "UTC"})
	require.NoError(t, err)

	second, err := repo.Create(context.Background(), &entity.User{Name: "Second", Email: "set-username-second@example.com", Timezone: "UTC"})
	require.NoError(t, err)

	_, err = repo.SetUsername(context.Background(), second.ID, taken)

	require.ErrorIs(t, err, errs.ErrUsernameAlreadyExists)
}

func TestUserRepository_SetUsername_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	repo := repository.NewUserRepository(pool)

	_, err := repo.SetUsername(context.Background(), uuid.New(), testUsername())

	require.ErrorIs(t, err, errs.ErrNotFound)
}
