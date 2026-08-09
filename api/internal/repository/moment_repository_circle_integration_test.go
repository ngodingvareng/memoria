//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/ngodingvareng/memoria/internal/repository"
)

// seedTestMomentInThread depends on seedTestUser/setupTestDB (defined in
// thread_repository_integration_test.go) and seedTestCircle/
// seedTestCircleThread (defined in
// thread_repository_search_integration_test.go), all in this package.
func seedTestMomentInThread(t *testing.T, pool *pgxpool.Pool, userID, threadID uuid.UUID) uuid.UUID {
	t.Helper()
	occurredAt := time.Now()
	var id uuid.UUID
	err := pool.QueryRow(context.Background(),
		`INSERT INTO moments (user_id, thread_id, occurred_at, occurred_local, occurred_utc_offset_minutes)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		userID, threadID, occurredAt, occurredAt, 0,
	).Scan(&id)
	require.NoError(t, err)
	return id
}

func shareTestMomentToCircle(t *testing.T, pool *pgxpool.Pool, momentID, circleID, sharedByUserID uuid.UUID) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO moment_circles (moment_id, circle_id, shared_by_user_id) VALUES ($1, $2, $3)`,
		momentID, circleID, sharedByUserID)
	require.NoError(t, err)
}

func TestMomentRepository_ListByCircle_IncludesThreadMomentsAndSharedMoments(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	circleID := seedTestCircle(t, pool, userID)
	threadID := seedTestCircleThread(t, pool, userID, circleID, "Team Thread")
	threadMomentID := seedTestMomentInThread(t, pool, userID, threadID)

	otherUserID := seedTestUser(t, pool)
	sharedMomentID := seedTestMoment(t, pool, otherUserID)
	shareTestMomentToCircle(t, pool, sharedMomentID, circleID, otherUserID)

	// A personal moment that was never shared into the circle must not
	// leak in.
	seedTestMoment(t, pool, userID)

	repo := repository.NewMomentRepository(pool)

	moments, err := repo.ListByCircle(context.Background(), circleID, 20, 0)

	require.NoError(t, err)
	ids := make([]uuid.UUID, len(moments))
	for i, m := range moments {
		ids[i] = m.ID
	}
	require.ElementsMatch(t, []uuid.UUID{threadMomentID, sharedMomentID}, ids)
}

func TestMomentRepository_ListByCircle_NoMoments_ReturnsEmpty(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	circleID := seedTestCircle(t, pool, userID)
	repo := repository.NewMomentRepository(pool)

	moments, err := repo.ListByCircle(context.Background(), circleID, 20, 0)

	require.NoError(t, err)
	require.Empty(t, moments)
}
