//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/go-openapi/testify/v2/require"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/repository"
)

func TestUserPrivacyRepository_IsBlockedEitherDirection(t *testing.T) {
	pool := setupTestDB(t)
	userA := seedTestUser(t, pool)
	userB := seedTestUser(t, pool)
	repo := repository.NewUserPrivacyRepository(pool)
	ctx := context.Background()

	blocked, err := repo.IsBlockedEitherDirection(ctx, userA, userB)
	require.NoError(t, err)
	require.False(t, blocked)

	_, err = pool.Exec(ctx,
		`INSERT INTO user_blocks (blocker_user_id, blocked_user_id) VALUES ($1, $2)`,
		userA, userB)
	require.NoError(t, err)

	// Only userA's row exists (blocking stores one row per pair), but the
	// check must catch both argument orderings — that's the entire point
	// of "either direction".
	blocked, err = repo.IsBlockedEitherDirection(ctx, userA, userB)
	require.NoError(t, err)
	require.True(t, blocked)

	blocked, err = repo.IsBlockedEitherDirection(ctx, userB, userA)
	require.NoError(t, err)
	require.True(t, blocked)
}

func TestUserPrivacyRepository_IsKnownTo_DirectMark(t *testing.T) {
	pool := setupTestDB(t)
	owner := seedTestUser(t, pool)
	other := seedTestUser(t, pool)
	repo := repository.NewUserPrivacyRepository(pool)
	ctx := context.Background()

	known, err := repo.IsKnownTo(ctx, owner, other)
	require.NoError(t, err)
	require.False(t, known)

	_, err = pool.Exec(ctx,
		`INSERT INTO user_knows (knower_user_id, known_user_id) VALUES ($1, $2)`,
		owner, other)
	require.NoError(t, err)

	known, err = repo.IsKnownTo(ctx, owner, other)
	require.NoError(t, err)
	require.True(t, known)

	// Argument order is not symmetric: other marking owner as known
	// would grant other nothing, since only the owner's outgoing mark
	// counts for the owner's own policy.
	known, err = repo.IsKnownTo(ctx, other, owner)
	require.NoError(t, err)
	require.False(t, known)
}

func TestUserPrivacyRepository_IsKnownTo_ViaSharedActiveCircle(t *testing.T) {
	pool := setupTestDB(t)
	owner := seedTestUser(t, pool)
	other := seedTestUser(t, pool)
	circleRepo := repository.NewCircleRepository(pool)
	repo := repository.NewUserPrivacyRepository(pool)
	ctx := context.Background()

	circle, err := circleRepo.Create(ctx, &entity.Circle{Name: "Shared circle"})
	require.NoError(t, err)
	_, err = circleRepo.AddCreatorAsAdmin(ctx, circle.ID, owner)
	require.NoError(t, err)
	_, err = pool.Exec(ctx,
		`INSERT INTO circle_members (circle_id, user_id, role, can_invite, can_capture) VALUES ($1, $2, 'member', false, true)`,
		circle.ID, other)
	require.NoError(t, err)

	// Neither marked the other known directly, but sharing an active
	// Circle is "known" on its own — the load-bearing half of the union
	// that keeps the known-tier from deadlocking (see the user_knows
	// migration comment). True in both directions since circle
	// membership itself isn't directional.
	known, err := repo.IsKnownTo(ctx, owner, other)
	require.NoError(t, err)
	require.True(t, known)
	known, err = repo.IsKnownTo(ctx, other, owner)
	require.NoError(t, err)
	require.True(t, known)

	// other leaves the Circle: the shared-active-membership half no
	// longer applies, and nothing else grants access.
	_, err = pool.Exec(ctx,
		`UPDATE circle_members SET left_at = NOW() WHERE circle_id = $1 AND user_id = $2`,
		circle.ID, other)
	require.NoError(t, err)

	known, err = repo.IsKnownTo(ctx, owner, other)
	require.NoError(t, err)
	require.False(t, known)
}

func TestUserPrivacyRepository_IsMuted(t *testing.T) {
	pool := setupTestDB(t)
	muter := seedTestUser(t, pool)
	muted := seedTestUser(t, pool)
	repo := repository.NewUserPrivacyRepository(pool)
	ctx := context.Background()

	isMuted, err := repo.IsMuted(ctx, muter, muted)
	require.NoError(t, err)
	require.False(t, isMuted)

	_, err = pool.Exec(ctx,
		`INSERT INTO user_mutes (muter_user_id, muted_user_id) VALUES ($1, $2)`,
		muter, muted)
	require.NoError(t, err)

	isMuted, err = repo.IsMuted(ctx, muter, muted)
	require.NoError(t, err)
	require.True(t, isMuted)

	// One-directional: the muted user has not muted the muter back, so
	// checking the reverse pair must not pick up the same row.
	isMuted, err = repo.IsMuted(ctx, muted, muter)
	require.NoError(t, err)
	require.False(t, isMuted)
}
