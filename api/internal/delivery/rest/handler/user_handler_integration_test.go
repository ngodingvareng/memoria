//go:build integration

package handler_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/go-openapi/testify/v2/require"

	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
)

// seedTestUser (main_integration_test.go) and seedTestUserWithUsername
// (circle_handler_integration_test.go) are shared helpers already
// defined elsewhere in this package.

func TestUserHandler_CheckUsernameAvailability_Available(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	resp := testApp.doRequest(t, http.MethodGet, "/users/username-availability?username=budisantoso", nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.CheckUsernameAvailabilityResponse]](t, resp)
	require.True(t, body.Data.Available)
}

func TestUserHandler_CheckUsernameAvailability_Taken(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)
	_, takenUsername := seedTestUserWithUsername(t, testApp.pool)

	resp := testApp.doRequest(t, http.MethodGet, "/users/username-availability?username="+url.QueryEscape(takenUsername), nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.CheckUsernameAvailabilityResponse]](t, resp)
	require.False(t, body.Data.Available)
}

func TestUserHandler_CheckUsernameAvailability_InvalidFormat(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	resp := testApp.doRequest(t, http.MethodGet, "/users/username-availability?username=AB", nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUserHandler_CheckUsernameAvailability_Unauthenticated(t *testing.T) {
	testApp := setupTestApp(t)

	resp := testApp.doRequest(t, http.MethodGet, "/users/username-availability?username=budisantoso", nil, nil)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestUserHandler_SetUsername_Success(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	resp := testApp.doRequest(t, http.MethodPatch, "/users/me/username", dto.SetUsernameRequest{
		Username: "budisantoso",
	}, map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.UserResponse]](t, resp)
	require.NotNil(t, body.Data.Username)
	require.Equal(t, "budisantoso", *body.Data.Username)
}

func TestUserHandler_SetUsername_AlreadyTaken_Conflict(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)
	_, takenUsername := seedTestUserWithUsername(t, testApp.pool)

	resp := testApp.doRequest(t, http.MethodPatch, "/users/me/username", dto.SetUsernameRequest{
		Username: takenUsername,
	}, map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestUserHandler_SetUsername_InvalidFormat(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	resp := testApp.doRequest(t, http.MethodPatch, "/users/me/username", dto.SetUsernameRequest{
		Username: "AB", // uppercase and shorter than the min length of 3
	}, map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body := decodeBody[dto.WebResponse[any]](t, resp)
	require.Equal(t, "Validation failed", body.Message)
	require.NotNil(t, body.Errors)
}

func TestUserHandler_SetUsername_Unauthenticated(t *testing.T) {
	testApp := setupTestApp(t)

	resp := testApp.doRequest(t, http.MethodPatch, "/users/me/username", dto.SetUsernameRequest{
		Username: "budisantoso",
	}, nil)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
