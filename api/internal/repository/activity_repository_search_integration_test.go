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

// seedTestActivityFull inserts an activity with full control over
// name/is_fixed_schedule — unlike seedTestActivity in
// activity_image_repository_integration_test.go, which hardcodes both.
func seedTestActivityFull(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID, name string, isFixedSchedule bool) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := pool.QueryRow(context.Background(),
		`INSERT INTO activities (user_id, name, is_fixed_schedule) VALUES ($1, $2, $3) RETURNING id`,
		userID, name, isFixedSchedule,
	).Scan(&id)
	require.NoError(t, err)
	return id
}

// --- GetByID ---

func TestActivityRepository_GetByID_Success(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	activityID := seedTestActivityFull(t, pool, userID, "Morning run", true)
	repo := repository.NewActivityRepository(pool)

	found, err := repo.GetByID(context.Background(), activityID, userID)

	require.NoError(t, err)
	require.Equal(t, "Morning run", found.Name)
	require.Equal(t, userID, found.UserID)
}

func TestActivityRepository_GetByID_NotFound_WrongOwner(t *testing.T) {
	// GetByID's WHERE clause includes user_id specifically so a caller
	// can't fetch someone else's activity by guessing its id alone.
	pool := setupTestDB(t)
	ownerID := seedTestUser(t, pool)
	otherUserID := seedTestUser(t, pool)
	activityID := seedTestActivityFull(t, pool, ownerID, "Private activity", true)
	repo := repository.NewActivityRepository(pool)

	_, err := repo.GetByID(context.Background(), activityID, otherUserID)

	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestActivityRepository_GetByID_NotFound_DoesNotExist(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewActivityRepository(pool)

	_, err := repo.GetByID(context.Background(), uuid.New(), userID)

	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestActivityRepository_GetByID_NotFound_SoftDeleted(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	activityID := seedTestActivityFull(t, pool, userID, "Soon deleted", true)
	repo := repository.NewActivityRepository(pool)

	_, err := pool.Exec(context.Background(),
		`UPDATE activities SET deleted_at = NOW() WHERE id = $1`, activityID)
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), activityID, userID)

	require.ErrorIs(t, err, errs.ErrNotFound)
}

// --- Search ---

func TestActivityRepository_Search_ByNamePartialCaseInsensitive(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	seedTestActivityFull(t, pool, userID, "Morning Run", true)
	seedTestActivityFull(t, pool, userID, "Evening Walk", true)
	repo := repository.NewActivityRepository(pool)

	name := "run" // lowercase, partial — should still match "Morning Run"
	activities, total, err := repo.Search(context.Background(), usecase.SearchActivitiesParams{
		UserID: userID, Name: &name, Limit: 20, Offset: 0,
	})

	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, activities, 1)
	require.Equal(t, "Morning Run", activities[0].Name)
}

func TestActivityRepository_Search_ByIsFixedSchedule(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	seedTestActivityFull(t, pool, userID, "Fixed one", true)
	seedTestActivityFull(t, pool, userID, "Flexible one", false)
	repo := repository.NewActivityRepository(pool)

	fixed := false
	activities, total, err := repo.Search(context.Background(), usecase.SearchActivitiesParams{
		UserID: userID, IsFixedSchedule: &fixed, Limit: 20, Offset: 0,
	})

	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, activities, 1)
	require.Equal(t, "Flexible one", activities[0].Name)
}

func TestActivityRepository_Search_NoFilters_ReturnsAllOwnedByUser(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	otherUserID := seedTestUser(t, pool)
	seedTestActivityFull(t, pool, userID, "Mine A", true)
	seedTestActivityFull(t, pool, userID, "Mine B", false)
	seedTestActivityFull(t, pool, otherUserID, "Not mine", true) // must not leak across users
	repo := repository.NewActivityRepository(pool)

	activities, total, err := repo.Search(context.Background(), usecase.SearchActivitiesParams{
		UserID: userID, Limit: 20, Offset: 0,
	})

	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, activities, 2)
}

func TestActivityRepository_Search_ExcludesSoftDeleted(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	deletedID := seedTestActivityFull(t, pool, userID, "Deleted activity", true)
	seedTestActivityFull(t, pool, userID, "Live activity", true)
	repo := repository.NewActivityRepository(pool)

	_, err := pool.Exec(context.Background(),
		`UPDATE activities SET deleted_at = NOW() WHERE id = $1`, deletedID)
	require.NoError(t, err)

	activities, total, err := repo.Search(context.Background(), usecase.SearchActivitiesParams{
		UserID: userID, Limit: 20, Offset: 0,
	})

	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, activities, 1)
	require.Equal(t, "Live activity", activities[0].Name)
}

func TestActivityRepository_Search_Pagination(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	for i := 0; i < 5; i++ {
		seedTestActivityFull(t, pool, userID, "Paged activity", true)
	}
	repo := repository.NewActivityRepository(pool)

	// Page size 2, second page -> offset 2, expect 2 rows out of 5 total.
	activities, total, err := repo.Search(context.Background(), usecase.SearchActivitiesParams{
		UserID: userID, Limit: 2, Offset: 2,
	})

	require.NoError(t, err)
	require.EqualValues(t, 5, total, "total should reflect ALL matching rows, not just the current page")
	require.Len(t, activities, 2, "this page should only return 2 rows")
}
