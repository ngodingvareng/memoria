//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/repository"
)

// seedTestActivity depends on seedTestUser and setupTestDB, both defined
// in activity_repository_integration_test.go (same package).
func seedTestActivity(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := pool.QueryRow(context.Background(),
		`INSERT INTO activities (user_id, name, is_fixed_schedule) VALUES ($1, $2, $3) RETURNING id`,
		userID, "Test Activity", true,
	).Scan(&id)
	require.NoError(t, err)
	return id
}

func TestActivityImageRepository_Create(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	activityID := seedTestActivity(t, pool, userID)
	repo := repository.NewActivityImageRepository(pool)

	alt := "a cat sitting on a windowsill"
	created, err := repo.Create(context.Background(), &entity.ActivityImage{
		ActivityID: activityID,
		ImagePath:  "activities/x/cat.jpg",
		ImageAlt:   &alt,
	})

	require.NoError(t, err)
	require.NotNil(t, created)
	require.Equal(t, "activities/x/cat.jpg", created.ImagePath)
	require.NotNil(t, created.ImageAlt)
	require.Equal(t, alt, *created.ImageAlt)
}

func TestActivityImageRepository_ListByActivityID(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	activityID := seedTestActivity(t, pool, userID)
	otherActivityID := seedTestActivity(t, pool, userID)
	repo := repository.NewActivityImageRepository(pool)

	ctx := context.Background()
	_, err := repo.Create(ctx, &entity.ActivityImage{ActivityID: activityID, ImagePath: "activities/x/1.jpg"})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &entity.ActivityImage{ActivityID: activityID, ImagePath: "activities/x/2.jpg"})
	require.NoError(t, err)
	// Belongs to a different activity — must not show up in the list below.
	_, err = repo.Create(ctx, &entity.ActivityImage{ActivityID: otherActivityID, ImagePath: "activities/y/1.jpg"})
	require.NoError(t, err)

	images, err := repo.ListByActivityID(ctx, activityID)

	require.NoError(t, err)
	require.Len(t, images, 2)
}

func TestActivityImageRepository_Delete(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	activityID := seedTestActivity(t, pool, userID)
	repo := repository.NewActivityImageRepository(pool)

	ctx := context.Background()
	created, err := repo.Create(ctx, &entity.ActivityImage{ActivityID: activityID, ImagePath: "activities/x/1.jpg"})
	require.NoError(t, err)

	err = repo.Delete(ctx, activityID, created.ID)
	require.NoError(t, err)

	images, err := repo.ListByActivityID(ctx, activityID)
	require.NoError(t, err)
	require.Empty(t, images)
}

func TestActivityImageRepository_Delete_WrongActivityID_NoOp(t *testing.T) {
	// Delete's WHERE clause includes activity_id specifically so a
	// caller can't delete an image by guessing its id alone without also
	// knowing (and matching) which activity it belongs to.
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	activityID := seedTestActivity(t, pool, userID)
	otherActivityID := seedTestActivity(t, pool, userID)
	repo := repository.NewActivityImageRepository(pool)

	ctx := context.Background()
	created, err := repo.Create(ctx, &entity.ActivityImage{ActivityID: activityID, ImagePath: "activities/x/1.jpg"})
	require.NoError(t, err)

	// DeleteActivityImage's underlying query is a no-op (0 rows
	// affected) when the id/activity_id pair doesn't match — sqlc's
	// :exec query mode doesn't surface that as an error, so the image
	// should still be there afterwards.
	err = repo.Delete(ctx, otherActivityID, created.ID)
	require.NoError(t, err)

	images, err := repo.ListByActivityID(ctx, activityID)
	require.NoError(t, err)
	require.Len(t, images, 1, "image should NOT have been deleted via the wrong activity id")
}
