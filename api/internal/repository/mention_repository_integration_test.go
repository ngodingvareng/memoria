//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-openapi/testify/v2/require"
	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
	"github.com/ngodingvareng/memoria/internal/repository"
)

func TestMentionRepository_CreateListRemove(t *testing.T) {
	pool := setupTestDB(t)
	ownerID := seedTestUser(t, pool)
	mentionedID := seedTestUser(t, pool)
	momentRepo := repository.NewMomentRepository(pool)
	mentionRepo := repository.NewMentionRepository(pool)

	moment, err := momentRepo.Create(context.Background(), &entity.Moment{
		UserID: ownerID, Origin: enum.MomentOriginManual, OccurredAt: time.Now(), OccurredLocal: time.Now(),
	})
	require.NoError(t, err)

	mention, err := mentionRepo.Create(context.Background(), moment.ID, mentionedID, "Gede")
	require.NoError(t, err)
	require.Equal(t, "Gede", mention.DisplayName)

	list, err := mentionRepo.ListByMomentID(context.Background(), moment.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)

	mentioned, err := mentionRepo.ListMentionedMoments(context.Background(), mentionedID, 20, 0)
	require.NoError(t, err)
	require.Len(t, mentioned, 1)
	require.Equal(t, moment.ID, mentioned[0].ID)

	// Self-removal takes it out of the mentioned user's view, not the owner's.
	require.NoError(t, mentionRepo.Remove(context.Background(), moment.ID, mentionedID))
	mentioned, err = mentionRepo.ListMentionedMoments(context.Background(), mentionedID, 20, 0)
	require.NoError(t, err)
	require.Empty(t, mentioned)

	ownerMoments, err := momentRepo.GetByID(context.Background(), moment.ID, ownerID)
	require.NoError(t, err)
	require.NotNil(t, ownerMoments)
}

func TestMentionRepository_ShareAndUnshareCircle(t *testing.T) {
	pool := setupTestDB(t)
	ownerID := seedTestUser(t, pool)
	momentRepo := repository.NewMomentRepository(pool)
	mentionRepo := repository.NewMentionRepository(pool)
	circleRepo := repository.NewCircleRepository(pool)

	moment, err := momentRepo.Create(context.Background(), &entity.Moment{
		UserID: ownerID, Origin: enum.MomentOriginManual, OccurredAt: time.Now(), OccurredLocal: time.Now(),
	})
	require.NoError(t, err)

	circle, err := circleRepo.Create(context.Background(), &entity.Circle{Name: "Share test"})
	require.NoError(t, err)
	_, err = circleRepo.AddCreatorAsAdmin(context.Background(), circle.ID, ownerID)
	require.NoError(t, err)

	share, err := mentionRepo.ShareToCircle(context.Background(), moment.ID, circle.ID, ownerID)
	require.NoError(t, err)
	require.NotNil(t, share)

	// Idempotent: sharing again is a no-op, not an error.
	share2, err := mentionRepo.ShareToCircle(context.Background(), moment.ID, circle.ID, ownerID)
	require.NoError(t, err)
	require.Nil(t, share2)

	ids, err := mentionRepo.ListSharedCircleIDs(context.Background(), moment.ID)
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{circle.ID}, ids)

	require.NoError(t, mentionRepo.UnshareFromCircle(context.Background(), moment.ID, circle.ID))
	ids, err = mentionRepo.ListSharedCircleIDs(context.Background(), moment.ID)
	require.NoError(t, err)
	require.Empty(t, ids)
}
