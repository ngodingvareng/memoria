//go:build integration

package handler_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/go-openapi/testify/v2/require"
	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
)

func strPtr(s string) *string { return &s }

// createThread is the ownership-check dependency ListThreadMoments/
// CreateMoment/UpdateMoment need whenever a test attaches a Moment to a
// Thread.
func createThread(t *testing.T, testApp *testApp, userID uuid.UUID, name string) dto.ThreadResponse {
	t.Helper()
	created := decodeBody[dto.WebResponse[dto.ThreadResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/threads", dto.CreateThreadRequest{Name: name},
			map[string]string{"Authorization": testApp.authHeader(t, userID)}))
	return created.Data
}

func TestMomentHandler_CreateMoment_Success_NoThread(t *testing.T) {
	// Manual Moment with no Thread is the primary capture path per
	// FEATURES.md — ThreadID is entirely omitted, not an empty string.
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	resp := testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
		OccurredAt: "2026-08-04T14:30:00+09:00",
		Note:       strPtr("Morning coffee"),
		ColorHex:   strPtr("#ff0000"),
	}, map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.MomentResponse]](t, resp)
	require.Equal(t, "manual", body.Data.Origin)
	require.Equal(t, userID.String(), body.Data.UserID)
	require.Nil(t, body.Data.ThreadID)
	require.Equal(t, "Morning coffee", *body.Data.Note)
	require.Equal(t, "#ff0000", *body.Data.ColorHex)
}

func TestMomentHandler_CreateMoment_Success_WithThread(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	thread := createThread(t, testApp, userID, "Morning Run")

	resp := testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
		ThreadID:   &thread.ID,
		OccurredAt: "2026-08-04T14:30:00+09:00",
	}, map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.MomentResponse]](t, resp)
	require.NotNil(t, body.Data.ThreadID)
	require.Equal(t, thread.ID, *body.Data.ThreadID)
}

func TestMomentHandler_CreateMoment_ValidationError_MissingOccurredAt(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	resp := testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
		// OccurredAt omitted — required.
	}, map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body := decodeBody[dto.WebResponse[any]](t, resp)
	require.Equal(t, "Validation failed", body.Message)
	require.NotNil(t, body.Errors)
}

func TestMomentHandler_CreateMoment_ValidationError_MalformedOccurredAt(t *testing.T) {
	// The struct validator's "datetime" tag parameter is literally
	// "2006-01-02T15:04:05Z07:00" — the same layout string the handler
	// later feeds to time.Parse(time.RFC3339, ...) — so anything that
	// fails one fails the other identically, and the request never gets
	// past validation (h.validate.Struct runs before the handler's own
	// time.Parse call) to reach that second branch.
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	resp := testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
		OccurredAt: "not-a-date",
	}, map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body := decodeBody[dto.WebResponse[any]](t, resp)
	require.Equal(t, "Validation failed", body.Message)
}

func TestMomentHandler_CreateMoment_Unauthorized(t *testing.T) {
	testApp := setupTestApp(t)

	resp := testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
		OccurredAt: "2026-08-04T14:30:00+09:00",
	}, nil)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestMomentHandler_UpdateMoment_Success(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	created := decodeBody[dto.WebResponse[dto.MomentResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
			OccurredAt: "2026-08-04T14:30:00+09:00",
			Note:       strPtr("Original note"),
		}, map[string]string{"Authorization": testApp.authHeader(t, userID)}))

	resp := testApp.doRequest(t, http.MethodPut, "/moments/"+created.Data.ID, dto.UpdateMomentRequest{
		OccurredAt: "2026-08-04T15:00:00+09:00",
		Note:       strPtr("Updated note"),
		ColorHex:   strPtr("#00ff00"),
	}, map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.MomentResponse]](t, resp)
	require.Equal(t, "Updated note", *body.Data.Note)
	require.Equal(t, "#00ff00", *body.Data.ColorHex)
}

func TestMomentHandler_UpdateMoment_WrongOwner_NotFound(t *testing.T) {
	// MomentRepository.Update scopes its WHERE clause by user_id and
	// returns errs.ErrNotFound on zero rows matched — same shape as
	// ThreadRepository.Update — so a non-owner gets 404, never 403.
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	outsiderID := seedTestUser(t, testApp.pool)

	created := decodeBody[dto.WebResponse[dto.MomentResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
			OccurredAt: "2026-08-04T14:30:00+09:00",
		}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)}))

	resp := testApp.doRequest(t, http.MethodPut, "/moments/"+created.Data.ID, dto.UpdateMomentRequest{
		OccurredAt: "2026-08-04T15:00:00+09:00",
		Note:       strPtr("Hijacked"),
	}, map[string]string{"Authorization": testApp.authHeader(t, outsiderID)})

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestMomentHandler_GetMoment_Success(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	created := decodeBody[dto.WebResponse[dto.MomentResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
			OccurredAt: "2026-08-04T14:30:00+09:00",
		}, map[string]string{"Authorization": testApp.authHeader(t, userID)}))

	resp := testApp.doRequest(t, http.MethodGet, "/moments/"+created.Data.ID, nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.MomentResponse]](t, resp)
	require.Equal(t, created.Data.ID, body.Data.ID)
}

func TestMomentHandler_GetMoment_WrongOwner_NotFound(t *testing.T) {
	// MomentRepository.GetByID returns errs.ErrNotFound if id/userID
	// don't match any row — a wrong owner is indistinguishable from a
	// wrong id, both 404.
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	outsiderID := seedTestUser(t, testApp.pool)

	created := decodeBody[dto.WebResponse[dto.MomentResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
			OccurredAt: "2026-08-04T14:30:00+09:00",
		}, map[string]string{"Authorization": testApp.authHeader(t, ownerID)}))

	resp := testApp.doRequest(t, http.MethodGet, "/moments/"+created.Data.ID, nil,
		map[string]string{"Authorization": testApp.authHeader(t, outsiderID)})

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestMomentHandler_DeleteMoment_ThenGet_NotFound(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	created := decodeBody[dto.WebResponse[dto.MomentResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
			OccurredAt: "2026-08-04T14:30:00+09:00",
		}, map[string]string{"Authorization": testApp.authHeader(t, userID)}))

	deleteResp := testApp.doRequest(t, http.MethodDelete, "/moments/"+created.Data.ID, nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})
	require.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

	getResp := testApp.doRequest(t, http.MethodGet, "/moments/"+created.Data.ID, nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})
	require.Equal(t, http.StatusNotFound, getResp.StatusCode)
}

func TestMomentHandler_ListMoments_FiltersByOwner(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)
	otherID := seedTestUser(t, testApp.pool)

	for i := range 3 {
		testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
			OccurredAt: "2026-08-04T14:30:00+09:00",
			Note:       strPtr(fmt.Sprintf("Moment %d", i)),
		}, map[string]string{"Authorization": testApp.authHeader(t, userID)})
	}
	testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
		OccurredAt: "2026-08-04T14:30:00+09:00",
		Note:       strPtr("Someone else's moment"),
	}, map[string]string{"Authorization": testApp.authHeader(t, otherID)})

	resp := testApp.doRequest(t, http.MethodGet, "/moments", nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.ListMomentsResponse]](t, resp)
	require.Len(t, body.Data.Moments, 3)
}

func TestMomentHandler_SearchMoments_FindsByNote(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
		OccurredAt: "2026-08-04T14:30:00+09:00",
		Note:       strPtr("A very distinctive zebra note"),
	}, map[string]string{"Authorization": testApp.authHeader(t, userID)})
	testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
		OccurredAt: "2026-08-04T14:30:00+09:00",
		Note:       strPtr("Unrelated note"),
	}, map[string]string{"Authorization": testApp.authHeader(t, userID)})

	resp := testApp.doRequest(t, http.MethodGet, "/moments/search?q=zebra", nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.ListMomentsResponse]](t, resp)
	require.Len(t, body.Data.Moments, 1)
	require.Contains(t, *body.Data.Moments[0].Note, "zebra")
}

func TestMomentHandler_SearchMoments_MissingQuery_ValidationError(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	resp := testApp.doRequest(t, http.MethodGet, "/moments/search", nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body := decodeBody[dto.WebResponse[any]](t, resp)
	require.Equal(t, "Validation failed", body.Message)
}

func TestMomentHandler_ListThreadMoments_FiltersByThread(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	thread := createThread(t, testApp, userID, "Morning Run")

	for i := range 2 {
		testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
			ThreadID:   &thread.ID,
			OccurredAt: "2026-08-04T14:30:00+09:00",
			Note:       strPtr(fmt.Sprintf("Run %d", i)),
		}, map[string]string{"Authorization": testApp.authHeader(t, userID)})
	}
	testApp.doRequest(t, http.MethodPost, "/moments", dto.CreateMomentRequest{
		OccurredAt: "2026-08-04T14:30:00+09:00",
		Note:       strPtr("Unattached moment"),
	}, map[string]string{"Authorization": testApp.authHeader(t, userID)})

	resp := testApp.doRequest(t, http.MethodGet, "/threads/"+thread.ID+"/moments", nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.ListMomentsResponse]](t, resp)
	require.Len(t, body.Data.Moments, 2)
	for _, m := range body.Data.Moments {
		require.NotNil(t, m.ThreadID)
		require.Equal(t, thread.ID, *m.ThreadID)
	}
}
