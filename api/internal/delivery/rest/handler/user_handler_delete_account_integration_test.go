//go:build integration

package handler_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/go-openapi/testify/v2/require"
	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
)

// TestUserHandler_DeleteAccount_FullCascade exercises every effect
// FEATURES.md's Lifecycle & Deletion entry for account deletion
// describes: own Moments/Threads/photos gone, comments/reactions on
// other people's Moments anonymized (not deleted), a sole-admin Circle
// with other members gets a promoted successor, a sole-admin Circle the
// user was alone in gets dissolved, and every session is revoked.
func TestUserHandler_DeleteAccount_FullCascade(t *testing.T) {
	testApp := setupTestApp(t)
	ctx := context.Background()

	userID := seedTestUser(t, testApp.pool)
	auth := map[string]string{"Authorization": testApp.authHeader(t, userID)}
	otherUserID, otherUsername := seedTestUserWithUsername(t, testApp.pool)

	// A live session, to prove RevokeAllByUserID actually runs.
	var refreshTokenID uuid.UUID
	require.NoError(t, testApp.pool.QueryRow(ctx,
		`INSERT INTO refresh_tokens (user_id, family_id, token_hash, expires_at) VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, uuid.New(), "test-hash-"+uuid.NewString(), time.Now().Add(24*time.Hour),
	).Scan(&refreshTokenID))

	// Thread + Moment owned by the deleting user, created through the
	// real API. Image rows are seeded directly — this test is about
	// cascade correctness, not the multipart upload endpoint.
	threadResp := testApp.doRequest(t, http.MethodPost, "/threads",
		dto.CreateThreadRequest{Name: "Morning Run"}, auth)
	require.Equal(t, http.StatusCreated, threadResp.StatusCode)
	thread := decodeBody[dto.WebResponse[dto.ThreadResponse]](t, threadResp).Data

	momentResp := testApp.doRequest(t, http.MethodPost, "/moments",
		dto.CreateMomentRequest{ThreadID: &thread.ID, OccurredAt: time.Now().Format(time.RFC3339)}, auth)
	require.Equal(t, http.StatusCreated, momentResp.StatusCode)
	moment := decodeBody[dto.WebResponse[dto.MomentResponse]](t, momentResp).Data

	_, err := testApp.pool.Exec(ctx,
		`INSERT INTO moment_images (moment_id, image_path) VALUES ($1, $2)`, moment.ID, "moments/test/a.jpg")
	require.NoError(t, err)
	_, err = testApp.pool.Exec(ctx,
		`INSERT INTO thread_images (thread_id, image_path) VALUES ($1, $2)`, thread.ID, "threads/test/b.jpg")
	require.NoError(t, err)

	// A Moment owned by someone else, with a comment and a reaction from
	// the deleting user — must survive, only their authorship anonymizes.
	var otherMomentID uuid.UUID
	require.NoError(t, testApp.pool.QueryRow(ctx,
		`INSERT INTO moments (user_id, occurred_at, occurred_local) VALUES ($1, NOW(), NOW()) RETURNING id`,
		otherUserID,
	).Scan(&otherMomentID))
	var commentID uuid.UUID
	require.NoError(t, testApp.pool.QueryRow(ctx,
		`INSERT INTO comments (moment_id, user_id, body) VALUES ($1, $2, $3) RETURNING id`,
		otherMomentID, userID, "Great moment!",
	).Scan(&commentID))
	_, err = testApp.pool.Exec(ctx,
		`INSERT INTO reactions (moment_id, user_id, kind) VALUES ($1, $2, 'heart')`, otherMomentID, userID)
	require.NoError(t, err)

	// Circle 1: deleting user is sole admin, one other active member —
	// must auto-promote them rather than error (LeaveCircle's default
	// behavior, which DeleteAccount deliberately does not reuse as-is).
	circle1 := createCircle(t, testApp, auth["Authorization"], "Promote Me")
	addDirectMember(t, testApp, uuid.MustParse(circle1.ID), auth["Authorization"], otherUsername)

	// Circle 2: deleting user is sole admin AND the only member — must
	// dissolve instead, since there is no one left to hand it to.
	circle2 := createCircle(t, testApp, auth["Authorization"], "Dissolve Me")

	// The actual deletion.
	delResp := testApp.doRequest(t, http.MethodDelete, "/users/me",
		dto.DeleteAccountRequest{Confirmation: "DELETE"}, auth)
	require.Equal(t, http.StatusNoContent, delResp.StatusCode)

	// The account is unreachable from here on — requireOnboarded's
	// GetByID lookup no longer finds the (now soft-deleted) row and
	// maps that to the same 401 an invalid token would get, not a 404,
	// so a deleted account is indistinguishable from "never logged in."
	meResp := testApp.doRequest(t, http.MethodGet, "/users/me", nil, auth)
	require.Equal(t, http.StatusUnauthorized, meResp.StatusCode)

	// Own Moment/Thread soft-deleted.
	var momentDeletedAt *time.Time
	require.NoError(t, testApp.pool.QueryRow(ctx,
		`SELECT deleted_at FROM moments WHERE id = $1`, moment.ID).Scan(&momentDeletedAt))
	require.NotNil(t, momentDeletedAt)
	var threadDeletedAt *time.Time
	require.NoError(t, testApp.pool.QueryRow(ctx,
		`SELECT deleted_at FROM threads WHERE id = $1`, thread.ID).Scan(&threadDeletedAt))
	require.NotNil(t, threadDeletedAt)

	// Comment/reaction on the OTHER user's Moment survive, anonymized.
	var commentUserID *uuid.UUID
	require.NoError(t, testApp.pool.QueryRow(ctx,
		`SELECT user_id FROM comments WHERE id = $1`, commentID).Scan(&commentUserID))
	require.Nil(t, commentUserID)
	var reactionUserID *uuid.UUID
	require.NoError(t, testApp.pool.QueryRow(ctx,
		`SELECT user_id FROM reactions WHERE moment_id = $1`, otherMomentID).Scan(&reactionUserID))
	require.Nil(t, reactionUserID)
	// The other user's Moment itself is completely untouched.
	var otherMomentDeletedAt *time.Time
	require.NoError(t, testApp.pool.QueryRow(ctx,
		`SELECT deleted_at FROM moments WHERE id = $1`, otherMomentID).Scan(&otherMomentDeletedAt))
	require.Nil(t, otherMomentDeletedAt)

	// Circle 1: successor promoted to admin, deleting user's own
	// membership ended, the Circle itself survives.
	var successorRole string
	var successorLeftAt *time.Time
	require.NoError(t, testApp.pool.QueryRow(ctx,
		`SELECT role, left_at FROM circle_members WHERE circle_id = $1 AND user_id = $2`,
		circle1.ID, otherUserID,
	).Scan(&successorRole, &successorLeftAt))
	require.Equal(t, "admin", successorRole)
	require.Nil(t, successorLeftAt)

	var deletedUserLeftAt *time.Time
	require.NoError(t, testApp.pool.QueryRow(ctx,
		`SELECT left_at FROM circle_members WHERE circle_id = $1 AND user_id = $2`,
		circle1.ID, userID,
	).Scan(&deletedUserLeftAt))
	require.NotNil(t, deletedUserLeftAt)

	var circle1DissolvedAt *time.Time
	require.NoError(t, testApp.pool.QueryRow(ctx,
		`SELECT dissolved_at FROM circles WHERE id = $1`, circle1.ID).Scan(&circle1DissolvedAt))
	require.Nil(t, circle1DissolvedAt)

	// Circle 2: dissolved, since no other member existed to hand it to.
	var circle2DissolvedAt *time.Time
	require.NoError(t, testApp.pool.QueryRow(ctx,
		`SELECT dissolved_at FROM circles WHERE id = $1`, circle2.ID).Scan(&circle2DissolvedAt))
	require.NotNil(t, circle2DissolvedAt)

	// The session is revoked.
	var revokedAt *time.Time
	require.NoError(t, testApp.pool.QueryRow(ctx,
		`SELECT revoked_at FROM refresh_tokens WHERE id = $1`, refreshTokenID).Scan(&revokedAt))
	require.NotNil(t, revokedAt)
}

func TestUserHandler_DeleteAccount_WrongConfirmation_Rejected(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	resp := testApp.doRequest(t, http.MethodDelete, "/users/me",
		dto.DeleteAccountRequest{Confirmation: "delete"},
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// The account must still be fully intact.
	var deletedAt *time.Time
	require.NoError(t, testApp.pool.QueryRow(context.Background(),
		`SELECT deleted_at FROM users WHERE id = $1`, userID).Scan(&deletedAt))
	require.Nil(t, deletedAt)
}

func TestUserHandler_DeleteAccount_Unauthenticated(t *testing.T) {
	testApp := setupTestApp(t)

	resp := testApp.doRequest(t, http.MethodDelete, "/users/me",
		dto.DeleteAccountRequest{Confirmation: "DELETE"}, nil)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
