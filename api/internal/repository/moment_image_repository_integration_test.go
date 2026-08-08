//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/repository"
)

func TestMomentImageRepository_Create(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	momentID := seedTestMoment(t, pool, userID)
	repo := repository.NewMomentImageRepository(pool)

	alt := "wet shoes after the drizzle"
	contentType := "image/jpeg"
	size := int64(12345)
	created, err := repo.Create(context.Background(), &entity.MomentImage{
		MomentID:    momentID,
		ImagePath:   "moments/x/shoes.jpg",
		ImageAlt:    &alt,
		ContentType: &contentType,
		ByteSize:    &size,
	})

	require.NoError(t, err)
	require.NotNil(t, created)
	require.Equal(t, "moments/x/shoes.jpg", created.ImagePath)
	require.NotNil(t, created.ImageAlt)
	require.Equal(t, alt, *created.ImageAlt)
	require.NotNil(t, created.ContentType)
	require.Equal(t, contentType, *created.ContentType)
	require.NotNil(t, created.ByteSize)
	require.Equal(t, size, *created.ByteSize)
	require.False(t, created.MetadataStripped)
}

func TestMomentImageRepository_ListByMomentID(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	momentID := seedTestMoment(t, pool, userID)
	otherMomentID := seedTestMoment(t, pool, userID)
	repo := repository.NewMomentImageRepository(pool)

	ctx := context.Background()
	_, err := repo.Create(ctx, &entity.MomentImage{MomentID: momentID, ImagePath: "moments/x/1.jpg"})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &entity.MomentImage{MomentID: momentID, ImagePath: "moments/x/2.jpg"})
	require.NoError(t, err)
	// Belongs to a different Moment — must not show up in the list below.
	_, err = repo.Create(ctx, &entity.MomentImage{MomentID: otherMomentID, ImagePath: "moments/y/1.jpg"})
	require.NoError(t, err)

	images, err := repo.ListByMomentID(ctx, momentID)

	require.NoError(t, err)
	require.Len(t, images, 2)
}

func TestMomentImageRepository_Delete(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	momentID := seedTestMoment(t, pool, userID)
	repo := repository.NewMomentImageRepository(pool)

	ctx := context.Background()
	created, err := repo.Create(ctx, &entity.MomentImage{MomentID: momentID, ImagePath: "moments/x/1.jpg"})
	require.NoError(t, err)

	err = repo.Delete(ctx, momentID, created.ID)
	require.NoError(t, err)

	images, err := repo.ListByMomentID(ctx, momentID)
	require.NoError(t, err)
	require.Empty(t, images)
}

func TestMomentImageRepository_Delete_WrongMomentID_NoOp(t *testing.T) {
	// Delete's WHERE clause includes moment_id specifically, same
	// contract as ThreadImageRepository.Delete.
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	momentID := seedTestMoment(t, pool, userID)
	otherMomentID := seedTestMoment(t, pool, userID)
	repo := repository.NewMomentImageRepository(pool)

	ctx := context.Background()
	created, err := repo.Create(ctx, &entity.MomentImage{MomentID: momentID, ImagePath: "moments/x/1.jpg"})
	require.NoError(t, err)

	err = repo.Delete(ctx, otherMomentID, created.ID)
	require.NoError(t, err)

	images, err := repo.ListByMomentID(ctx, momentID)
	require.NoError(t, err)
	require.Len(t, images, 1, "image should NOT have been deleted via the wrong moment id")
}
