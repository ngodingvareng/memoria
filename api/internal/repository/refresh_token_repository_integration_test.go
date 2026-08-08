//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/repository"
)

// mustCreateTestUser inserts a bare-minimum users row directly via SQL —
// refresh_tokens.user_id has a FK to users(id), so every test below
// needs a real user row to attach its tokens to.
func mustCreateTestUser(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := pool.QueryRow(context.Background(),
		`INSERT INTO users (name, username, email, timezone) VALUES ($1, $2, $3, 'UTC') RETURNING id`,
		"Refresh Token Test User", testUsername(), uuid.NewString()+"@example.com",
	).Scan(&id)
	require.NoError(t, err)
	return id
}

func TestRefreshTokenRepository_CreateAndGetByTokenHash(t *testing.T) {
	pool := setupTestDB(t)
	repo := repository.NewRefreshTokenRepository(pool)
	userID := mustCreateTestUser(t, pool)
	familyID := uuid.New()

	created, err := repo.Create(context.Background(), &entity.RefreshToken{
		UserID:    userID,
		FamilyID:  familyID,
		TokenHash: "hash-1",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, created.ID)

	fetched, err := repo.GetByTokenHash(context.Background(), "hash-1")
	require.NoError(t, err)
	require.Equal(t, created.ID, fetched.ID)
	require.Equal(t, userID, fetched.UserID)
	require.Equal(t, familyID, fetched.FamilyID)
	require.Nil(t, fetched.RevokedAt)
	require.Nil(t, fetched.ReplacedByID)
}

func TestRefreshTokenRepository_GetByTokenHash_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	repo := repository.NewRefreshTokenRepository(pool)

	_, err := repo.GetByTokenHash(context.Background(), "does-not-exist")

	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestRefreshTokenRepository_Revoke(t *testing.T) {
	pool := setupTestDB(t)
	repo := repository.NewRefreshTokenRepository(pool)
	userID := mustCreateTestUser(t, pool)

	created, err := repo.Create(context.Background(), &entity.RefreshToken{
		UserID: userID, FamilyID: uuid.New(), TokenHash: "hash-revoke",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	require.NoError(t, repo.Revoke(context.Background(), created.ID))

	fetched, err := repo.GetByTokenHash(context.Background(), "hash-revoke")
	require.NoError(t, err)
	require.NotNil(t, fetched.RevokedAt)
}

func TestRefreshTokenRepository_Rotate_Success(t *testing.T) {
	pool := setupTestDB(t)
	repo := repository.NewRefreshTokenRepository(pool)
	userID := mustCreateTestUser(t, pool)
	familyID := uuid.New()

	oldToken, err := repo.Create(context.Background(), &entity.RefreshToken{
		UserID: userID, FamilyID: familyID, TokenHash: "hash-old",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	newToken, err := repo.Create(context.Background(), &entity.RefreshToken{
		UserID: userID, FamilyID: familyID, TokenHash: "hash-new",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	require.NoError(t, repo.Rotate(context.Background(), oldToken.ID, newToken.ID))

	fetchedOld, err := repo.GetByTokenHash(context.Background(), "hash-old")
	require.NoError(t, err)
	require.NotNil(t, fetchedOld.RevokedAt)
	require.NotNil(t, fetchedOld.ReplacedByID)
	require.Equal(t, newToken.ID, *fetchedOld.ReplacedByID)
}

func TestRefreshTokenRepository_Rotate_ConflictWhenAlreadyRotated(t *testing.T) {
	pool := setupTestDB(t)
	repo := repository.NewRefreshTokenRepository(pool)
	userID := mustCreateTestUser(t, pool)
	familyID := uuid.New()

	oldToken, err := repo.Create(context.Background(), &entity.RefreshToken{
		UserID: userID, FamilyID: familyID, TokenHash: "hash-conflict-old",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	firstReplacement, err := repo.Create(context.Background(), &entity.RefreshToken{
		UserID: userID, FamilyID: familyID, TokenHash: "hash-conflict-first",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	// First rotation wins normally.
	require.NoError(t, repo.Rotate(context.Background(), oldToken.ID, firstReplacement.ID))

	secondReplacement, err := repo.Create(context.Background(), &entity.RefreshToken{
		UserID: userID, FamilyID: familyID, TokenHash: "hash-conflict-second",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	// Simulates a concurrent request that read oldToken before the first
	// rotation committed, then tried to rotate it too — this is exactly
	// the race usecase.authUsecase.Refresh's errs.ErrConflict handling
	// exists for.
	err = repo.Rotate(context.Background(), oldToken.ID, secondReplacement.ID)

	require.ErrorIs(t, err, errs.ErrConflict)
}

func TestRefreshTokenRepository_RevokeFamily(t *testing.T) {
	pool := setupTestDB(t)
	repo := repository.NewRefreshTokenRepository(pool)
	userID := mustCreateTestUser(t, pool)
	familyID := uuid.New()

	_, err := repo.Create(context.Background(), &entity.RefreshToken{
		UserID: userID, FamilyID: familyID, TokenHash: "hash-family-a",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)
	_, err = repo.Create(context.Background(), &entity.RefreshToken{
		UserID: userID, FamilyID: familyID, TokenHash: "hash-family-b",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	require.NoError(t, repo.RevokeFamily(context.Background(), familyID))

	fetchedA, err := repo.GetByTokenHash(context.Background(), "hash-family-a")
	require.NoError(t, err)
	require.NotNil(t, fetchedA.RevokedAt)

	fetchedB, err := repo.GetByTokenHash(context.Background(), "hash-family-b")
	require.NoError(t, err)
	require.NotNil(t, fetchedB.RevokedAt)
}

func TestRefreshTokenRepository_RevokeAllByUserID(t *testing.T) {
	pool := setupTestDB(t)
	repo := repository.NewRefreshTokenRepository(pool)
	userID := mustCreateTestUser(t, pool)

	_, err := repo.Create(context.Background(), &entity.RefreshToken{
		UserID: userID, FamilyID: uuid.New(), TokenHash: "hash-user-a",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)
	_, err = repo.Create(context.Background(), &entity.RefreshToken{
		UserID: userID, FamilyID: uuid.New(), TokenHash: "hash-user-b",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	require.NoError(t, repo.RevokeAllByUserID(context.Background(), userID))

	fetchedA, err := repo.GetByTokenHash(context.Background(), "hash-user-a")
	require.NoError(t, err)
	require.NotNil(t, fetchedA.RevokedAt)

	fetchedB, err := repo.GetByTokenHash(context.Background(), "hash-user-b")
	require.NoError(t, err)
	require.NotNil(t, fetchedB.RevokedAt)
}
