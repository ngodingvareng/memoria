//go:build integration

package handler_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-openapi/testify/v2/require"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
)

// seedTestUserWithUsername mirrors seedTestUser but also hands back the
// generated username — several Circle Invite flows below act by
// username rather than id, so the plain uuid seedTestUser returns isn't
// enough on its own. Shared across all three circle_*_integration_test.go
// files in this package.
func seedTestUserWithUsername(t *testing.T, pool *pgxpool.Pool) (uuid.UUID, string) {
	t.Helper()
	username := testUsername()
	var id uuid.UUID
	err := pool.QueryRow(context.Background(),
		`INSERT INTO users (name, username, email) VALUES ($1, $2, $3) RETURNING id`,
		"Test User", username, uuid.NewString()+"@example.com",
	).Scan(&id)
	require.NoError(t, err)
	return id, username
}

// addDirectMember uses POST /circles/:id/members/direct (default
// circle_invite_policy is 'anyone', see users migration) to seat
// targetUsername as a plain member immediately, without going through
// the accept-invite flow. Shared helper for tests across this package
// that just need "a second member in the circle" as setup.
func addDirectMember(t *testing.T, testApp *testApp, circleID uuid.UUID, adminAuth string, targetUsername string) {
	t.Helper()
	resp := testApp.doRequest(t, http.MethodPost, "/circles/"+circleID.String()+"/members/direct",
		dto.InviteDirectRequest{Usernames: []string{targetUsername}},
		map[string]string{"Authorization": adminAuth})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.InviteDirectResponse]](t, resp)
	require.Len(t, body.Data.AddedMembers, 1)
}

func createCircle(t *testing.T, testApp *testApp, adminAuth string, name string) dto.CircleResponse {
	t.Helper()
	resp := testApp.doRequest(t, http.MethodPost, "/circles",
		dto.CreateCircleRequest{Name: name}, map[string]string{"Authorization": adminAuth})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	return decodeBody[dto.WebResponse[dto.CircleResponse]](t, resp).Data
}

func TestCircleHandler_CreateCircle_Success(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)
	auth := testApp.authHeader(t, userID)

	circle := createCircle(t, testApp, auth, "Family")
	require.Equal(t, "Family", circle.Name)

	// FEATURES.md, Circle Permissions: the creator becomes the circle's admin.
	resp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID+"/members", nil,
		map[string]string{"Authorization": auth})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	members := decodeBody[dto.WebResponse[dto.ListCircleMembersResponse]](t, resp).Data.Members
	require.Len(t, members, 1)
	require.Equal(t, userID.String(), members[0].UserID)
	require.Equal(t, "admin", members[0].Role)
	require.True(t, members[0].CanInvite)
}

func TestCircleHandler_UpdateCircle_Success(t *testing.T) {
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	auth := testApp.authHeader(t, adminID)

	circle := createCircle(t, testApp, auth, "Family")

	resp := testApp.doRequest(t, http.MethodPut, "/circles/"+circle.ID, dto.UpdateCircleRequest{
		Name: "Renamed Family",
	}, map[string]string{"Authorization": auth})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.CircleResponse]](t, resp)
	require.Equal(t, "Renamed Family", body.Data.Name)
}

func TestCircleHandler_UpdateCircle_NonAdmin_NotFound(t *testing.T) {
	// CircleRepository.Update enforces admin-only in its WHERE clause, so
	// a non-admin member's attempt is indistinguishable from a wrong id —
	// both surface as 404, never 403.
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)
	memberID, memberUsername := seedTestUserWithUsername(t, testApp.pool)

	circle := createCircle(t, testApp, adminAuth, "Family")
	addDirectMember(t, testApp, uuid.MustParse(circle.ID), adminAuth, memberUsername)

	resp := testApp.doRequest(t, http.MethodPut, "/circles/"+circle.ID, dto.UpdateCircleRequest{
		Name: "Hijacked",
	}, map[string]string{"Authorization": testApp.authHeader(t, memberID)})

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCircleHandler_UpdateCircle_WrongID_NotFound(t *testing.T) {
	// Same 404 as the non-admin case above — proves the two are genuinely
	// indistinguishable from the caller's point of view.
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)
	auth := testApp.authHeader(t, userID)

	resp := testApp.doRequest(t, http.MethodPut, "/circles/"+uuid.NewString(), dto.UpdateCircleRequest{
		Name: "Hijacked",
	}, map[string]string{"Authorization": auth})

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCircleHandler_DissolveCircle_Success(t *testing.T) {
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	auth := testApp.authHeader(t, adminID)

	circle := createCircle(t, testApp, auth, "Family")

	deleteResp := testApp.doRequest(t, http.MethodDelete, "/circles/"+circle.ID, nil,
		map[string]string{"Authorization": auth})
	require.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

	getResp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID, nil,
		map[string]string{"Authorization": auth})
	require.Equal(t, http.StatusNotFound, getResp.StatusCode)
}

func TestCircleHandler_DissolveCircle_NonAdmin_SilentNoop(t *testing.T) {
	// DissolveCircle is a sqlc :exec query — a WHERE clause matching zero
	// rows (here, because the actor isn't an admin) is a silent no-op,
	// not errs.ErrNotFound. The handler has no way to tell the two apart
	// from a real dissolve, so it still answers 204 despite the circle
	// staying alive.
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)
	memberID, memberUsername := seedTestUserWithUsername(t, testApp.pool)

	circle := createCircle(t, testApp, adminAuth, "Family")
	addDirectMember(t, testApp, uuid.MustParse(circle.ID), adminAuth, memberUsername)

	deleteResp := testApp.doRequest(t, http.MethodDelete, "/circles/"+circle.ID, nil,
		map[string]string{"Authorization": testApp.authHeader(t, memberID)})
	require.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

	getResp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID, nil,
		map[string]string{"Authorization": adminAuth})
	require.Equal(t, http.StatusOK, getResp.StatusCode)
}

func TestCircleHandler_GetCircle_ActiveMember_Success(t *testing.T) {
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	auth := testApp.authHeader(t, adminID)

	circle := createCircle(t, testApp, auth, "Family")

	resp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID, nil,
		map[string]string{"Authorization": auth})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.CircleResponse]](t, resp)
	require.Equal(t, circle.ID, body.Data.ID)
}

func TestCircleHandler_GetCircle_NonMember_NotFound(t *testing.T) {
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	outsiderID := seedTestUser(t, testApp.pool)
	auth := testApp.authHeader(t, adminID)

	circle := createCircle(t, testApp, auth, "Family")

	resp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID, nil,
		map[string]string{"Authorization": testApp.authHeader(t, outsiderID)})
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCircleHandler_ListMyCircles_OnlyActiveMemberships(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)
	otherID := seedTestUser(t, testApp.pool)
	auth := testApp.authHeader(t, userID)

	mine := createCircle(t, testApp, auth, "Mine")
	createCircle(t, testApp, testApp.authHeader(t, otherID), "Not mine")

	resp := testApp.doRequest(t, http.MethodGet, "/circles", nil, map[string]string{"Authorization": auth})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	circles := decodeBody[dto.WebResponse[dto.ListCirclesResponse]](t, resp).Data.Circles
	require.Len(t, circles, 1)
	require.Equal(t, mine.ID, circles[0].ID)
}

func TestCircleHandler_ListMyCircles_ExcludesLeftCircles(t *testing.T) {
	// ListCirclesByUserID filters on circle_members.left_at IS NULL, so a
	// circle the user has since left must drop out of "my circles" even
	// though the membership row still exists.
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)
	memberID, memberUsername := seedTestUserWithUsername(t, testApp.pool)
	memberAuth := testApp.authHeader(t, memberID)

	circle := createCircle(t, testApp, adminAuth, "Family")
	addDirectMember(t, testApp, uuid.MustParse(circle.ID), adminAuth, memberUsername)

	leaveResp := testApp.doRequest(t, http.MethodDelete,
		"/circles/"+circle.ID+"/members/"+memberID.String(), nil,
		map[string]string{"Authorization": memberAuth})
	require.Equal(t, http.StatusNoContent, leaveResp.StatusCode)

	resp := testApp.doRequest(t, http.MethodGet, "/circles", nil, map[string]string{"Authorization": memberAuth})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	circles := decodeBody[dto.WebResponse[dto.ListCirclesResponse]](t, resp).Data.Circles
	require.Empty(t, circles)
}

func TestCircleHandler_ListMembers_Success(t *testing.T) {
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)
	_, memberUsername := seedTestUserWithUsername(t, testApp.pool)

	circle := createCircle(t, testApp, adminAuth, "Family")
	addDirectMember(t, testApp, uuid.MustParse(circle.ID), adminAuth, memberUsername)

	resp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID+"/members", nil,
		map[string]string{"Authorization": adminAuth})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	members := decodeBody[dto.WebResponse[dto.ListCircleMembersResponse]](t, resp).Data.Members
	require.Len(t, members, 2)
}

func TestCircleHandler_RemoveMember_SelfLeave(t *testing.T) {
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)
	memberID, memberUsername := seedTestUserWithUsername(t, testApp.pool)

	circle := createCircle(t, testApp, adminAuth, "Family")
	addDirectMember(t, testApp, uuid.MustParse(circle.ID), adminAuth, memberUsername)

	// userId in the path equals the caller's own id — CircleHandler.RemoveMember
	// routes this to LeaveCircle, not the admin-forced RemoveMember path.
	resp := testApp.doRequest(t, http.MethodDelete,
		"/circles/"+circle.ID+"/members/"+memberID.String(), nil,
		map[string]string{"Authorization": testApp.authHeader(t, memberID)})
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	membersResp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID+"/members", nil,
		map[string]string{"Authorization": adminAuth})
	members := decodeBody[dto.WebResponse[dto.ListCircleMembersResponse]](t, membersResp).Data.Members
	require.Len(t, members, 1)
	require.Equal(t, adminID.String(), members[0].UserID)
}

func TestCircleHandler_RemoveMember_SelfLeave_SoleAdmin_Rejected(t *testing.T) {
	// A Circle's only active admin self-leaving would orphan it — no
	// one left could manage it, its invites, or its join requests.
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)

	circle := createCircle(t, testApp, adminAuth, "Solo")

	resp := testApp.doRequest(t, http.MethodDelete,
		"/circles/"+circle.ID+"/members/"+adminID.String(), nil,
		map[string]string{"Authorization": adminAuth})
	require.Equal(t, http.StatusConflict, resp.StatusCode)

	membersResp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID+"/members", nil,
		map[string]string{"Authorization": adminAuth})
	members := decodeBody[dto.WebResponse[dto.ListCircleMembersResponse]](t, membersResp).Data.Members
	require.Len(t, members, 1, "the sole admin must still be seated — the leave was rejected")
}

func TestCircleHandler_RemoveMember_AdminForcedRemoval(t *testing.T) {
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)
	memberID, memberUsername := seedTestUserWithUsername(t, testApp.pool)

	circle := createCircle(t, testApp, adminAuth, "Family")
	addDirectMember(t, testApp, uuid.MustParse(circle.ID), adminAuth, memberUsername)

	resp := testApp.doRequest(t, http.MethodDelete,
		"/circles/"+circle.ID+"/members/"+memberID.String(), nil,
		map[string]string{"Authorization": adminAuth})
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	membersResp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID+"/members", nil,
		map[string]string{"Authorization": adminAuth})
	members := decodeBody[dto.WebResponse[dto.ListCircleMembersResponse]](t, membersResp).Data.Members
	require.Len(t, members, 1)
	require.Equal(t, adminID.String(), members[0].UserID)
}

func TestCircleHandler_RemoveMember_NonAdminRemovesOther_SilentNoop(t *testing.T) {
	// RemoveCircleMember requires the actor to be an active admin, checked
	// in the same statement as the target row — a non-admin's attempt is a
	// silent no-op (204, nothing removed), same reasoning as
	// DissolveCircle_NonAdmin.
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)
	targetID, targetUsername := seedTestUserWithUsername(t, testApp.pool)
	removerID, removerUsername := seedTestUserWithUsername(t, testApp.pool)

	circle := createCircle(t, testApp, adminAuth, "Family")
	addDirectMember(t, testApp, uuid.MustParse(circle.ID), adminAuth, targetUsername)
	addDirectMember(t, testApp, uuid.MustParse(circle.ID), adminAuth, removerUsername)

	resp := testApp.doRequest(t, http.MethodDelete,
		"/circles/"+circle.ID+"/members/"+targetID.String(), nil,
		map[string]string{"Authorization": testApp.authHeader(t, removerID)})
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	membersResp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID+"/members", nil,
		map[string]string{"Authorization": adminAuth})
	members := decodeBody[dto.WebResponse[dto.ListCircleMembersResponse]](t, membersResp).Data.Members
	require.Len(t, members, 3, "target must still be seated — the removal was a no-op")
}

func TestCircleHandler_UpdateMemberRole_AdminOnly_Success(t *testing.T) {
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)
	memberID, memberUsername := seedTestUserWithUsername(t, testApp.pool)

	circle := createCircle(t, testApp, adminAuth, "Family")
	addDirectMember(t, testApp, uuid.MustParse(circle.ID), adminAuth, memberUsername)

	resp := testApp.doRequest(t, http.MethodPatch,
		"/circles/"+circle.ID+"/members/"+memberID.String()+"/role",
		dto.UpdateCircleMemberRoleRequest{Role: "admin"},
		map[string]string{"Authorization": adminAuth})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.CircleMemberResponse]](t, resp)
	require.Equal(t, "admin", body.Data.Role)
	// chk_circle_members_admin_can_invite: promoting to admin forces
	// can_invite = TRUE in the same statement.
	require.True(t, body.Data.CanInvite)
}

func TestCircleHandler_UpdateMemberRole_DemoteSoleAdmin_Rejected(t *testing.T) {
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)

	circle := createCircle(t, testApp, adminAuth, "Solo")

	resp := testApp.doRequest(t, http.MethodPatch,
		"/circles/"+circle.ID+"/members/"+adminID.String()+"/role",
		dto.UpdateCircleMemberRoleRequest{Role: "member"},
		map[string]string{"Authorization": adminAuth})

	require.Equal(t, http.StatusConflict, resp.StatusCode)

	membersResp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID+"/members", nil,
		map[string]string{"Authorization": adminAuth})
	members := decodeBody[dto.WebResponse[dto.ListCircleMembersResponse]](t, membersResp).Data.Members
	require.Equal(t, "admin", members[0].Role, "the sole admin's role must be unchanged")
}

func TestCircleHandler_UpdateMemberRole_NonAdmin_NotFound(t *testing.T) {
	// UpdateCircleMemberRole is a sqlc :one query whose WHERE clause
	// requires the actor to be an active admin — zero rows maps to
	// errs.ErrNotFound (404), not 403, same "no row = 404" pattern as
	// UpdateCircle.
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)
	targetID, targetUsername := seedTestUserWithUsername(t, testApp.pool)
	actorID, actorUsername := seedTestUserWithUsername(t, testApp.pool)

	circle := createCircle(t, testApp, adminAuth, "Family")
	addDirectMember(t, testApp, uuid.MustParse(circle.ID), adminAuth, targetUsername)
	addDirectMember(t, testApp, uuid.MustParse(circle.ID), adminAuth, actorUsername)

	resp := testApp.doRequest(t, http.MethodPatch,
		"/circles/"+circle.ID+"/members/"+targetID.String()+"/role",
		dto.UpdateCircleMemberRoleRequest{Role: "admin"},
		map[string]string{"Authorization": testApp.authHeader(t, actorID)})

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCircleHandler_UpdateMemberPermissions_AdminOnly_Success(t *testing.T) {
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)
	memberID, memberUsername := seedTestUserWithUsername(t, testApp.pool)

	circle := createCircle(t, testApp, adminAuth, "Family")
	addDirectMember(t, testApp, uuid.MustParse(circle.ID), adminAuth, memberUsername)

	resp := testApp.doRequest(t, http.MethodPatch,
		"/circles/"+circle.ID+"/members/"+memberID.String(),
		dto.UpdateCircleMemberPermissionsRequest{CanInvite: true, CanCapture: false},
		map[string]string{"Authorization": adminAuth})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.CircleMemberResponse]](t, resp)
	require.True(t, body.Data.CanInvite)
	require.False(t, body.Data.CanCapture)
}

func TestCircleHandler_UpdateMemberPermissions_NonAdmin_NotFound(t *testing.T) {
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)
	targetID, targetUsername := seedTestUserWithUsername(t, testApp.pool)
	actorID, actorUsername := seedTestUserWithUsername(t, testApp.pool)

	circle := createCircle(t, testApp, adminAuth, "Family")
	addDirectMember(t, testApp, uuid.MustParse(circle.ID), adminAuth, targetUsername)
	addDirectMember(t, testApp, uuid.MustParse(circle.ID), adminAuth, actorUsername)

	resp := testApp.doRequest(t, http.MethodPatch,
		"/circles/"+circle.ID+"/members/"+targetID.String(),
		dto.UpdateCircleMemberPermissionsRequest{CanInvite: true, CanCapture: true},
		map[string]string{"Authorization": testApp.authHeader(t, actorID)})

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}
