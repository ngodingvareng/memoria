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

func TestReactionHandler_SetReaction_MentionContext_Success(t *testing.T) {
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	resp := testApp.doRequest(t, http.MethodPut, "/moments/"+moment.ID+"/reactions", dto.SetReactionRequest{
		Kind: "heart",
	}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.ReactionResponse]](t, resp)
	require.Equal(t, "heart", body.Data.Kind)
	require.Nil(t, body.Data.CircleID)
}

func TestReactionHandler_SetReaction_ChangeKind_UpdatesInPlace(t *testing.T) {
	// SetReaction has no separate update endpoint — reacting again with a
	// different kind on the same Moment/context just changes the existing
	// row (reactions.UpsertReaction's ON CONFLICT ... DO UPDATE).
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	first := decodeBody[dto.WebResponse[dto.ReactionResponse]](t,
		testApp.doRequest(t, http.MethodPut, "/moments/"+moment.ID+"/reactions", dto.SetReactionRequest{
			Kind: "heart",
		}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})).Data

	second := decodeBody[dto.WebResponse[dto.ReactionResponse]](t,
		testApp.doRequest(t, http.MethodPut, "/moments/"+moment.ID+"/reactions", dto.SetReactionRequest{
			Kind: "joy",
		}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})).Data

	require.Equal(t, first.ID, second.ID)
	require.Equal(t, "joy", second.Kind)

	listResp := testApp.doRequest(t, http.MethodGet, "/moments/"+moment.ID+"/reactions", nil,
		map[string]string{"Authorization": testApp.authHeader(t, ownerID)})
	listBody := decodeBody[dto.WebResponse[dto.ListReactionsResponse]](t, listResp)
	require.Len(t, listBody.Data.Reactions, 1)
}

func TestReactionHandler_SetReaction_CircleContext_Success(t *testing.T) {
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
	resp := testApp.doRequest(t, http.MethodPut, "/moments/"+moment.ID+"/reactions", dto.SetReactionRequest{
		CircleID: &circleID,
		Kind:     "tender",
	}, map[string]string{"Authorization": testApp.authHeader(t, memberID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.ReactionResponse]](t, resp)
	require.NotNil(t, body.Data.CircleID)
	require.Equal(t, circle.ID, *body.Data.CircleID)
}

func TestReactionHandler_SetReaction_UnrelatedUser_NotFound(t *testing.T) {
	// Same access model as CommentUsecase (both share
	// ResponseAccessChecker): a viewer with zero path to the Moment gets
	// errs.ErrNotFound from MomentReader.GetWithAccess, before the
	// mention/circle-eligibility check ever runs.
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	strangerID := seedTestUser(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	resp := testApp.doRequest(t, http.MethodPut, "/moments/"+moment.ID+"/reactions", dto.SetReactionRequest{
		Kind: "heart",
	}, map[string]string{"Authorization": testApp.authHeader(t, strangerID)})

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestReactionHandler_SetReaction_CircleContext_NotMember_Forbidden(t *testing.T) {
	// Mirrors TestCommentHandler_CreateComment_CircleContext_NotMember_Forbidden:
	// mentionedID reaches the Moment via an active mention (GetWithAccess
	// succeeds) but isn't a member of the named Circle, so it's absent
	// from audience.CircleIDs — requireAudience returns errs.ErrForbidden.
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
	resp := testApp.doRequest(t, http.MethodPut, "/moments/"+moment.ID+"/reactions", dto.SetReactionRequest{
		CircleID: &circleID,
		Kind:     "heart",
	}, map[string]string{"Authorization": testApp.authHeader(t, mentionedID)})

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestReactionHandler_RemoveReaction_MentionContext_Success(t *testing.T) {
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	testApp.doRequest(t, http.MethodPut, "/moments/"+moment.ID+"/reactions", dto.SetReactionRequest{
		Kind: "heart",
	}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})

	resp := testApp.doRequest(t, http.MethodDelete, "/moments/"+moment.ID+"/reactions", nil,
		map[string]string{"Authorization": testApp.authHeader(t, ownerID)})
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	listResp := testApp.doRequest(t, http.MethodGet, "/moments/"+moment.ID+"/reactions", nil,
		map[string]string{"Authorization": testApp.authHeader(t, ownerID)})
	listBody := decodeBody[dto.WebResponse[dto.ListReactionsResponse]](t, listResp)
	require.Empty(t, listBody.Data.Reactions)
}

func TestReactionHandler_RemoveReaction_Idempotent(t *testing.T) {
	// ReactionRepository.Delete is a :exec query — a WHERE clause
	// matching zero rows (already removed, never reacted) is a silent
	// no-op, so removing twice is still 204 the second time.
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	firstResp := testApp.doRequest(t, http.MethodDelete, "/moments/"+moment.ID+"/reactions", nil,
		map[string]string{"Authorization": testApp.authHeader(t, ownerID)})
	require.Equal(t, http.StatusNoContent, firstResp.StatusCode)

	secondResp := testApp.doRequest(t, http.MethodDelete, "/moments/"+moment.ID+"/reactions", nil,
		map[string]string{"Authorization": testApp.authHeader(t, ownerID)})
	require.Equal(t, http.StatusNoContent, secondResp.StatusCode)
}

func TestReactionHandler_RemoveReaction_CircleContext_Success(t *testing.T) {
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
	testApp.doRequest(t, http.MethodPut, "/moments/"+moment.ID+"/reactions", dto.SetReactionRequest{
		CircleID: &circleID,
		Kind:     "joy",
	}, map[string]string{"Authorization": testApp.authHeader(t, memberID)})

	resp := testApp.doRequest(t, http.MethodDelete, "/moments/"+moment.ID+"/reactions?circle_id="+circle.ID, nil,
		map[string]string{"Authorization": testApp.authHeader(t, memberID)})

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestReactionHandler_ListReactions_NoCountExposed(t *testing.T) {
	// FEATURES.md, Response: reaction counts are never displayed as a
	// number — dto.ListReactionsResponse only carries the rows
	// themselves, so the only way to know "how many" is len(Reactions);
	// there is no separate count field to read instead.
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	otherID, otherUsername := seedTestUserWithUsername(t, testApp.pool)

	moment := createTestMoment(t, testApp, ownerID)

	testApp.doRequest(t, http.MethodPut, "/moments/"+moment.ID+"/reactions", dto.SetReactionRequest{
		Kind: "heart",
	}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})

	testApp.doRequest(t, http.MethodPost, "/moments/"+moment.ID+"/mentions", dto.CreateMentionRequest{
		Username: otherUsername,
	}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})
	testApp.doRequest(t, http.MethodPut, "/moments/"+moment.ID+"/reactions", dto.SetReactionRequest{
		Kind: "joy",
	}, map[string]string{"Authorization": testApp.authHeader(t, otherID)})

	resp := testApp.doRequest(t, http.MethodGet, "/moments/"+moment.ID+"/reactions", nil,
		map[string]string{"Authorization": testApp.authHeader(t, ownerID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.ListReactionsResponse]](t, resp)
	require.Len(t, body.Data.Reactions, 2)
}
