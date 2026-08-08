//go:build integration

package handler_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-openapi/testify/v2/require"
	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
)

func TestCircleInviteHandler_InviteDirect_ExistingDiscoverableUsername_AddedImmediately(t *testing.T) {
	// seedTestUser's row takes every column but name/username/email at
	// its DB default, and circle_invite_policy defaults to 'anyone' — so
	// a direct invite admits the target immediately, no accept step.
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)
	_, targetUsername := seedTestUserWithUsername(t, testApp.pool)

	circle := createCircle(t, testApp, adminAuth, "Family")

	resp := testApp.doRequest(t, http.MethodPost, "/circles/"+circle.ID+"/members/direct",
		dto.InviteDirectRequest{Usernames: []string{targetUsername}},
		map[string]string{"Authorization": adminAuth})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.InviteDirectResponse]](t, resp).Data
	require.Len(t, body.AddedMembers, 1)
	require.Empty(t, body.PendingInvites)
	require.Empty(t, body.SkippedUsernames)
	require.Equal(t, "member", body.AddedMembers[0].Role)
}

func TestCircleInviteHandler_InviteDirect_UnknownUsername_Skipped(t *testing.T) {
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)

	circle := createCircle(t, testApp, adminAuth, "Family")

	resp := testApp.doRequest(t, http.MethodPost, "/circles/"+circle.ID+"/members/direct",
		dto.InviteDirectRequest{Usernames: []string{"nobody_by_this_handle"}},
		map[string]string{"Authorization": adminAuth})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.InviteDirectResponse]](t, resp).Data
	require.Empty(t, body.AddedMembers)
	require.Empty(t, body.PendingInvites)
	require.Equal(t, []string{"nobody_by_this_handle"}, body.SkippedUsernames)
}

func TestCircleInviteHandler_InviteDirect_NonAnyonePolicy_PendingInvite(t *testing.T) {
	// A recipient whose circle_invite_policy isn't 'anyone' is held back
	// as a pending username invite instead of being admitted immediately
	// (FEATURES.md, Circle Invite).
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)
	targetID, targetUsername := seedTestUserWithUsername(t, testApp.pool)

	_, err := testApp.pool.Exec(context.Background(),
		`UPDATE users SET circle_invite_policy = 'known' WHERE id = $1`, targetID)
	require.NoError(t, err)

	circle := createCircle(t, testApp, adminAuth, "Family")

	resp := testApp.doRequest(t, http.MethodPost, "/circles/"+circle.ID+"/members/direct",
		dto.InviteDirectRequest{Usernames: []string{targetUsername}},
		map[string]string{"Authorization": adminAuth})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.InviteDirectResponse]](t, resp).Data
	require.Empty(t, body.AddedMembers)
	require.Empty(t, body.SkippedUsernames)
	require.Len(t, body.PendingInvites, 1)
	require.Equal(t, "username", body.PendingInvites[0].Kind)
	require.Equal(t, "pending", body.PendingInvites[0].Status)
	require.Equal(t, targetID.String(), *body.PendingInvites[0].InviteeUserID)
}

func TestCircleInviteHandler_InviteDirect_WithoutCanInvite_Forbidden(t *testing.T) {
	// AddCircleMembers seats direct-invite admits with can_invite = FALSE
	// by default (see AddCircleMembers's ON CONFLICT clause) — a plain
	// member can't invite until an admin grants it.
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)
	plainMemberID, plainMemberUsername := seedTestUserWithUsername(t, testApp.pool)
	_, thirdUsername := seedTestUserWithUsername(t, testApp.pool)

	circle := createCircle(t, testApp, adminAuth, "Family")
	addDirectMember(t, testApp, uuid.MustParse(circle.ID), adminAuth, plainMemberUsername)

	resp := testApp.doRequest(t, http.MethodPost, "/circles/"+circle.ID+"/members/direct",
		dto.InviteDirectRequest{Usernames: []string{thirdUsername}},
		map[string]string{"Authorization": testApp.authHeader(t, plainMemberID)})

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestCircleInviteHandler_ListMyPendingInvites_Success(t *testing.T) {
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)
	targetID, targetUsername := seedTestUserWithUsername(t, testApp.pool)

	_, err := testApp.pool.Exec(context.Background(),
		`UPDATE users SET circle_invite_policy = 'nobody' WHERE id = $1`, targetID)
	require.NoError(t, err)

	circle := createCircle(t, testApp, adminAuth, "Family")
	testApp.doRequest(t, http.MethodPost, "/circles/"+circle.ID+"/members/direct",
		dto.InviteDirectRequest{Usernames: []string{targetUsername}},
		map[string]string{"Authorization": adminAuth})

	resp := testApp.doRequest(t, http.MethodGet, "/circle-invites", nil,
		map[string]string{"Authorization": testApp.authHeader(t, targetID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	invites := decodeBody[dto.WebResponse[dto.ListCircleInvitesResponse]](t, resp).Data.Invites
	require.Len(t, invites, 1)
	require.Equal(t, circle.ID, invites[0].CircleID)
}

// seedPendingUsernameInvite is the common setup for Accept/Decline/Revoke
// below: an admin's circle, and a pending username invite to targetID
// because targetID's circle_invite_policy isn't 'anyone'.
func seedPendingUsernameInvite(t *testing.T, testApp *testApp) (circle dto.CircleResponse, adminAuth string, targetID uuid.UUID, inviteID string) {
	t.Helper()
	adminID := seedTestUser(t, testApp.pool)
	adminAuth = testApp.authHeader(t, adminID)
	targetID, targetUsername := seedTestUserWithUsername(t, testApp.pool)

	_, err := testApp.pool.Exec(context.Background(),
		`UPDATE users SET circle_invite_policy = 'known' WHERE id = $1`, targetID)
	require.NoError(t, err)

	circle = createCircle(t, testApp, adminAuth, "Family")
	inviteResp := testApp.doRequest(t, http.MethodPost, "/circles/"+circle.ID+"/members/direct",
		dto.InviteDirectRequest{Usernames: []string{targetUsername}},
		map[string]string{"Authorization": adminAuth})
	body := decodeBody[dto.WebResponse[dto.InviteDirectResponse]](t, inviteResp).Data
	require.Len(t, body.PendingInvites, 1)
	return circle, adminAuth, targetID, body.PendingInvites[0].ID
}

func TestCircleInviteHandler_AcceptInvite_Success(t *testing.T) {
	testApp := setupTestApp(t)
	circle, adminAuth, targetID, inviteID := seedPendingUsernameInvite(t, testApp)

	resp := testApp.doRequest(t, http.MethodPost, "/circle-invites/"+inviteID+"/accept", nil,
		map[string]string{"Authorization": testApp.authHeader(t, targetID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.CircleMemberResponse]](t, resp).Data
	require.Equal(t, targetID.String(), body.UserID)
	require.Equal(t, "member", body.Role)

	membersResp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID+"/members", nil,
		map[string]string{"Authorization": adminAuth})
	members := decodeBody[dto.WebResponse[dto.ListCircleMembersResponse]](t, membersResp).Data.Members
	require.Len(t, members, 2)
}

func TestCircleInviteHandler_DeclineInvite_Success(t *testing.T) {
	testApp := setupTestApp(t)
	_, _, targetID, inviteID := seedPendingUsernameInvite(t, testApp)

	resp := testApp.doRequest(t, http.MethodPost, "/circle-invites/"+inviteID+"/decline", nil,
		map[string]string{"Authorization": testApp.authHeader(t, targetID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.CircleInviteResponse]](t, resp).Data
	require.Equal(t, "declined", body.Status)
}

func TestCircleInviteHandler_RevokeInvite_Success(t *testing.T) {
	testApp := setupTestApp(t)
	circle, adminAuth, _, inviteID := seedPendingUsernameInvite(t, testApp)

	resp := testApp.doRequest(t, http.MethodPost,
		"/circles/"+circle.ID+"/invites/"+inviteID+"/revoke", nil,
		map[string]string{"Authorization": adminAuth})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.CircleInviteResponse]](t, resp).Data
	require.Equal(t, "revoked", body.Status)
}

func TestCircleInviteHandler_GetInviteLink_AnyActiveMember(t *testing.T) {
	// Viewing an existing link is open to any active member, not just
	// invite-privileged ones (CircleInviteUsecase.GetInviteLink's own
	// comment) — generating/rotating/toggling is the privileged action.
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)
	plainMemberID, plainMemberUsername := seedTestUserWithUsername(t, testApp.pool)

	circle := createCircle(t, testApp, adminAuth, "Family")
	addDirectMember(t, testApp, uuid.MustParse(circle.ID), adminAuth, plainMemberUsername)

	createResp := testApp.doRequest(t, http.MethodPost, "/circles/"+circle.ID+"/invite-link",
		dto.CreateOrRotateInviteLinkRequest{RequiresApproval: false},
		map[string]string{"Authorization": adminAuth})
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	resp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID+"/invite-link", nil,
		map[string]string{"Authorization": testApp.authHeader(t, plainMemberID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.CircleInviteResponse]](t, resp).Data
	require.Equal(t, "link", body.Kind)
	require.Equal(t, "active", body.Status)
}

func TestCircleInviteHandler_CreateOrRotateInviteLink_Success(t *testing.T) {
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)

	circle := createCircle(t, testApp, adminAuth, "Family")

	resp := testApp.doRequest(t, http.MethodPost, "/circles/"+circle.ID+"/invite-link",
		dto.CreateOrRotateInviteLinkRequest{RequiresApproval: true},
		map[string]string{"Authorization": adminAuth})

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.CircleInviteLinkResponse]](t, resp).Data
	require.Equal(t, "link", body.Invite.Kind)
	require.NotEmpty(t, body.Token)
	require.True(t, *body.Invite.RequiresApproval)
}

func TestCircleInviteHandler_CreateOrRotateInviteLink_WithoutCanInvite_Forbidden(t *testing.T) {
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)
	plainMemberID, plainMemberUsername := seedTestUserWithUsername(t, testApp.pool)

	circle := createCircle(t, testApp, adminAuth, "Family")
	addDirectMember(t, testApp, uuid.MustParse(circle.ID), adminAuth, plainMemberUsername)

	resp := testApp.doRequest(t, http.MethodPost, "/circles/"+circle.ID+"/invite-link",
		dto.CreateOrRotateInviteLinkRequest{RequiresApproval: false},
		map[string]string{"Authorization": testApp.authHeader(t, plainMemberID)})

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestCircleInviteHandler_SetInviteLinkRequiresApproval_Success(t *testing.T) {
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)

	circle := createCircle(t, testApp, adminAuth, "Family")
	testApp.doRequest(t, http.MethodPost, "/circles/"+circle.ID+"/invite-link",
		dto.CreateOrRotateInviteLinkRequest{RequiresApproval: false},
		map[string]string{"Authorization": adminAuth})

	resp := testApp.doRequest(t, http.MethodPatch, "/circles/"+circle.ID+"/invite-link/approval",
		dto.SetInviteLinkRequiresApprovalRequest{RequiresApproval: true},
		map[string]string{"Authorization": adminAuth})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.CircleInviteResponse]](t, resp).Data
	require.True(t, *body.RequiresApproval)
}
