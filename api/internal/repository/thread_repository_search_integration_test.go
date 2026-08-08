//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/repository"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

// seedTestThreadFull inserts a thread with full control over
// name/archived_at — unlike seedTestThread in
// thread_image_repository_integration_test.go, which hardcodes name.
func seedTestThreadFull(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID, name string, archived bool) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := pool.QueryRow(context.Background(),
		`INSERT INTO threads (user_id, name, archived_at) VALUES ($1, $2, CASE WHEN $3 THEN NOW() ELSE NULL END) RETURNING id`,
		userID, name, archived,
	).Scan(&id)
	require.NoError(t, err)
	return id
}

// --- GetByID ---

func TestThreadRepository_GetByID_Success(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	threadID := seedTestThreadFull(t, pool, userID, "Morning run", false)
	repo := repository.NewThreadRepository(pool)

	found, err := repo.GetByID(context.Background(), threadID, userID)

	require.NoError(t, err)
	require.Equal(t, "Morning run", found.Name)
	require.Equal(t, userID, found.UserID)
}

func TestThreadRepository_GetByID_NotFound_WrongOwner(t *testing.T) {
	// GetByID's WHERE clause includes user_id specifically so a caller
	// can't fetch someone else's thread by guessing its id alone.
	pool := setupTestDB(t)
	ownerID := seedTestUser(t, pool)
	otherUserID := seedTestUser(t, pool)
	threadID := seedTestThreadFull(t, pool, ownerID, "Private thread", false)
	repo := repository.NewThreadRepository(pool)

	_, err := repo.GetByID(context.Background(), threadID, otherUserID)

	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestThreadRepository_GetByID_NotFound_DoesNotExist(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewThreadRepository(pool)

	_, err := repo.GetByID(context.Background(), uuid.New(), userID)

	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestThreadRepository_GetByID_NotFound_SoftDeleted(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	threadID := seedTestThreadFull(t, pool, userID, "Soon deleted", false)
	repo := repository.NewThreadRepository(pool)

	_, err := pool.Exec(context.Background(),
		`UPDATE threads SET deleted_at = NOW() WHERE id = $1`, threadID)
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), threadID, userID)

	require.ErrorIs(t, err, errs.ErrNotFound)
}

// --- Search ---

func TestThreadRepository_Search_ByNamePartialCaseInsensitive(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	seedTestThreadFull(t, pool, userID, "Morning Run", false)
	seedTestThreadFull(t, pool, userID, "Evening Walk", false)
	repo := repository.NewThreadRepository(pool)

	name := "run" // lowercase, partial — should still match "Morning Run"
	threads, total, err := repo.Search(context.Background(), usecase.SearchThreadsParams{
		UserID: userID, Name: &name, Limit: 20, Offset: 0,
	})

	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, threads, 1)
	require.Equal(t, "Morning Run", threads[0].Name)
}

func TestThreadRepository_Search_ByArchived(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	seedTestThreadFull(t, pool, userID, "Archived one", true)
	seedTestThreadFull(t, pool, userID, "Active one", false)
	repo := repository.NewThreadRepository(pool)

	archived := true
	threads, total, err := repo.Search(context.Background(), usecase.SearchThreadsParams{
		UserID: userID, Archived: &archived, Limit: 20, Offset: 0,
	})

	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, threads, 1)
	require.Equal(t, "Archived one", threads[0].Name)
}

func TestThreadRepository_Search_NoFilters_ReturnsAllOwnedByUser(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	otherUserID := seedTestUser(t, pool)
	seedTestThreadFull(t, pool, userID, "Mine A", false)
	seedTestThreadFull(t, pool, userID, "Mine B", true)
	seedTestThreadFull(t, pool, otherUserID, "Not mine", false) // must not leak across users
	repo := repository.NewThreadRepository(pool)

	threads, total, err := repo.Search(context.Background(), usecase.SearchThreadsParams{
		UserID: userID, Limit: 20, Offset: 0,
	})

	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, threads, 2)
}

func TestThreadRepository_Search_ExcludesSoftDeleted(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	deletedID := seedTestThreadFull(t, pool, userID, "Deleted thread", false)
	seedTestThreadFull(t, pool, userID, "Live thread", false)
	repo := repository.NewThreadRepository(pool)

	_, err := pool.Exec(context.Background(),
		`UPDATE threads SET deleted_at = NOW() WHERE id = $1`, deletedID)
	require.NoError(t, err)

	threads, total, err := repo.Search(context.Background(), usecase.SearchThreadsParams{
		UserID: userID, Limit: 20, Offset: 0,
	})

	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, threads, 1)
	require.Equal(t, "Live thread", threads[0].Name)
}

func TestThreadRepository_Search_Pagination(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	for i := 0; i < 5; i++ {
		seedTestThreadFull(t, pool, userID, "Paged thread", false)
	}
	repo := repository.NewThreadRepository(pool)

	// Page size 2, second page -> offset 2, expect 2 rows out of 5 total.
	threads, total, err := repo.Search(context.Background(), usecase.SearchThreadsParams{
		UserID: userID, Limit: 2, Offset: 2,
	})

	require.NoError(t, err)
	require.EqualValues(t, 5, total, "total should reflect ALL matching rows, not just the current page")
	require.Len(t, threads, 2, "this page should only return 2 rows")
}
