//go:build integration

package handler_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/go-openapi/testify/v2/require"

	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
)

func TestThreadHandler_CreateThread_Success(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	resp := testApp.doRequest(t, http.MethodPost, "/threads", dto.CreateThreadRequest{
		Name: "Morning workout",
	}, map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.ThreadResponse]](t, resp)
	require.Equal(t, "Morning workout", body.Data.Name)
	require.Equal(t, userID.String(), body.Data.UserID)
}

func TestThreadHandler_CreateThread_Unauthorized(t *testing.T) {
	// No Authorization header at all — RequireAuth must reject before
	// the handler/usecase ever runs.
	testApp := setupTestApp(t)

	resp := testApp.doRequest(t, http.MethodPost, "/threads", dto.CreateThreadRequest{
		Name: "Morning workout",
	}, nil)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestThreadHandler_CreateThread_ValidationError(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	resp := testApp.doRequest(t, http.MethodPost, "/threads", dto.CreateThreadRequest{
		Name: "", // required,min=1
	}, map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body := decodeBody[dto.WebResponse[any]](t, resp)
	require.Equal(t, "Validation failed", body.Message)
	require.NotNil(t, body.Errors)
}

func TestThreadHandler_GetThread_Success(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	created := decodeBody[dto.WebResponse[dto.ThreadResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/threads", dto.CreateThreadRequest{Name: "Morning workout"},
			map[string]string{"Authorization": testApp.authHeader(t, userID)}))

	resp := testApp.doRequest(t, http.MethodGet, "/threads/"+created.Data.ID, nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.ThreadResponse]](t, resp)
	require.Equal(t, created.Data.ID, body.Data.ID)
}

func TestThreadHandler_GetThread_WrongOwner_NotFound(t *testing.T) {
	// GetByID's ownership check (WHERE ... AND user_id = ?) makes a
	// non-owner indistinguishable from a wrong id — both are 404, never
	// 403 (see thread_usecase.go's GetByID comment).
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	outsiderID := seedTestUser(t, testApp.pool)

	created := decodeBody[dto.WebResponse[dto.ThreadResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/threads", dto.CreateThreadRequest{Name: "Private thread"},
			map[string]string{"Authorization": testApp.authHeader(t, ownerID)}))

	resp := testApp.doRequest(t, http.MethodGet, "/threads/"+created.Data.ID, nil,
		map[string]string{"Authorization": testApp.authHeader(t, outsiderID)})

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestThreadHandler_UpdateThread_WrongOwner_NotFound(t *testing.T) {
	testApp := setupTestApp(t)
	ownerID := seedTestUser(t, testApp.pool)
	outsiderID := seedTestUser(t, testApp.pool)

	created := decodeBody[dto.WebResponse[dto.ThreadResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/threads", dto.CreateThreadRequest{Name: "Private thread"},
			map[string]string{"Authorization": testApp.authHeader(t, ownerID)}))

	resp := testApp.doRequest(t, http.MethodPut, "/threads/"+created.Data.ID, dto.UpdateThreadRequest{
		Name: "Hijacked",
	}, map[string]string{"Authorization": testApp.authHeader(t, outsiderID)})

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestThreadHandler_DeleteThread_ThenGet_NotFound(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	created := decodeBody[dto.WebResponse[dto.ThreadResponse]](t,
		testApp.doRequest(t, http.MethodPost, "/threads", dto.CreateThreadRequest{Name: "Morning workout"},
			map[string]string{"Authorization": testApp.authHeader(t, userID)}))

	deleteResp := testApp.doRequest(t, http.MethodDelete, "/threads/"+created.Data.ID, nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})
	require.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

	getResp := testApp.doRequest(t, http.MethodGet, "/threads/"+created.Data.ID, nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})
	require.Equal(t, http.StatusNotFound, getResp.StatusCode)
}

func TestThreadHandler_SearchThreads_FiltersByOwner(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)
	otherID := seedTestUser(t, testApp.pool)

	for i := range 3 {
		testApp.doRequest(t, http.MethodPost, "/threads",
			dto.CreateThreadRequest{Name: fmt.Sprintf("Thread %d", i)},
			map[string]string{"Authorization": testApp.authHeader(t, userID)})
	}
	testApp.doRequest(t, http.MethodPost, "/threads", dto.CreateThreadRequest{Name: "Someone else's thread"},
		map[string]string{"Authorization": testApp.authHeader(t, otherID)})

	resp := testApp.doRequest(t, http.MethodGet, "/threads", nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.SearchThreadsResponse]](t, resp)
	require.Len(t, body.Data.Threads, 3)
	require.EqualValues(t, 3, body.Data.Pagination.Total)
}
