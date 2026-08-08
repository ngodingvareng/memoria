//go:build integration

package handler_test

import (
	"net/http"
	"testing"

	"github.com/go-openapi/testify/v2/require"

	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
)

// createTestMoment (moment_image_handler_integration_test.go) and
// seedTestUserWithUsername (circle_handler_integration_test.go) are
// shared helpers already defined elsewhere in this package.

func TestMentionHandler_CreateMention_Success(t *testing.T) {
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	_, targetUsername := seedTestUserWithUsername(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	resp := testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/mentions", dto.CreateMentionRequest{
		Username: targetUsername,
	}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.MentionResponse]](t, resp)
	require.Equal(t, moment.ID, body.Data.MomentID)
}

func TestMentionHandler_CreateMention_NotOwner_NotFound(t *testing.T) {
	// MentionUsecase.CreateMention checks moment ownership via
	// MomentRepository.GetByID, which returns errs.ErrNotFound (404) when
	// id/userID don't match any row — a non-owner is indistinguishable
	// from a wrong id.
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	outsiderID := seedTestUser(t, testApp.pool)
	_, targetUsername := seedTestUserWithUsername(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	resp := testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/mentions", dto.CreateMentionRequest{
		Username: targetUsername,
	}, map[string]string{"Authorization": testApp.authHeader(t, outsiderID)})

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestMentionHandler_CreateMention_UnknownUsername_NotFound(t *testing.T) {
	// UserRepository.GetByUsername surfaces errs.ErrNotFound (404) when
	// no user matches.
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	resp := testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/mentions", dto.CreateMentionRequest{
		Username: "no_such_user_at_all",
	}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestMentionHandler_ListMentions_OwnerOnly(t *testing.T) {
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	_, targetUsername := seedTestUserWithUsername(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/mentions", dto.CreateMentionRequest{
		Username: targetUsername,
	}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})

	resp := testApp.doRequest(t, http.MethodGet, "/moments/"+moment.ID+"/mentions", nil,
		map[string]string{"Authorization": testApp.authHeader(t, ownerID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.ListMentionsResponse]](t, resp)
	require.Len(t, body.Data.Mentions, 1)
}

func TestMentionHandler_ListMentions_NotOwner_NotFound(t *testing.T) {
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	outsiderID := seedTestUser(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	resp := testApp.doRequest(t, http.MethodGet, "/moments/"+moment.ID+"/mentions", nil,
		map[string]string{"Authorization": testApp.authHeader(t, outsiderID)})

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestMentionHandler_DeleteMention_Owner_Success(t *testing.T) {
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	_, targetUsername := seedTestUserWithUsername(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	mention := decodeBody[dto.WebResponse[dto.MentionResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/mentions", dto.CreateMentionRequest{
			Username: targetUsername,
		}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})).Data

	resp := testApp.doRequest(t, http.MethodDelete, "/moments/"+moment.ID+"/mentions/"+mention.ID, nil,
		map[string]string{"Authorization": testApp.authHeader(t, ownerID)})

	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	listResp := testApp.doRequest(t, http.MethodGet, "/moments/"+moment.ID+"/mentions", nil,
		map[string]string{"Authorization": testApp.authHeader(t, ownerID)})
	listBody := decodeBody[dto.WebResponse[dto.ListMentionsResponse]](t, listResp)
	require.Empty(t, listBody.Data.Mentions)
}

func TestMentionHandler_LeaveMention_MentionedUser_Success(t *testing.T) {
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	mentionedID, mentionedUsername := seedTestUserWithUsername(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/mentions", dto.CreateMentionRequest{
		Username: mentionedUsername,
	}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})

	resp := testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/mentions/leave", nil,
		map[string]string{"Authorization": testApp.authHeader(t, mentionedID)})

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestMentionHandler_LeaveMention_NonExistent_SilentNoOp(t *testing.T) {
	// MentionRepository.Remove is a :exec query — a WHERE clause matching
	// zero rows (never mentioned) is a silent no-op, still 204.
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	outsiderID := seedTestUser(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	resp := testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/mentions/leave", nil,
		map[string]string{"Authorization": testApp.authHeader(t, outsiderID)})

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestMentionHandler_ListMentionedMoments_Success(t *testing.T) {
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	mentionedID, mentionedUsername := seedTestUserWithUsername(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/mentions", dto.CreateMentionRequest{
		Username: mentionedUsername,
	}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})

	resp := testApp.doRequest(t, http.MethodGet, "/mentions", nil,
		map[string]string{"Authorization": testApp.authHeader(t, mentionedID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.ListMomentsResponse]](t, resp)
	require.Len(t, body.Data.Moments, 1)
	require.Equal(t, moment.ID, body.Data.Moments[0].ID)
}

func TestMentionHandler_ShareToCircle_Success(t *testing.T) {
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	circle := decodeBody[dto.WebResponse[dto.CircleResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/circles", dto.CreateCircleRequest{Name: "Family"},
			map[string]string{"Authorization": testApp.authHeader(t, ownerID)})).Data

	resp := testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/share", dto.ShareMomentToCircleRequest{
		CircleID: circle.ID,
	}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.MomentCircleResponse]](t, resp)
	require.Equal(t, circle.ID, body.Data.CircleID)
}

func TestMentionHandler_ShareToCircle_NotMember_NotFound(t *testing.T) {
	// ShareToCircle requires the caller to be an active member of the
	// target Circle (CircleRepository.GetActiveMember -> errs.ErrNotFound
	// when not a member).
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	otherID := seedTestUser(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	// A Circle the Moment owner does not belong to.
	circle := decodeBody[dto.WebResponse[dto.CircleResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/circles", dto.CreateCircleRequest{Name: "Not mine"},
			map[string]string{"Authorization": testApp.authHeader(t, otherID)})).Data

	resp := testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/share", dto.ShareMomentToCircleRequest{
		CircleID: circle.ID,
	}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestMentionHandler_UnshareFromCircle_Success(t *testing.T) {
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	circle := decodeBody[dto.WebResponse[dto.CircleResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/circles", dto.CreateCircleRequest{Name: "Family"},
			map[string]string{"Authorization": testApp.authHeader(t, ownerID)})).Data

	testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/share", dto.ShareMomentToCircleRequest{
		CircleID: circle.ID,
	}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})

	resp := testApp.doRequest(t, http.MethodDelete, "/moments/"+moment.ID+"/share/"+circle.ID, nil,
		map[string]string{"Authorization": testApp.authHeader(t, ownerID)})

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}
