//go:build integration

package handler_test

import (
	"net/http"
	"testing"

	"github.com/go-openapi/testify/v2/require"
	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
)

// createTestMoment, seedTestUserWithUsername, createCircle, and
// addDirectMember are shared helpers already defined elsewhere in this
// package (moment_image_handler_integration_test.go and
// circle_handler_integration_test.go respectively).

func TestCommentHandler_CreateComment_Owner_MentionContext_Success(t *testing.T) {
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	resp := testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/comments", dto.CreateCommentRequest{
		Body: "Owner commenting on their own moment",
	}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.CommentResponse]](t, resp)
	require.Nil(t, body.Data.CircleID)
}

func TestCommentHandler_CreateComment_ActiveMention_MentionContext_Success(t *testing.T) {
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	mentionedID, mentionedUsername := seedTestUserWithUsername(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/mentions", dto.CreateMentionRequest{
		Username: mentionedUsername,
	}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})

	resp := testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/comments", dto.CreateCommentRequest{
		Body: "The mentioned user chiming in",
	}, map[string]string{"Authorization": testApp.authHeader(t, mentionedID)})

	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestCommentHandler_CreateComment_UnrelatedUser_MentionContext_NotFound(t *testing.T) {
	// A viewer with zero path to the Moment (not owner, no mention, no
	// Circle share, no Thread membership) makes MomentReader.GetWithAccess
	// return errs.ErrNotFound before ResolveAudience ever reaches the
	// mention-eligibility check — 404, not 403.
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	strangerID := seedTestUser(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	resp := testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/comments", dto.CreateCommentRequest{
		Body: "I have no relationship to this moment",
	}, map[string]string{"Authorization": testApp.authHeader(t, strangerID)})

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCommentHandler_CreateComment_CircleMemberNotMentioned_MentionContext_Forbidden(t *testing.T) {
	// A Circle member the Moment was shared with CAN reach it
	// (GetWithAccess succeeds via the Circle-share path), but the mention
	// context (circle_id omitted) is only for the owner or an actively
	// mentioned user (ResponseAccessChecker.ResolveAudience) — a Circle
	// member who is neither gets requireAudience's errs.ErrForbidden, 403.
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	ownerAuth := testApp.authHeader(t, ownerID)
	memberID, memberUsername := seedTestUserWithUsername(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)
	circle := createCircle(t, testApp, ownerAuth, "Test circle")
	addDirectMember(t, testApp, uuid.MustParse(circle.ID), ownerAuth, memberUsername)

	shareResp := testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/share", dto.ShareMomentToCircleRequest{
		CircleID: circle.ID,
	}, map[string]string{"Authorization": ownerAuth})
	require.Equal(t, http.StatusOK, shareResp.StatusCode)

	resp := testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/comments", dto.CreateCommentRequest{
		Body: "Trying to comment in the mention context",
	}, map[string]string{"Authorization": testApp.authHeader(t, memberID)})

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestCommentHandler_CreateComment_CircleContext_Success(t *testing.T) {
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	ownerAuth := testApp.authHeader(t, ownerID)
	memberID, memberUsername := seedTestUserWithUsername(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)
	circle := createCircle(t, testApp, ownerAuth, "Test circle")
	addDirectMember(t, testApp, uuid.MustParse(circle.ID), ownerAuth, memberUsername)

	shareResp := testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/share", dto.ShareMomentToCircleRequest{
		CircleID: circle.ID,
	}, map[string]string{"Authorization": ownerAuth})
	require.Equal(t, http.StatusOK, shareResp.StatusCode)

	circleID := circle.ID
	resp := testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/comments", dto.CreateCommentRequest{
		CircleID: &circleID,
		Body:     "Commenting in the shared circle",
	}, map[string]string{"Authorization": testApp.authHeader(t, memberID)})

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.CommentResponse]](t, resp)
	require.NotNil(t, body.Data.CircleID)
	require.Equal(t, circle.ID, *body.Data.CircleID)
}

func TestCommentHandler_CreateComment_CircleContext_NotMember_Forbidden(t *testing.T) {
	// mentionedID has SOME access to the Moment (an active mention, so
	// GetWithAccess succeeds and ResolveAudience runs to completion) but
	// is not a member of the Circle the comment names — circleID isn't in
	// audience.CircleIDs, so requireAudience returns errs.ErrForbidden.
	// A user with zero access at all would 404 instead (see the
	// mention-context unrelated-user test above) — this test deliberately
	// picks a viewer GetWithAccess *can* reach to exercise the 403 path.
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	ownerAuth := testApp.authHeader(t, ownerID)
	_, circleMemberUsername := seedTestUserWithUsername(t, testApp.pool)
	mentionedID, mentionedUsername := seedTestUserWithUsername(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)
	circle := createCircle(t, testApp, ownerAuth, "Test circle")
	addDirectMember(t, testApp, uuid.MustParse(circle.ID), ownerAuth, circleMemberUsername)

	shareResp := testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/share", dto.ShareMomentToCircleRequest{
		CircleID: circle.ID,
	}, map[string]string{"Authorization": ownerAuth})
	require.Equal(t, http.StatusOK, shareResp.StatusCode)

	testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/mentions", dto.CreateMentionRequest{
		Username: mentionedUsername,
	}, map[string]string{"Authorization": ownerAuth})

	circleID := circle.ID
	resp := testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/comments", dto.CreateCommentRequest{
		CircleID: &circleID,
		Body:     "Trying to comment in a circle I'm not part of",
	}, map[string]string{"Authorization": testApp.authHeader(t, mentionedID)})

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestCommentHandler_ListComments_Success(t *testing.T) {
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/comments", dto.CreateCommentRequest{
		Body: "First comment",
	}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})

	resp := testApp.doRequest(t, http.MethodGet, "/moments/"+moment.ID+"/comments", nil,
		map[string]string{"Authorization": testApp.authHeader(t, ownerID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.ListCommentsResponse]](t, resp)
	require.Len(t, body.Data.Comments, 1)
	require.Equal(t, "First comment", body.Data.Comments[0].Body)
}

func TestCommentHandler_UpdateComment_Author_Success(t *testing.T) {
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	comment := decodeBody[dto.WebResponse[dto.CommentResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/comments", dto.CreateCommentRequest{
			Body: "Original body",
		}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})).Data

	resp := testApp.doRequest(t, http.MethodPut, "/moments/"+moment.ID+"/comments/"+comment.ID, dto.UpdateCommentRequest{
		Body: "Edited body",
	}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.CommentResponse]](t, resp)
	require.Equal(t, "Edited body", body.Data.Body)
}

func TestCommentHandler_UpdateComment_NotAuthor_NotFound(t *testing.T) {
	// CommentRepository.Update scopes its WHERE clause to id AND user_id
	// — a wrong author is indistinguishable from a wrong id, both
	// errs.ErrNotFound (404).
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	otherID := seedTestUser(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	comment := decodeBody[dto.WebResponse[dto.CommentResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/comments", dto.CreateCommentRequest{
			Body: "Original body",
		}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})).Data

	resp := testApp.doRequest(t, http.MethodPut, "/moments/"+moment.ID+"/comments/"+comment.ID, dto.UpdateCommentRequest{
		Body: "Hijacked body",
	}, map[string]string{"Authorization": testApp.authHeader(t, otherID)})

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCommentHandler_DeleteComment_Author_Success(t *testing.T) {
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	comment := decodeBody[dto.WebResponse[dto.CommentResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/comments", dto.CreateCommentRequest{
			Body: "To be deleted",
		}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})).Data

	resp := testApp.doRequest(t, http.MethodDelete, "/moments/"+moment.ID+"/comments/"+comment.ID, nil,
		map[string]string{"Authorization": testApp.authHeader(t, ownerID)})
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	listResp := testApp.doRequest(t, http.MethodGet, "/moments/"+moment.ID+"/comments", nil,
		map[string]string{"Authorization": testApp.authHeader(t, ownerID)})
	listBody := decodeBody[dto.WebResponse[dto.ListCommentsResponse]](t, listResp)
	require.Empty(t, listBody.Data.Comments)
}

func TestCommentHandler_DeleteComment_NotAuthor_SilentNoOp(t *testing.T) {
	// CommentRepository.SoftDelete is a :exec query — a WHERE clause
	// matching zero rows (wrong author) is a silent no-op: still 204, and
	// the comment stays undeleted.
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	otherID := seedTestUser(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	comment := decodeBody[dto.WebResponse[dto.CommentResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/comments", dto.CreateCommentRequest{
			Body: "Not going anywhere",
		}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})).Data

	resp := testApp.doRequest(t, http.MethodDelete, "/moments/"+moment.ID+"/comments/"+comment.ID, nil,
		map[string]string{"Authorization": testApp.authHeader(t, otherID)})
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	listResp := testApp.doRequest(t, http.MethodGet, "/moments/"+moment.ID+"/comments", nil,
		map[string]string{"Authorization": testApp.authHeader(t, ownerID)})
	listBody := decodeBody[dto.WebResponse[dto.ListCommentsResponse]](t, listResp)
	require.Len(t, listBody.Data.Comments, 1)
}
