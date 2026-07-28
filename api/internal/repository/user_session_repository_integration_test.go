//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/repository"
)

func TestSessionRepository_Create_Success(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewUserSessionRepository(pool)

	ip := "127.0.0.1"
	ua := "test-agent/1.0"
	session, err := repo.Create(context.Background(), &entity.UserSession{
		UserID:    userID,
		TokenHash: "some-token-hash",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IPAddress: &ip,
		UserAgent: &ua,
	})

	require.NoError(t, err)
	require.Equal(t, userID, session.UserID)
	require.NotNil(t, session.IPAddress)
	require.Equal(t, "127.0.0.1", *session.IPAddress)
}

func TestSessionRepository_GetByTokenHash_Found(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewUserSessionRepository(pool)

	created, err := repo.Create(context.Background(), &entity.UserSession{
		UserID:    userID,
		TokenHash: "findable-hash",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	found, err := repo.GetByTokenHash(context.Background(), "findable-hash")

	require.NoError(t, err)
	require.Equal(t, created.ID, found.ID)
}

func TestSessionRepository_GetByTokenHash_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	repo := repository.NewUserSessionRepository(pool)

	_, err := repo.GetByTokenHash(context.Background(), "nonexistent-hash")

	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestSessionRepository_DeleteByTokenHash(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewUserSessionRepository(pool)

	_, err := repo.Create(context.Background(), &entity.UserSession{
		UserID: userID, TokenHash: "to-delete", ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	err = repo.DeleteByTokenHash(context.Background(), "to-delete")
	require.NoError(t, err)

	_, err = repo.GetByTokenHash(context.Background(), "to-delete")
	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestSessionRepository_DeleteAllByUserID(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	otherUserID := seedTestUser(t, pool)
	repo := repository.NewUserSessionRepository(pool)

	_, err := repo.Create(context.Background(), &entity.UserSession{UserID: userID, TokenHash: "session-1", ExpiresAt: time.Now().Add(24 * time.Hour)})
	require.NoError(t, err)
	_, err = repo.Create(context.Background(), &entity.UserSession{UserID: userID, TokenHash: "session-2", ExpiresAt: time.Now().Add(24 * time.Hour)})
	require.NoError(t, err)
	_, err = repo.Create(context.Background(), &entity.UserSession{UserID: otherUserID, TokenHash: "session-other", ExpiresAt: time.Now().Add(24 * time.Hour)})
	require.NoError(t, err)

	err = repo.DeleteAllByUserID(context.Background(), userID)
	require.NoError(t, err)

	_, err = repo.GetByTokenHash(context.Background(), "session-1")
	require.ErrorIs(t, err, errs.ErrNotFound)
	_, err = repo.GetByTokenHash(context.Background(), "session-2")
	require.ErrorIs(t, err, errs.ErrNotFound)

	// The other user's session must survive — DeleteAllByUserID should
	// never touch sessions belonging to a different user.
	_, err = repo.GetByTokenHash(context.Background(), "session-other")
	require.NoError(t, err)
}
