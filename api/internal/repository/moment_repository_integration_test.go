//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/repository"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

// seedTestMoment depends on seedTestUser and setupTestDB, both defined
// in thread_repository_integration_test.go (same package).
func seedTestMoment(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID) uuid.UUID {
	t.Helper()
	occurredAt := time.Now()
	var id uuid.UUID
	err := pool.QueryRow(context.Background(),
		`INSERT INTO moments (user_id, occurred_at, occurred_local, occurred_utc_offset_minutes)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, occurredAt, occurredAt, 0,
	).Scan(&id)
	require.NoError(t, err)
	return id
}

func newTestMoment(userID uuid.UUID, occurredAt time.Time) *entity.Moment {
	return &entity.Moment{
		UserID:        userID,
		Origin:        enum.MomentOriginManual,
		OccurredAt:    occurredAt,
		OccurredLocal: occurredAt,
	}
}

func TestMomentRepository_Create_MinimalMoment(t *testing.T) {
	// A Moment saved with nothing attached is still a complete, valid
	// Moment (FEATURES.md Principle 7) — no ThreadID, no Note, etc.
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewMomentRepository(pool)

	occurredAt := time.Date(2026, 8, 4, 5, 30, 0, 0, time.UTC)
	created, err := repo.Create(context.Background(), newTestMoment(userID, occurredAt))

	require.NoError(t, err)
	require.NotNil(t, created)
	require.Nil(t, created.ThreadID)
	require.Nil(t, created.Note)
	require.WithinDuration(t, occurredAt, created.OccurredAt, time.Second)
	require.False(t, created.RecordedAt.IsZero(), "recorded_at should default to NOW() when omitted")
}

func TestMomentRepository_Create_WithLatLng_RoundTrips(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewMomentRepository(pool)

	moment := newTestMoment(userID, time.Now())
	lat, lng := -6.914744, 107.609810
	moment.Latitude = &lat
	moment.Longitude = &lng

	created, err := repo.Create(context.Background(), moment)

	require.NoError(t, err)
	require.NotNil(t, created.Latitude)
	require.NotNil(t, created.Longitude)
	require.InDelta(t, lat, *created.Latitude, 0.000001)
	require.InDelta(t, lng, *created.Longitude, 0.000001)
}

func TestMomentRepository_Create_IdempotentOnClientID(t *testing.T) {
	// Mirrors uq_moments_user_client_id: syncing the same offline
	// capture twice must not create a duplicate.
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewMomentRepository(pool)

	clientID := uuid.New()
	moment := newTestMoment(userID, time.Now())
	moment.ClientID = &clientID

	first, err := repo.Create(context.Background(), moment)
	require.NoError(t, err)
	require.NotNil(t, first)

	second, err := repo.Create(context.Background(), moment)
	require.NoError(t, err, "ON CONFLICT DO NOTHING must not surface as an error")
	require.NotNil(t, second, "the already-persisted Moment should be re-fetched, not dropped")
	require.Equal(t, first.ID, second.ID, "the retry must return the same Moment, not a new one")

	var count int
	err = pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM moments WHERE client_id = $1`, clientID,
	).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestMomentRepository_GetByID_WrongUserID_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	otherUserID := seedTestUser(t, pool)
	repo := repository.NewMomentRepository(pool)

	created, err := repo.Create(context.Background(), newTestMoment(userID, time.Now()))
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), created.ID, otherUserID)
	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestMomentRepository_Update_Success(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewMomentRepository(pool)

	created, err := repo.Create(context.Background(), newTestMoment(userID, time.Now()))
	require.NoError(t, err)

	note := "Managed to run 5km despite the drizzle"
	newOccurredAt := time.Date(2026, 8, 5, 6, 0, 0, 0, time.UTC)
	updated, err := repo.Update(context.Background(), &entity.Moment{
		ID:            created.ID,
		UserID:        userID,
		OccurredAt:    newOccurredAt,
		OccurredLocal: newOccurredAt,
		Note:          &note,
	})

	require.NoError(t, err)
	require.NotNil(t, updated.Note)
	require.Equal(t, note, *updated.Note)
	require.WithinDuration(t, newOccurredAt, updated.OccurredAt, time.Second)
	// recorded_at is never touched by Update, per FEATURES.md.
	require.Equal(t, created.RecordedAt.Unix(), updated.RecordedAt.Unix())
}

func TestMomentRepository_Update_WrongUserID_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	otherUserID := seedTestUser(t, pool)
	repo := repository.NewMomentRepository(pool)

	created, err := repo.Create(context.Background(), newTestMoment(userID, time.Now()))
	require.NoError(t, err)

	_, err = repo.Update(context.Background(), &entity.Moment{
		ID:     created.ID,
		UserID: otherUserID,
	})
	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestMomentRepository_SoftDelete_ExcludesFromListAndGet(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewMomentRepository(pool)

	created, err := repo.Create(context.Background(), newTestMoment(userID, time.Now()))
	require.NoError(t, err)

	err = repo.SoftDelete(context.Background(), created.ID, userID)
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), created.ID, userID)
	require.ErrorIs(t, err, errs.ErrNotFound)

	moments, err := repo.ListByUserID(context.Background(), userID, 20, 0)
	require.NoError(t, err)
	require.Empty(t, moments)
}

func TestMomentRepository_SoftDelete_WrongUserID_NoOp(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	otherUserID := seedTestUser(t, pool)
	repo := repository.NewMomentRepository(pool)

	created, err := repo.Create(context.Background(), newTestMoment(userID, time.Now()))
	require.NoError(t, err)

	err = repo.SoftDelete(context.Background(), created.ID, otherUserID)
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), created.ID, userID)
	require.NoError(t, err, "moment must still exist — delete scoped to the wrong user")
}

func TestMomentRepository_ListByUserID_NewestFirst(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	otherUserID := seedTestUser(t, pool)
	repo := repository.NewMomentRepository(pool)

	older := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	_, err := repo.Create(context.Background(), newTestMoment(userID, older))
	require.NoError(t, err)
	_, err = repo.Create(context.Background(), newTestMoment(userID, newer))
	require.NoError(t, err)
	// Belongs to a different user — must not show up below.
	_, err = repo.Create(context.Background(), newTestMoment(otherUserID, newer))
	require.NoError(t, err)

	moments, err := repo.ListByUserID(context.Background(), userID, 20, 0)

	require.NoError(t, err)
	require.Len(t, moments, 2)
	require.True(t, moments[0].OccurredAt.After(moments[1].OccurredAt), "newest occurrence must come first")
}

func TestMomentRepository_ListByThreadID_ScopedToThread(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	threadID := seedTestThread(t, pool, userID)
	otherThreadID := seedTestThread(t, pool, userID)
	repo := repository.NewMomentRepository(pool)

	inThread := newTestMoment(userID, time.Now())
	inThread.ThreadID = &threadID
	_, err := repo.Create(context.Background(), inThread)
	require.NoError(t, err)

	inOtherThread := newTestMoment(userID, time.Now())
	inOtherThread.ThreadID = &otherThreadID
	_, err = repo.Create(context.Background(), inOtherThread)
	require.NoError(t, err)

	noThread := newTestMoment(userID, time.Now())
	_, err = repo.Create(context.Background(), noThread)
	require.NoError(t, err)

	moments, err := repo.ListByThreadID(context.Background(), threadID, 20, 0)

	require.NoError(t, err)
	require.Len(t, moments, 1)
	require.NotNil(t, moments[0].ThreadID)
	require.Equal(t, threadID, *moments[0].ThreadID)
}

func TestMomentRepository_Search_MatchesNoteText(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewMomentRepository(pool)

	matching := newTestMoment(userID, time.Now())
	note := "wet shoes after the drizzle"
	matching.Note = &note
	_, err := repo.Create(context.Background(), matching)
	require.NoError(t, err)

	nonMatching := newTestMoment(userID, time.Now())
	otherNote := "coffee with an old friend"
	nonMatching.Note = &otherNote
	_, err = repo.Create(context.Background(), nonMatching)
	require.NoError(t, err)

	results, err := repo.Search(context.Background(), userID, "drizzle", 20, 0)

	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, matching.ClientID, results[0].ClientID)
	require.NotNil(t, results[0].Note)
	require.Equal(t, note, *results[0].Note)
}

func TestMomentRepository_WithTransaction_RollsBackOnError(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewMomentRepository(pool)

	sentinelErr := errors.New("boom")
	occurredAt := time.Now()

	err := repo.WithTransaction(context.Background(), func(tx usecase.MomentRepository) error {
		_, err := tx.Create(context.Background(), newTestMoment(userID, occurredAt))
		if err != nil {
			return err
		}
		return sentinelErr
	})
	require.ErrorIs(t, err, sentinelErr)

	moments, err := repo.ListByUserID(context.Background(), userID, 20, 0)
	require.NoError(t, err)
	require.Empty(t, moments, "insert should have been rolled back, not committed")
}
