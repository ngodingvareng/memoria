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

func TestReactionRepository_UpsertChangesKindDeletes(t *testing.T) {
	pool := setupTestDB(t)
	ownerID := seedTestUser(t, pool)
	reactorID := seedTestUser(t, pool)
	momentRepo := repository.NewMomentRepository(pool)
	reactionRepo := repository.NewReactionRepository(pool)

	moment, err := momentRepo.Create(context.Background(), &entity.Moment{
		UserID: ownerID, Origin: enum.MomentOriginManual, OccurredAt: time.Now(), OccurredLocal: time.Now(),
	})
	require.NoError(t, err)

	created, err := reactionRepo.Upsert(context.Background(), moment.ID, reactorID, nil, enum.ReactionKindHeart)
	require.NoError(t, err)
	require.Equal(t, enum.ReactionKindHeart, created.Kind)

	// Reacting again with a different kind changes it in place — no
	// second row (matches the unique index on moment/user/circle).
	changed, err := reactionRepo.Upsert(context.Background(), moment.ID, reactorID, nil, enum.ReactionKindJoy)
	require.NoError(t, err)
	require.Equal(t, created.ID, changed.ID)
	require.Equal(t, enum.ReactionKindJoy, changed.Kind)

	list, err := reactionRepo.ListForMoment(context.Background(), moment.ID, ownerID, true, []uuid.UUID{})
	require.NoError(t, err)
	require.Len(t, list, 1)

	require.NoError(t, reactionRepo.Delete(context.Background(), moment.ID, reactorID, nil))
	list, err = reactionRepo.ListForMoment(context.Background(), moment.ID, ownerID, true, []uuid.UUID{})
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestReactionRepository_SameUserCanReactOnceMoreOnCircle(t *testing.T) {
	// The unique index is (moment_id, user_id, circle_id) — the same
	// user reacting once in the mention context and once more in a
	// Circle context on the same Moment are two distinct rows.
	pool := setupTestDB(t)
	ownerID := seedTestUser(t, pool)
	reactorID := seedTestUser(t, pool)
	momentRepo := repository.NewMomentRepository(pool)
	reactionRepo := repository.NewReactionRepository(pool)
	circleRepo := repository.NewCircleRepository(pool)

	moment, err := momentRepo.Create(context.Background(), &entity.Moment{
		UserID: ownerID, Origin: enum.MomentOriginManual, OccurredAt: time.Now(), OccurredLocal: time.Now(),
	})
	require.NoError(t, err)
	circle, err := circleRepo.Create(context.Background(), &entity.Circle{Name: "Reaction test"})
	require.NoError(t, err)
	_, err = circleRepo.AddCreatorAsAdmin(context.Background(), circle.ID, ownerID)
	require.NoError(t, err)

	_, err = reactionRepo.Upsert(context.Background(), moment.ID, reactorID, nil, enum.ReactionKindHeart)
	require.NoError(t, err)
	_, err = reactionRepo.Upsert(context.Background(), moment.ID, reactorID, &circle.ID, enum.ReactionKindTender)
	require.NoError(t, err)

	list, err := reactionRepo.ListForMoment(context.Background(), moment.ID, ownerID, true, []uuid.UUID{circle.ID})
	require.NoError(t, err)
	require.Len(t, list, 2)
}
