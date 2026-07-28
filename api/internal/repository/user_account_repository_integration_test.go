//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/repository"
)

func TestUserAccountRepository_CreateCredential_Success(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewUserAccountRepository(pool)

	account, err := repo.CreateCredential(context.Background(), userID, userID.String(), "some-hash")

	require.NoError(t, err)
	require.Equal(t, userID, account.UserID)
	require.NotNil(t, account.PasswordHash)
	require.Equal(t, "some-hash", *account.PasswordHash)
}

func TestUserAccountRepository_GetCredentialByUserID_Found(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewUserAccountRepository(pool)

	_, err := repo.CreateCredential(context.Background(), userID, userID.String(), "some-hash")
	require.NoError(t, err)

	account, err := repo.GetCredentialByUserID(context.Background(), userID)

	require.NoError(t, err)
	require.Equal(t, userID, account.UserID)
}

func TestUserAccountRepository_GetCredentialByUserID_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool) // no credential account created for this user
	repo := repository.NewUserAccountRepository(pool)

	_, err := repo.GetCredentialByUserID(context.Background(), userID)

	require.ErrorIs(t, err, errs.ErrNotFound)
}
