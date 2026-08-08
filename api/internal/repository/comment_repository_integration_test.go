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
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/repository"
)

func TestCommentRepository_CreateGetUpdateSoftDelete(t *testing.T) {
	pool := setupTestDB(t)
	ownerID := seedTestUser(t, pool)
	commenterID := seedTestUser(t, pool)
	momentRepo := repository.NewMomentRepository(pool)
	commentRepo := repository.NewCommentRepository(pool)

	moment, err := momentRepo.Create(context.Background(), &entity.Moment{
		UserID: ownerID, Origin: enum.MomentOriginManual, OccurredAt: time.Now(), OccurredLocal: time.Now(),
	})
	require.NoError(t, err)

	created, err := commentRepo.Create(context.Background(), moment.ID, commenterID, nil, "great run!")
	require.NoError(t, err)
	require.Equal(t, "great run!", created.Body)
	require.Nil(t, created.CircleID)

	fetched, err := commentRepo.GetByID(context.Background(), created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, fetched.ID)

	updated, err := commentRepo.Update(context.Background(), created.ID, commenterID, "edited: great run!")
	require.NoError(t, err)
	require.Equal(t, "edited: great run!", updated.Body)

	// Wrong owner cannot update — indistinguishable from a missing row.
	_, err = commentRepo.Update(context.Background(), created.ID, ownerID, "hijacked")
	require.ErrorIs(t, err, errs.ErrNotFound)

	require.NoError(t, commentRepo.SoftDelete(context.Background(), created.ID, commenterID))
	_, err = commentRepo.GetByID(context.Background(), created.ID)
	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestCommentRepository_ListForMoment_VisibilityScoping(t *testing.T) {
	pool := setupTestDB(t)
	ownerID := seedTestUser(t, pool)
	mentionedID := seedTestUser(t, pool)
	circleMemberID := seedTestUser(t, pool)
	momentRepo := repository.NewMomentRepository(pool)
	commentRepo := repository.NewCommentRepository(pool)
	circleRepo := repository.NewCircleRepository(pool)

	moment, err := momentRepo.Create(context.Background(), &entity.Moment{
		UserID: ownerID, Origin: enum.MomentOriginManual, OccurredAt: time.Now(), OccurredLocal: time.Now(),
	})
	require.NoError(t, err)

	circle, err := circleRepo.Create(context.Background(), &entity.Circle{Name: "Scoping test"})
	require.NoError(t, err)
	_, err = circleRepo.AddCreatorAsAdmin(context.Background(), circle.ID, ownerID)
	require.NoError(t, err)
	_, err = pool.Exec(context.Background(),
		`INSERT INTO circle_members (circle_id, user_id, role, can_invite, can_capture) VALUES ($1, $2, 'member', false, true)`,
		circle.ID, circleMemberID)
	require.NoError(t, err)

	mentionComment, err := commentRepo.Create(context.Background(), moment.ID, mentionedID, nil, "mention context")
	require.NoError(t, err)
	circleComment, err := commentRepo.Create(context.Background(), moment.ID, circleMemberID, &circle.ID, "circle context")
	require.NoError(t, err)

	// The mentioned user (mentionAllowed=true, no visible circles) sees
	// only the mention-context comment.
	visible, err := commentRepo.ListForMoment(context.Background(), moment.ID, mentionedID, true, []uuid.UUID{})
	require.NoError(t, err)
	require.Len(t, visible, 1)
	require.Equal(t, mentionComment.ID, visible[0].ID)

	// The circle member (mentionAllowed=false, sees circle.ID) sees only
	// the circle-context comment — never the mention-context one, even
	// though they can technically reach the Moment (FEATURES.md,
	// Response Terms: "Visibility stays scoped to context").
	visible, err = commentRepo.ListForMoment(context.Background(), moment.ID, circleMemberID, false, []uuid.UUID{circle.ID})
	require.NoError(t, err)
	require.Len(t, visible, 1)
	require.Equal(t, circleComment.ID, visible[0].ID)

	// The owner sees both.
	visible, err = commentRepo.ListForMoment(context.Background(), moment.ID, ownerID, true, []uuid.UUID{circle.ID})
	require.NoError(t, err)
	require.Len(t, visible, 2)
}

func TestCommentRepository_ListForMoment_ExcludesMutedAuthors(t *testing.T) {
	pool := setupTestDB(t)
	ownerID := seedTestUser(t, pool)
	commenterID := seedTestUser(t, pool)
	momentRepo := repository.NewMomentRepository(pool)
	commentRepo := repository.NewCommentRepository(pool)

	moment, err := momentRepo.Create(context.Background(), &entity.Moment{
		UserID: ownerID, Origin: enum.MomentOriginManual, OccurredAt: time.Now(), OccurredLocal: time.Now(),
	})
	require.NoError(t, err)

	_, err = commentRepo.Create(context.Background(), moment.ID, commenterID, nil, "hello")
	require.NoError(t, err)

	visible, err := commentRepo.ListForMoment(context.Background(), moment.ID, ownerID, true, []uuid.UUID{})
	require.NoError(t, err)
	require.Len(t, visible, 1)

	_, err = pool.Exec(context.Background(),
		`INSERT INTO user_mutes (muter_user_id, muted_user_id) VALUES ($1, $2)`, ownerID, commenterID)
	require.NoError(t, err)

	visible, err = commentRepo.ListForMoment(context.Background(), moment.ID, ownerID, true, []uuid.UUID{})
	require.NoError(t, err)
	require.Empty(t, visible, "a muted author's comments must be hidden from the muter's own view")
}
