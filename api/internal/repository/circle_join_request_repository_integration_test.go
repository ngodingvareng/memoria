//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/go-openapi/testify/v2/require"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/repository"
)

func TestCircleJoinRequestRepository_CreateApprove(t *testing.T) {
	pool := setupTestDB(t)
	adminID := seedTestUser(t, pool)
	requesterID := seedTestUser(t, pool)
	circleRepo := repository.NewCircleRepository(pool)
	joinRepo := repository.NewCircleJoinRequestRepository(pool)

	circle, err := circleRepo.Create(context.Background(), &entity.Circle{Name: "Join test"})
	require.NoError(t, err)
	_, err = circleRepo.AddCreatorAsAdmin(context.Background(), circle.ID, adminID)
	require.NoError(t, err)

	request, err := joinRepo.Create(context.Background(), circle.ID, requesterID, nil)
	require.NoError(t, err)

	pending, err := joinRepo.GetPendingByUser(context.Background(), circle.ID, requesterID)
	require.NoError(t, err)
	require.Equal(t, request.ID, pending.ID)

	list, err := joinRepo.ListPendingByCircleID(context.Background(), circle.ID, 20, 0)
	require.NoError(t, err)
	require.Len(t, list, 1)

	count, err := joinRepo.CountPendingByCircleID(context.Background(), circle.ID)
	require.NoError(t, err)
	require.EqualValues(t, 1, count)

	approved, err := joinRepo.Approve(context.Background(), request.ID, circle.ID, adminID)
	require.NoError(t, err)
	require.Equal(t, "approved", string(approved.Status))

	// Deciding twice is a no-op-turned-404.
	_, err = joinRepo.Reject(context.Background(), request.ID, circle.ID, adminID)
	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestCircleJoinRequestRepository_Cancel(t *testing.T) {
	pool := setupTestDB(t)
	adminID := seedTestUser(t, pool)
	requesterID := seedTestUser(t, pool)
	circleRepo := repository.NewCircleRepository(pool)
	joinRepo := repository.NewCircleJoinRequestRepository(pool)

	circle, err := circleRepo.Create(context.Background(), &entity.Circle{Name: "Cancel test"})
	require.NoError(t, err)
	_, err = circleRepo.AddCreatorAsAdmin(context.Background(), circle.ID, adminID)
	require.NoError(t, err)

	request, err := joinRepo.Create(context.Background(), circle.ID, requesterID, nil)
	require.NoError(t, err)

	cancelled, err := joinRepo.Cancel(context.Background(), request.ID, requesterID)
	require.NoError(t, err)
	require.Equal(t, "cancelled", string(cancelled.Status))
	require.Nil(t, cancelled.DecidedByUserID, "the requester withdrew — nobody in the circle made a decision")
}
