//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/go-openapi/testify/v2/require"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/repository"
)

func TestCircleRepository_Create_And_AddCreatorAsAdmin(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewCircleRepository(pool)

	created, err := repo.Create(context.Background(), &entity.Circle{Name: "NgodingVareng"})
	require.NoError(t, err)

	member, err := repo.AddCreatorAsAdmin(context.Background(), created.ID, userID)
	require.NoError(t, err)
	require.Equal(t, enum.CircleRoleAdmin, member.Role)
	require.True(t, member.CanInvite)
	require.True(t, member.CanCapture)

	// The creator can now reach the Circle via GetByID (active membership).
	fetched, err := repo.GetByID(context.Background(), created.ID, userID)
	require.NoError(t, err)
	require.Equal(t, "NgodingVareng", fetched.Name)
}

func TestCircleRepository_GetByID_NonMember_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	creatorID := seedTestUser(t, pool)
	outsiderID := seedTestUser(t, pool)
	repo := repository.NewCircleRepository(pool)

	created, err := repo.Create(context.Background(), &entity.Circle{Name: "Private"})
	require.NoError(t, err)
	_, err = repo.AddCreatorAsAdmin(context.Background(), created.ID, creatorID)
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), created.ID, outsiderID)
	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestCircleRepository_Update_NonAdmin_NotFound(t *testing.T) {
	// Update's WHERE clause requires the caller to be an active admin —
	// a plain member (or anyone else) attempting it is indistinguishable
	// from a wrong/foreign id, same "no row = 404" reasoning as
	// ThreadRepository.Update.
	pool := setupTestDB(t)
	adminID := seedTestUser(t, pool)
	memberID := seedTestUser(t, pool)
	repo := repository.NewCircleRepository(pool)

	created, err := repo.Create(context.Background(), &entity.Circle{Name: "Original"})
	require.NoError(t, err)
	_, err = repo.AddCreatorAsAdmin(context.Background(), created.ID, adminID)
	require.NoError(t, err)
	_, err = pool.Exec(context.Background(),
		`INSERT INTO circle_members (circle_id, user_id, role, can_invite, can_capture) VALUES ($1, $2, 'member', false, true)`,
		created.ID, memberID)
	require.NoError(t, err)

	_, err = repo.Update(context.Background(), &entity.Circle{ID: created.ID, Name: "Hijacked"}, memberID)
	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestCircleRepository_Update_Admin_Success(t *testing.T) {
	pool := setupTestDB(t)
	adminID := seedTestUser(t, pool)
	repo := repository.NewCircleRepository(pool)

	created, err := repo.Create(context.Background(), &entity.Circle{Name: "Original"})
	require.NoError(t, err)
	_, err = repo.AddCreatorAsAdmin(context.Background(), created.ID, adminID)
	require.NoError(t, err)

	newDescription := "Updated description"
	updated, err := repo.Update(context.Background(), &entity.Circle{
		ID: created.ID, Name: "Renamed", Description: &newDescription,
	}, adminID)

	require.NoError(t, err)
	require.Equal(t, "Renamed", updated.Name)
	require.Equal(t, newDescription, *updated.Description)
}

func TestCircleRepository_Dissolve_Admin_Success(t *testing.T) {
	pool := setupTestDB(t)
	adminID := seedTestUser(t, pool)
	repo := repository.NewCircleRepository(pool)

	created, err := repo.Create(context.Background(), &entity.Circle{Name: "To dissolve"})
	require.NoError(t, err)
	_, err = repo.AddCreatorAsAdmin(context.Background(), created.ID, adminID)
	require.NoError(t, err)

	err = repo.Dissolve(context.Background(), created.ID, adminID)
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), created.ID, adminID)
	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestCircleRepository_LeaveAndRemoveMember(t *testing.T) {
	pool := setupTestDB(t)
	adminID := seedTestUser(t, pool)
	memberID := seedTestUser(t, pool)
	repo := repository.NewCircleRepository(pool)

	created, err := repo.Create(context.Background(), &entity.Circle{Name: "Roster test"})
	require.NoError(t, err)
	_, err = repo.AddCreatorAsAdmin(context.Background(), created.ID, adminID)
	require.NoError(t, err)

	_, err = pool.Exec(context.Background(),
		`INSERT INTO circle_members (circle_id, user_id, role, can_invite, can_capture) VALUES ($1, $2, 'member', false, true)`,
		created.ID, memberID)
	require.NoError(t, err)

	members, err := repo.ListMembers(context.Background(), created.ID)
	require.NoError(t, err)
	require.Len(t, members, 2)

	// Self-leave.
	require.NoError(t, repo.Leave(context.Background(), created.ID, memberID))
	_, err = repo.GetActiveMember(context.Background(), created.ID, memberID)
	require.ErrorIs(t, err, errs.ErrNotFound)

	// Re-seat (the row still exists with left_at set — clear it directly
	// rather than re-inserting, which would collide with the PK) and
	// have the admin remove them instead.
	_, err = pool.Exec(context.Background(),
		`UPDATE circle_members SET left_at = NULL WHERE circle_id = $1 AND user_id = $2`,
		created.ID, memberID)
	require.NoError(t, err)
	require.NoError(t, repo.RemoveMember(context.Background(), created.ID, memberID, adminID))
	_, err = repo.GetActiveMember(context.Background(), created.ID, memberID)
	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestCircleRepository_UpdateMemberRole_PromotingForcesCanInvite(t *testing.T) {
	pool := setupTestDB(t)
	adminID := seedTestUser(t, pool)
	memberID := seedTestUser(t, pool)
	repo := repository.NewCircleRepository(pool)

	created, err := repo.Create(context.Background(), &entity.Circle{Name: "Promotion test"})
	require.NoError(t, err)
	_, err = repo.AddCreatorAsAdmin(context.Background(), created.ID, adminID)
	require.NoError(t, err)
	_, err = pool.Exec(context.Background(),
		`INSERT INTO circle_members (circle_id, user_id, role, can_invite, can_capture) VALUES ($1, $2, 'member', false, true)`,
		created.ID, memberID)
	require.NoError(t, err)

	promoted, err := repo.UpdateMemberRole(context.Background(), created.ID, memberID, adminID, enum.CircleRoleAdmin)
	require.NoError(t, err)
	require.Equal(t, enum.CircleRoleAdmin, promoted.Role)
	require.True(t, promoted.CanInvite, "chk_circle_members_admin_can_invite requires an admin to carry can_invite")
}

func TestCircleRepository_ListByUserID(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewCircleRepository(pool)

	created, err := repo.Create(context.Background(), &entity.Circle{Name: "Mine"})
	require.NoError(t, err)
	_, err = repo.AddCreatorAsAdmin(context.Background(), created.ID, userID)
	require.NoError(t, err)

	circles, err := repo.ListByUserID(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, circles, 1)
	require.Equal(t, created.ID, circles[0].ID)
}

func TestCircleRepository_DoUsersShareAnyCircle(t *testing.T) {
	pool := setupTestDB(t)
	userA := seedTestUser(t, pool)
	userB := seedTestUser(t, pool)
	repo := repository.NewCircleRepository(pool)

	shared, err := repo.DoUsersShareAnyCircle(context.Background(), userA, userB)
	require.NoError(t, err)
	require.False(t, shared)

	created, err := repo.Create(context.Background(), &entity.Circle{Name: "Shared"})
	require.NoError(t, err)
	_, err = repo.AddCreatorAsAdmin(context.Background(), created.ID, userA)
	require.NoError(t, err)
	_, err = pool.Exec(context.Background(),
		`INSERT INTO circle_members (circle_id, user_id, role, can_invite, can_capture) VALUES ($1, $2, 'member', false, true)`,
		created.ID, userB)
	require.NoError(t, err)

	shared, err = repo.DoUsersShareAnyCircle(context.Background(), userA, userB)
	require.NoError(t, err)
	require.True(t, shared)
}
