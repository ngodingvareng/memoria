//go:build integration

package handler_test

import (
	"net/http"
	"testing"

	"github.com/go-openapi/testify/v2/require"

	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
)

func TestSearchHandler_Suggestions_MatchesThreadsAndMoments(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	thread := decodeBody[dto.WebResponse[dto.ThreadResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/threads", dto.CreateThreadRequest{Name: "Morning workout"},
			map[string]string{"Authorization": testApp.authHeader(t, userID)})).Data

	note := "Great gym session"
	moment := decodeBody[dto.WebResponse[dto.MomentResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
			OccurredAt: "2026-08-04T14:30:00+09:00",
			Note:       &note,
		}, map[string]string{"Authorization": testApp.authHeader(t, userID)})).Data

	threadResp := testApp.doRequest(t, http.MethodGet, "/search/suggestions?q=morning", nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})
	require.Equal(t, http.StatusOK, threadResp.StatusCode)
	threadBody := decodeBody[dto.WebResponse[dto.SearchSuggestionsResponse]](t, threadResp)
	require.Len(t, threadBody.Data.Threads, 1)
	require.Equal(t, thread.ID, threadBody.Data.Threads[0].ID)
	require.Empty(t, threadBody.Data.Moments)

	momentResp := testApp.doRequest(t, http.MethodGet, "/search/suggestions?q=gym", nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})
	require.Equal(t, http.StatusOK, momentResp.StatusCode)
	momentBody := decodeBody[dto.WebResponse[dto.SearchSuggestionsResponse]](t, momentResp)
	require.Len(t, momentBody.Data.Moments, 1)
	require.Equal(t, moment.ID, momentBody.Data.Moments[0].ID)
	require.Empty(t, momentBody.Data.Threads)
}

func TestSearchHandler_Suggestions_TrigramMatchesTypo(t *testing.T) {
	// pg_trgm similarity fallback: "grocry" (typo) must still surface
	// "Grocery run" via the `name % query` predicate, not just exact
	// prefix tsquery matching.
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	thread := decodeBody[dto.WebResponse[dto.ThreadResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/threads", dto.CreateThreadRequest{Name: "Grocery run"},
			map[string]string{"Authorization": testApp.authHeader(t, userID)})).Data

	resp := testApp.doRequest(t, http.MethodGet, "/search/suggestions?q=grocry", nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.SearchSuggestionsResponse]](t, resp)
	require.Len(t, body.Data.Threads, 1)
	require.Equal(t, thread.ID, body.Data.Threads[0].ID)
}

func TestSearchHandler_Suggestions_ScopedToOwnData(t *testing.T) {
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	outsiderID := seedTestUser(t, testApp.pool)

	testApp.doRequest(t, http.MethodPost, "/threads", dto.CreateThreadRequest{Name: "Private planning"},
		map[string]string{"Authorization": testApp.authHeader(t, ownerID)})

	note := "Private planning notes"
	testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
		OccurredAt: "2026-08-04T14:30:00+09:00",
		Note:       &note,
	}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)})

	resp := testApp.doRequest(t, http.MethodGet, "/search/suggestions?q=planning", nil,
		map[string]string{"Authorization": testApp.authHeader(t, outsiderID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.SearchSuggestionsResponse]](t, resp)
	require.Empty(t, body.Data.Threads)
	require.Empty(t, body.Data.Moments)
}

func TestSearchHandler_Suggestions_MissingQuery_BadRequest(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	resp := testApp.doRequest(t, http.MethodGet, "/search/suggestions", nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
