//go:build integration

package handler_test

import (
	"net/http"
	"testing"

	"github.com/go-openapi/testify/v2/require"
	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
)

// createInviteLink creates (or rotates) circleID's invite link and hands
// back the raw token — the one time it's ever exposed
// (dto.CircleInviteLinkResponse's own comment).
func createInviteLink(t *testing.T, testApp *testApp, circleID, adminAuth string, requiresApproval bool) string {
	t.Helper()
	resp := testApp.doRequest(t, http.MethodPost, "/circles/"+circleID+"/invite-link",
		dto.CreateOrRotateInviteLinkRequest{RequiresApproval: requiresApproval},
		map[string]string{"Authorization": adminAuth})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	return decodeBody[dto.WebResponse[dto.CircleInviteLinkResponse]](t, resp).Data.Token
}

func TestCircleJoinRequestHandler_FollowInviteLink_NoApprovalRequired_ImmediateJoin(t *testing.T) {
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)
	followerID := seedTestUser(t, testApp.pool)

	circle := createCircle(t, testApp, adminAuth, "Family")
	token := createInviteLink(t, testApp, circle.ID, adminAuth, false)

	resp := testApp.doRequest(t, http.MethodPost, "/circle-join-requests",
		dto.FollowInviteLinkRequest{Token: token},
		map[string]string{"Authorization": testApp.authHeader(t, followerID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.FollowInviteLinkResponse]](t, resp).Data
	require.Nil(t, body.JoinRequest)
	require.NotNil(t, body.Member)
	require.Equal(t, followerID.String(), body.Member.UserID)
	require.Equal(t, "member", body.Member.Role)

	membersResp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID+"/members", nil,
		map[string]string{"Authorization": adminAuth})
	members := decodeBody[dto.WebResponse[dto.ListCircleMembersResponse]](t, membersResp).Data.Members
	require.Len(t, members, 2)
}

func TestCircleJoinRequestHandler_FollowInviteLink_RequiresApproval_PendingJoinRequest(t *testing.T) {
	testApp := setupTestApp(t)
	adminID := seedTestUser(t, testApp.pool)
	adminAuth := testApp.authHeader(t, adminID)
	followerID := seedTestUser(t, testApp.pool)

	circle := createCircle(t, testApp, adminAuth, "Family")
	token := createInviteLink(t, testApp, circle.ID, adminAuth, true)

	resp := testApp.doRequest(t, http.MethodPost, "/circle-join-requests",
		dto.FollowInviteLinkRequest{Token: token},
		map[string]string{"Authorization": testApp.authHeader(t, followerID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.FollowInviteLinkResponse]](t, resp).Data
	require.Nil(t, body.Member)
	require.NotNil(t, body.JoinRequest)
	require.Equal(t, "pending", body.JoinRequest.Status)
	require.Equal(t, followerID.String(), body.JoinRequest.UserID)

	// Not seated yet — following a requires-approval link never admits by itself.
	membersResp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID+"/members", nil,
		map[string]string{"Authorization": adminAuth})
	members := decodeBody[dto.WebResponse[dto.ListCircleMembersResponse]](t, membersResp).Data.Members
	require.Len(t, members, 1)
}

func TestCircleJoinRequestHandler_FollowInviteLink_InvalidToken_NotFound(t *testing.T) {
	// GetActiveLinkByTokenHash returns errs.ErrNotFound for a token with
	// no matching active link — a made-up token is indistinguishable from
	// a revoked/expired one.
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	resp := testApp.doRequest(t, http.MethodPost, "/circle-join-requests",
		dto.FollowInviteLinkRequest{Token: "not-a-real-token-" + uuid.NewString()},
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// seedPendingJoinRequest sets up an admin's circle with a requires-approval
// link, has followerID follow it into a pending join request, and returns
// everything the Approve/Reject/Cancel tests need.
func seedPendingJoinRequest(t *testing.T, testApp *testApp) (circle dto.CircleResponse, adminAuth string, followerID uuid.UUID, requestID string) {
	t.Helper()
	adminID := seedTestUser(t, testApp.pool)
	adminAuth = testApp.authHeader(t, adminID)
	followerID = seedTestUser(t, testApp.pool)

	circle = createCircle(t, testApp, adminAuth, "Family")
	token := createInviteLink(t, testApp, circle.ID, adminAuth, true)

	resp := testApp.doRequest(t, http.MethodPost, "/circle-join-requests",
		dto.FollowInviteLinkRequest{Token: token},
		map[string]string{"Authorization": testApp.authHeader(t, followerID)})
	body := decodeBody[dto.WebResponse[dto.FollowInviteLinkResponse]](t, resp).Data
	require.NotNil(t, body.JoinRequest)
	return circle, adminAuth, followerID, body.JoinRequest.ID
}

func TestCircleJoinRequestHandler_ListPending_RequiresCanInvite(t *testing.T) {
	testApp := setupTestApp(t)
	circle, adminAuth, _, _ := seedPendingJoinRequest(t, testApp)
	plainMemberID, plainMemberUsername := seedTestUserWithUsername(t, testApp.pool)
	addDirectMember(t, testApp, uuid.MustParse(circle.ID), adminAuth, plainMemberUsername)

	forbiddenResp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID+"/join-requests", nil,
		map[string]string{"Authorization": testApp.authHeader(t, plainMemberID)})
	require.Equal(t, http.StatusForbidden, forbiddenResp.StatusCode)

	resp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID+"/join-requests", nil,
		map[string]string{"Authorization": adminAuth})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	requests := decodeBody[dto.WebResponse[dto.ListCircleJoinRequestsResponse]](t, resp).Data.JoinRequests
	require.Len(t, requests, 1)
}

func TestCircleJoinRequestHandler_Approve_Success(t *testing.T) {
	testApp := setupTestApp(t)
	circle, adminAuth, followerID, requestID := seedPendingJoinRequest(t, testApp)

	resp := testApp.doRequest(t, http.MethodPost,
		"/circles/"+circle.ID+"/join-requests/"+requestID+"/approve", nil,
		map[string]string{"Authorization": adminAuth})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.CircleMemberResponse]](t, resp).Data
	require.Equal(t, followerID.String(), body.UserID)
	require.Equal(t, "member", body.Role)

	membersResp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID+"/members", nil,
		map[string]string{"Authorization": adminAuth})
	members := decodeBody[dto.WebResponse[dto.ListCircleMembersResponse]](t, membersResp).Data.Members
	require.Len(t, members, 2)

	pendingResp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID+"/join-requests", nil,
		map[string]string{"Authorization": adminAuth})
	pending := decodeBody[dto.WebResponse[dto.ListCircleJoinRequestsResponse]](t, pendingResp).Data.JoinRequests
	require.Empty(t, pending, "an approved request must drop out of the pending queue")
}

func TestCircleJoinRequestHandler_Reject_Success(t *testing.T) {
	testApp := setupTestApp(t)
	circle, adminAuth, _, requestID := seedPendingJoinRequest(t, testApp)

	resp := testApp.doRequest(t, http.MethodPost,
		"/circles/"+circle.ID+"/join-requests/"+requestID+"/reject", nil,
		map[string]string{"Authorization": adminAuth})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.CircleJoinRequestResponse]](t, resp).Data
	require.Equal(t, "rejected", body.Status)

	membersResp := testApp.doRequest(t, http.MethodGet, "/circles/"+circle.ID+"/members", nil,
		map[string]string{"Authorization": adminAuth})
	members := decodeBody[dto.WebResponse[dto.ListCircleMembersResponse]](t, membersResp).Data.Members
	require.Len(t, members, 1, "a rejected request must not seat a membership")
}

func TestCircleJoinRequestHandler_Cancel_Success(t *testing.T) {
	testApp := setupTestApp(t)
	_, _, followerID, requestID := seedPendingJoinRequest(t, testApp)

	resp := testApp.doRequest(t, http.MethodPost, "/circle-join-requests/"+requestID+"/cancel", nil,
		map[string]string{"Authorization": testApp.authHeader(t, followerID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.CircleJoinRequestResponse]](t, resp).Data
	require.Equal(t, "cancelled", body.Status)
}

func TestCircleJoinRequestHandler_Cancel_WrongUser_NotFound(t *testing.T) {
	// CancelCircleJoinRequest's WHERE clause matches user_id = the
	// requester, so anyone else — even the circle's admin — cancelling on
	// someone else's behalf gets the same "no row = 404" treatment used
	// throughout, not a permission-flavored error.
	testApp := setupTestApp(t)
	_, adminAuth, _, requestID := seedPendingJoinRequest(t, testApp)

	resp := testApp.doRequest(t, http.MethodPost, "/circle-join-requests/"+requestID+"/cancel", nil,
		map[string]string{"Authorization": adminAuth})

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}
