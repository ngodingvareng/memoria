//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/repository"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

func TestAuthUnitOfWork_WithTransaction_CommitsBothTablesTogether(t *testing.T) {
	pool := setupTestDB(t)
	uow := repository.NewAuthUnitOfWork(pool)

	email := "uow-success@example.com"
	var createdUserID uuid.UUID

	err := uow.WithTransaction(context.Background(), func(repos usecase.AuthRepositories) error {
		user, err := repos.User.Create(context.Background(), &entity.User{
			Name:     "UoW Test",
			Username: testUsername(),
			Email:    email,
			Timezone: "UTC",
		})
		if err != nil {
			return err
		}
		createdUserID = user.ID

		_, err = repos.UserAccount.CreateCredential(context.Background(), user.ID, user.ID.String(), "some-hash")
		return err
	})

	require.NoError(t, err)

	// Verify BOTH rows are visible outside the transaction — i.e. really
	// committed, not just returned in-memory.
	var userCount, accountCount int
	err = pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM users WHERE id = $1`, createdUserID).Scan(&userCount)
	require.NoError(t, err)
	require.Equal(t, 1, userCount)

	err = pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM user_accounts WHERE user_id = $1`, createdUserID).Scan(&accountCount)
	require.NoError(t, err)
	require.Equal(t, 1, accountCount)
}

func TestAuthUnitOfWork_WithTransaction_RollsBackBothTablesOnFailure(t *testing.T) {
	pool := setupTestDB(t)
	uow := repository.NewAuthUnitOfWork(pool)

	email := "uow-rollback@example.com"
	sentinelErr := errors.New("forced failure after user create")

	err := uow.WithTransaction(context.Background(), func(repos usecase.AuthRepositories) error {
		_, err := repos.User.Create(context.Background(), &entity.User{
			Name:     "UoW Rollback Test",
			Username: testUsername(),
			Email:    email,
			Timezone: "UTC",
		})
		if err != nil {
			return err
		}
		// Force a failure AFTER the user insert succeeded — this is
		// exactly the scenario the whole Unit of Work exists to guard
		// against: an orphaned users row with no matching
		// user_accounts row, which would leave a registered user who
		// can never log in.
		return sentinelErr
	})

	require.ErrorIs(t, err, sentinelErr)

	var userCount int
	err = pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM users WHERE email = $1`, email).Scan(&userCount)
	require.NoError(t, err)
	require.Equal(t, 0, userCount, "the user insert should have been rolled back, not left as an orphan")
}
