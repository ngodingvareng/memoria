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

// seedTestCircleThread inserts a collaborative Thread owned by a Circle
// instead of a personal user_id — Search's inclusion rule is keyed off
// circle_id, not creatorUserID.
func seedTestCircleThread(t *testing.T, pool *pgxpool.Pool, creatorUserID, circleID uuid.UUID, name string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := pool.QueryRow(context.Background(),
		`INSERT INTO threads (user_id, circle_id, name) VALUES ($1, $2, $3) RETURNING id`,
		creatorUserID, circleID, name,
	).Scan(&id)
	require.NoError(t, err)
	return id
}

// seedTestCircle inserts a Circle and seats memberUserID as an active
// member (role defaults to 'member').
func seedTestCircle(t *testing.T, pool *pgxpool.Pool, memberUserID uuid.UUID) uuid.UUID {
	t.Helper()
	var circleID uuid.UUID
	err := pool.QueryRow(context.Background(),
		`INSERT INTO circles (name, created_by_user_id) VALUES ($1, $2) RETURNING id`,
		"Test Circle", memberUserID,
	).Scan(&circleID)
	require.NoError(t, err)

	_, err = pool.Exec(context.Background(),
		`INSERT INTO circle_members (circle_id, user_id) VALUES ($1, $2)`,
		circleID, memberUserID,
	)
	require.NoError(t, err)
	return circleID
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

func TestThreadRepository_Search_IncludesCircleThreadsForActiveMember(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	otherMemberID := seedTestUser(t, pool)
	circleID := seedTestCircle(t, pool, userID)
	seedTestThreadFull(t, pool, userID, "My personal thread", false)
	seedTestCircleThread(t, pool, otherMemberID, circleID, "Circle thread")
	repo := repository.NewThreadRepository(pool)

	threads, total, err := repo.Search(context.Background(), usecase.SearchThreadsParams{
		UserID: userID, Limit: 20, Offset: 0,
	})

	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	names := []string{threads[0].Name, threads[1].Name}
	require.ElementsMatch(t, []string{"My personal thread", "Circle thread"}, names)
}

func TestThreadRepository_Search_ExcludesCircleThreadsForNonMember(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	circleAdminID := seedTestUser(t, pool)
	circleID := seedTestCircle(t, pool, circleAdminID)
	seedTestCircleThread(t, pool, circleAdminID, circleID, "Circle thread")
	repo := repository.NewThreadRepository(pool)

	threads, total, err := repo.Search(context.Background(), usecase.SearchThreadsParams{
		UserID: userID, Limit: 20, Offset: 0,
	})

	require.NoError(t, err)
	require.EqualValues(t, 0, total)
	require.Empty(t, threads)
}

func TestThreadRepository_Search_ExcludesCircleThreadsAfterLeaving(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	circleID := seedTestCircle(t, pool, userID)
	seedTestCircleThread(t, pool, userID, circleID, "Circle thread")
	repo := repository.NewThreadRepository(pool)

	_, err := pool.Exec(context.Background(),
		`UPDATE circle_members SET left_at = NOW() WHERE circle_id = $1 AND user_id = $2`,
		circleID, userID)
	require.NoError(t, err)

	threads, total, err := repo.Search(context.Background(), usecase.SearchThreadsParams{
		UserID: userID, Limit: 20, Offset: 0,
	})

	require.NoError(t, err)
	require.EqualValues(t, 0, total)
	require.Empty(t, threads)
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
