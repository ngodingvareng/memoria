//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-openapi/testify/v2/require"
	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/repository"
)

func TestCircleInviteRepository_UsernameInvite_CreateAcceptDecline(t *testing.T) {
	pool := setupTestDB(t)
	adminID := seedTestUser(t, pool)
	inviteeID := seedTestUser(t, pool)
	circleRepo := repository.NewCircleRepository(pool)
	inviteRepo := repository.NewCircleInviteRepository(pool)

	circle, err := circleRepo.Create(context.Background(), &entity.Circle{Name: "Invite test"})
	require.NoError(t, err)
	_, err = circleRepo.AddCreatorAsAdmin(context.Background(), circle.ID, adminID)
	require.NoError(t, err)

	invite, err := inviteRepo.CreateUsernameInvite(context.Background(), circle.ID, adminID, inviteeID, time.Now().Add(24*time.Hour))
	require.NoError(t, err)
	require.Equal(t, inviteeID, *invite.InviteeUserID)

	// A second create for the same (circle, invitee) collides with
	// uq_circle_invites_pending_username — callers must check
	// GetPendingUsernameInvite first, which should now find it.
	pending, err := inviteRepo.GetPendingUsernameInvite(context.Background(), circle.ID, inviteeID)
	require.NoError(t, err)
	require.Equal(t, invite.ID, pending.ID)

	listed, err := inviteRepo.ListPendingByInviteeID(context.Background(), inviteeID)
	require.NoError(t, err)
	require.Len(t, listed, 1)

	accepted, err := inviteRepo.AcceptUsernameInvite(context.Background(), invite.ID, inviteeID)
	require.NoError(t, err)
	require.Equal(t, "accepted", string(accepted.Status))

	// Answering twice is a no-op-turned-404, same reasoning as every
	// other WHERE-scoped state transition in this codebase.
	_, err = inviteRepo.DeclineUsernameInvite(context.Background(), invite.ID, inviteeID)
	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestCircleInviteRepository_UsernameInvite_Revoke(t *testing.T) {
	pool := setupTestDB(t)
	adminID := seedTestUser(t, pool)
	inviteeID := seedTestUser(t, pool)
	circleRepo := repository.NewCircleRepository(pool)
	inviteRepo := repository.NewCircleInviteRepository(pool)

	circle, err := circleRepo.Create(context.Background(), &entity.Circle{Name: "Revoke test"})
	require.NoError(t, err)
	_, err = circleRepo.AddCreatorAsAdmin(context.Background(), circle.ID, adminID)
	require.NoError(t, err)

	invite, err := inviteRepo.CreateUsernameInvite(context.Background(), circle.ID, adminID, inviteeID, time.Now().Add(24*time.Hour))
	require.NoError(t, err)

	revoked, err := inviteRepo.RevokeUsernameInvite(context.Background(), invite.ID, circle.ID)
	require.NoError(t, err)
	require.Equal(t, "revoked", string(revoked.Status))

	_, err = inviteRepo.AcceptUsernameInvite(context.Background(), invite.ID, inviteeID)
	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestCircleInviteRepository_Link_CreateRotateGet(t *testing.T) {
	pool := setupTestDB(t)
	adminID := seedTestUser(t, pool)
	circleRepo := repository.NewCircleRepository(pool)
	inviteRepo := repository.NewCircleInviteRepository(pool)

	circle, err := circleRepo.Create(context.Background(), &entity.Circle{Name: "Link test"})
	require.NoError(t, err)
	_, err = circleRepo.AddCreatorAsAdmin(context.Background(), circle.ID, adminID)
	require.NoError(t, err)

	// First generation: nothing to revoke yet, not an error.
	revoked, err := inviteRepo.RevokeActiveLink(context.Background(), circle.ID)
	require.NoError(t, err)
	require.Nil(t, revoked)

	link, err := inviteRepo.CreateLink(context.Background(), circle.ID, adminID, "hash-1", true)
	require.NoError(t, err)
	require.True(t, *link.RequiresApproval)

	fetched, err := inviteRepo.GetActiveLinkByCircleID(context.Background(), circle.ID)
	require.NoError(t, err)
	require.Equal(t, link.ID, fetched.ID)

	byToken, err := inviteRepo.GetActiveLinkByTokenHash(context.Background(), "hash-1")
	require.NoError(t, err)
	require.Equal(t, link.ID, byToken.ID)

	// Rotation: the old link stops resolving, the new one takes over.
	oldRevoked, err := inviteRepo.RevokeActiveLink(context.Background(), circle.ID)
	require.NoError(t, err)
	require.Equal(t, link.ID, oldRevoked.ID)

	newLink, err := inviteRepo.CreateLink(context.Background(), circle.ID, adminID, "hash-2", false)
	require.NoError(t, err)

	_, err = inviteRepo.GetActiveLinkByTokenHash(context.Background(), "hash-1")
	require.ErrorIs(t, err, errs.ErrNotFound)

	byNewToken, err := inviteRepo.GetActiveLinkByTokenHash(context.Background(), "hash-2")
	require.NoError(t, err)
	require.Equal(t, newLink.ID, byNewToken.ID)
}

func TestCircleInviteRepository_AddMembers_IdempotentForActiveMembers(t *testing.T) {
	pool := setupTestDB(t)
	adminID := seedTestUser(t, pool)
	newUserID := seedTestUser(t, pool)
	circleRepo := repository.NewCircleRepository(pool)
	inviteRepo := repository.NewCircleInviteRepository(pool)

	circle, err := circleRepo.Create(context.Background(), &entity.Circle{Name: "Members test"})
	require.NoError(t, err)
	_, err = circleRepo.AddCreatorAsAdmin(context.Background(), circle.ID, adminID)
	require.NoError(t, err)

	added, err := inviteRepo.AddMembers(context.Background(), circle.ID, []uuid.UUID{newUserID, adminID})
	require.NoError(t, err)
	// adminID is already an active member — only newUserID is newly seated.
	require.Len(t, added, 1)
	require.Equal(t, newUserID, added[0].UserID)
}
