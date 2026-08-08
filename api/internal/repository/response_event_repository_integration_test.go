//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-openapi/testify/v2/require"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
	"github.com/ngodingvareng/memoria/internal/repository"
)

// responseEventRow mirrors the columns fetchResponseEvent reads back
// directly, since the repository is write-only (see
// response_event_repository.go) and nothing else in the codebase reads
// response_events yet.
type responseEventRow struct {
	RecipientUserID uuid.UUID
	ActorUserID     *uuid.UUID
	MomentID        uuid.UUID
	Kind            string
	CommentID       *uuid.UUID
	ReactionID      *uuid.UUID
	DigestedAt      *time.Time
}

func fetchResponseEvent(t *testing.T, pool *pgxpool.Pool, momentID uuid.UUID) responseEventRow {
	t.Helper()
	var row responseEventRow
	err := pool.QueryRow(context.Background(),
		`SELECT recipient_user_id, actor_user_id, moment_id, kind, comment_id, reaction_id, digested_at
		 FROM response_events WHERE moment_id = $1`,
		momentID,
	).Scan(&row.RecipientUserID, &row.ActorUserID, &row.MomentID, &row.Kind, &row.CommentID, &row.ReactionID, &row.DigestedAt)
	require.NoError(t, err)
	return row
}

func TestResponseEventRepository_Create_Comment(t *testing.T) {
	pool := setupTestDB(t)
	ownerID := seedTestUser(t, pool)
	commenterID := seedTestUser(t, pool)
	momentRepo := repository.NewMomentRepository(pool)
	commentRepo := repository.NewCommentRepository(pool)
	repo := repository.NewResponseEventRepository(pool)
	ctx := context.Background()

	moment, err := momentRepo.Create(ctx, &entity.Moment{
		UserID: ownerID, Origin: enum.MomentOriginManual, OccurredAt: time.Now(), OccurredLocal: time.Now(),
	})
	require.NoError(t, err)
	comment, err := commentRepo.Create(ctx, moment.ID, commenterID, nil, "nice one!")
	require.NoError(t, err)

	err = repo.Create(ctx, &entity.ResponseEvent{
		RecipientUserID: ownerID,
		ActorUserID:     &commenterID,
		MomentID:        moment.ID,
		Kind:            enum.ResponseEventKindComment,
		CommentID:       &comment.ID,
	})
	require.NoError(t, err)

	row := fetchResponseEvent(t, pool, moment.ID)
	require.Equal(t, ownerID, row.RecipientUserID)
	require.NotNil(t, row.ActorUserID)
	require.Equal(t, commenterID, *row.ActorUserID)
	require.Equal(t, string(enum.ResponseEventKindComment), row.Kind)
	require.NotNil(t, row.CommentID)
	require.Equal(t, comment.ID, *row.CommentID)
	require.Nil(t, row.ReactionID)
	require.Nil(t, row.DigestedAt)
}

func TestResponseEventRepository_Create_Reaction(t *testing.T) {
	pool := setupTestDB(t)
	ownerID := seedTestUser(t, pool)
	reactorID := seedTestUser(t, pool)
	momentRepo := repository.NewMomentRepository(pool)
	reactionRepo := repository.NewReactionRepository(pool)
	repo := repository.NewResponseEventRepository(pool)
	ctx := context.Background()

	moment, err := momentRepo.Create(ctx, &entity.Moment{
		UserID: ownerID, Origin: enum.MomentOriginManual, OccurredAt: time.Now(), OccurredLocal: time.Now(),
	})
	require.NoError(t, err)
	reaction, err := reactionRepo.Upsert(ctx, moment.ID, reactorID, nil, enum.ReactionKindHeart)
	require.NoError(t, err)

	err = repo.Create(ctx, &entity.ResponseEvent{
		RecipientUserID: ownerID,
		ActorUserID:     &reactorID,
		MomentID:        moment.ID,
		Kind:            enum.ResponseEventKindReaction,
		ReactionID:      &reaction.ID,
	})
	require.NoError(t, err)

	row := fetchResponseEvent(t, pool, moment.ID)
	require.Equal(t, string(enum.ResponseEventKindReaction), row.Kind)
	require.NotNil(t, row.ReactionID)
	require.Equal(t, reaction.ID, *row.ReactionID)
	require.Nil(t, row.CommentID)
}

func TestResponseEventRepository_Create_ActorDeleted_SetsActorNull(t *testing.T) {
	// actor_user_id is ON DELETE SET NULL (unlike recipient/moment, which
	// cascade) so the digest queue keeps a row about the response even
	// after the actor's account is gone.
	pool := setupTestDB(t)
	ownerID := seedTestUser(t, pool)
	reactorID := seedTestUser(t, pool)
	momentRepo := repository.NewMomentRepository(pool)
	reactionRepo := repository.NewReactionRepository(pool)
	repo := repository.NewResponseEventRepository(pool)
	ctx := context.Background()

	moment, err := momentRepo.Create(ctx, &entity.Moment{
		UserID: ownerID, Origin: enum.MomentOriginManual, OccurredAt: time.Now(), OccurredLocal: time.Now(),
	})
	require.NoError(t, err)
	reaction, err := reactionRepo.Upsert(ctx, moment.ID, reactorID, nil, enum.ReactionKindHeart)
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, &entity.ResponseEvent{
		RecipientUserID: ownerID,
		ActorUserID:     &reactorID,
		MomentID:        moment.ID,
		Kind:            enum.ResponseEventKindReaction,
		ReactionID:      &reaction.ID,
	}))

	_, err = pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, reactorID)
	require.NoError(t, err)

	row := fetchResponseEvent(t, pool, moment.ID)
	require.Nil(t, row.ActorUserID)
}

func TestResponseEventRepository_Create_MomentDeleted_Cascades(t *testing.T) {
	pool := setupTestDB(t)
	ownerID := seedTestUser(t, pool)
	commenterID := seedTestUser(t, pool)
	momentRepo := repository.NewMomentRepository(pool)
	commentRepo := repository.NewCommentRepository(pool)
	repo := repository.NewResponseEventRepository(pool)
	ctx := context.Background()

	moment, err := momentRepo.Create(ctx, &entity.Moment{
		UserID: ownerID, Origin: enum.MomentOriginManual, OccurredAt: time.Now(), OccurredLocal: time.Now(),
	})
	require.NoError(t, err)
	comment, err := commentRepo.Create(ctx, moment.ID, commenterID, nil, "nice one!")
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, &entity.ResponseEvent{
		RecipientUserID: ownerID,
		ActorUserID:     &commenterID,
		MomentID:        moment.ID,
		Kind:            enum.ResponseEventKindComment,
		CommentID:       &comment.ID,
	}))

	_, err = pool.Exec(ctx, `DELETE FROM moments WHERE id = $1`, moment.ID)
	require.NoError(t, err)

	var count int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM response_events WHERE moment_id = $1`, moment.ID).Scan(&count)
	require.NoError(t, err)
	require.Zero(t, count)
}
