//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-openapi/testify/v2/require"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/repository"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

func TestThreadRepository_WithTransaction_CommitsOnSuccess(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewThreadRepository(pool)

	var created *entity.Thread
	err := repo.WithTransaction(context.Background(), func(tx usecase.ThreadRepository) error {
		var err error
		created, err = tx.Create(context.Background(), &entity.Thread{
			UserID: userID,
			Name:   "Morning cook",
		})
		return err
	})
	require.NoError(t, err)
	require.NotNil(t, created)

	// Confirm it's actually visible outside the transaction that created
	// it — i.e. it was really committed, not just returned in-memory.
	var count int
	err = pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM threads WHERE id = $1`, created.ID,
	).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestThreadRepository_WithTransaction_RollsBackOnError(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewThreadRepository(pool)

	sentinelErr := errors.New("boom")

	err := repo.WithTransaction(context.Background(), func(tx usecase.ThreadRepository) error {
		_, err := tx.Create(context.Background(), &entity.Thread{
			UserID: userID,
			Name:   "Should not persist",
		})
		if err != nil {
			return err
		}
		// Force the whole transaction to fail after the insert succeeded.
		return sentinelErr
	})
	require.ErrorIs(t, err, sentinelErr)

	var count int
	err = pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM threads WHERE name = $1`, "Should not persist",
	).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count, "insert should have been rolled back, not committed")
}

func TestThreadRepository_Update_Success(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewThreadRepository(pool)

	created, err := repo.Create(context.Background(), &entity.Thread{
		UserID: userID, Name: "Original name",
	})
	require.NoError(t, err)

	newColor := "#123456"
	updated, err := repo.Update(context.Background(), &entity.Thread{
		ID: created.ID, UserID: userID, Name: "Updated name", ColorHex: &newColor,
	})

	require.NoError(t, err)
	require.Equal(t, "Updated name", updated.Name)
	require.Equal(t, "#123456", *updated.ColorHex)
}

func TestThreadRepository_Update_WrongUserID_NotFound(t *testing.T) {
	// UserID in the WHERE clause doubles as the ownership check — a
	// mismatched userID must behave identically to a nonexistent id.
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	otherUserID := seedTestUser(t, pool)
	repo := repository.NewThreadRepository(pool)

	created, err := repo.Create(context.Background(), &entity.Thread{
		UserID: userID, Name: "Mine",
	})
	require.NoError(t, err)

	_, err = repo.Update(context.Background(), &entity.Thread{
		ID: created.ID, UserID: otherUserID, Name: "Hijacked",
	})

	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestThreadRepository_SoftDelete_Success(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewThreadRepository(pool)

	created, err := repo.Create(context.Background(), &entity.Thread{
		UserID: userID, Name: "To delete",
	})
	require.NoError(t, err)

	err = repo.SoftDelete(context.Background(), created.ID, userID)
	require.NoError(t, err)

	var deletedAt *time.Time
	err = pool.QueryRow(context.Background(),
		`SELECT deleted_at FROM threads WHERE id = $1`, created.ID,
	).Scan(&deletedAt)
	require.NoError(t, err)
	require.NotNil(t, deletedAt, "deleted_at should be set after Delete")
}

func TestThreadRepository_SoftDelete_WrongUserID_NoOp(t *testing.T) {
	// Same silent-no-op contract as ThreadImageRepository.Delete: the
	// underlying query is :exec, so a non-matching WHERE (wrong owner
	// here) doesn't surface as an error — the thread must still exist
	// and be unaffected afterwards.
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	otherUserID := seedTestUser(t, pool)
	repo := repository.NewThreadRepository(pool)

	created, err := repo.Create(context.Background(), &entity.Thread{
		UserID: userID, Name: "Not yours",
	})
	require.NoError(t, err)

	err = repo.SoftDelete(context.Background(), created.ID, otherUserID)
	require.NoError(t, err)

	var name string
	var deletedAt *time.Time
	err = pool.QueryRow(context.Background(),
		`SELECT name, deleted_at FROM threads WHERE id = $1`, created.ID,
	).Scan(&name, &deletedAt)
	require.NoError(t, err)
	require.Equal(t, "Not yours", name)
	require.Nil(t, deletedAt, "delete scoped to the wrong user must not soft-delete the row")
}

func TestThreadRepository_WithTransaction_GuardsAgainstNestedTx(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewThreadRepository(pool)

	err := repo.WithTransaction(context.Background(), func(tx usecase.ThreadRepository) error {
		// tx here is bound to a transaction already (its internal pool
		// field is nil). Calling WithTransaction on it again must be
		// rejected instead of silently opening a second, independent
		// transaction — see the guard in thread_repository.go.
		return tx.WithTransaction(context.Background(), func(usecase.ThreadRepository) error {
			return nil
		})
	})
	require.Error(t, err)

	_ = userID // seeded for consistency with the other tests; unused here
}
