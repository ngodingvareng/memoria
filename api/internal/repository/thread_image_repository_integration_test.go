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

// seedTestThread depends on seedTestUser and setupTestDB, both defined
// in thread_repository_integration_test.go (same package).
func seedTestThread(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := pool.QueryRow(context.Background(),
		`INSERT INTO threads (user_id, name) VALUES ($1, $2) RETURNING id`,
		userID, "Test Thread",
	).Scan(&id)
	require.NoError(t, err)
	return id
}

func TestThreadImageRepository_Create(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	threadID := seedTestThread(t, pool, userID)
	repo := repository.NewThreadImageRepository(pool)

	alt := "a cat sitting on a windowsill"
	created, err := repo.Create(context.Background(), &entity.ThreadImage{
		ThreadID:  threadID,
		ImagePath: "threads/x/cat.jpg",
		ImageAlt:  &alt,
	})

	require.NoError(t, err)
	require.NotNil(t, created)
	require.Equal(t, "threads/x/cat.jpg", created.ImagePath)
	require.NotNil(t, created.ImageAlt)
	require.Equal(t, alt, *created.ImageAlt)
}

func TestThreadImageRepository_ListByThreadID(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	threadID := seedTestThread(t, pool, userID)
	otherThreadID := seedTestThread(t, pool, userID)
	repo := repository.NewThreadImageRepository(pool)

	ctx := context.Background()
	_, err := repo.Create(ctx, &entity.ThreadImage{ThreadID: threadID, ImagePath: "threads/x/1.jpg"})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &entity.ThreadImage{ThreadID: threadID, ImagePath: "threads/x/2.jpg"})
	require.NoError(t, err)
	// Belongs to a different thread — must not show up in the list below.
	_, err = repo.Create(ctx, &entity.ThreadImage{ThreadID: otherThreadID, ImagePath: "threads/y/1.jpg"})
	require.NoError(t, err)

	images, err := repo.ListByThreadID(ctx, threadID)

	require.NoError(t, err)
	require.Len(t, images, 2)
}

func TestThreadImageRepository_Delete(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	threadID := seedTestThread(t, pool, userID)
	repo := repository.NewThreadImageRepository(pool)

	ctx := context.Background()
	created, err := repo.Create(ctx, &entity.ThreadImage{ThreadID: threadID, ImagePath: "threads/x/1.jpg"})
	require.NoError(t, err)

	deleted, err := repo.Delete(ctx, threadID, created.ID)
	require.NoError(t, err)
	require.NotNil(t, deleted)
	require.Equal(t, "threads/x/1.jpg", deleted.ImagePath)

	images, err := repo.ListByThreadID(ctx, threadID)
	require.NoError(t, err)
	require.Empty(t, images)
}

func TestThreadImageRepository_Delete_WrongThreadID_NoOp(t *testing.T) {
	// Delete's WHERE clause includes thread_id specifically so a
	// caller can't delete an image by guessing its id alone without also
	// knowing (and matching) which thread it belongs to.
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	threadID := seedTestThread(t, pool, userID)
	otherThreadID := seedTestThread(t, pool, userID)
	repo := repository.NewThreadImageRepository(pool)

	ctx := context.Background()
	created, err := repo.Create(ctx, &entity.ThreadImage{ThreadID: threadID, ImagePath: "threads/x/1.jpg"})
	require.NoError(t, err)

	// DeleteThreadImage's underlying query returns no rows when the
	// id/thread_id pair doesn't match — the repository treats
	// pgx.ErrNoRows as a silent no-op (nil, nil), so the image should
	// still be there afterwards.
	deleted, err := repo.Delete(ctx, otherThreadID, created.ID)
	require.NoError(t, err)
	require.Nil(t, deleted)

	images, err := repo.ListByThreadID(ctx, threadID)
	require.NoError(t, err)
	require.Len(t, images, 1, "image should NOT have been deleted via the wrong thread id")
}
