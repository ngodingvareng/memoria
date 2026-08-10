//go:build integration

package handler_test

import (
	"net/http"
	"testing"

	"github.com/go-openapi/testify/v2/require"

	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
)

// These exercise middleware.RequireUsername end to end via /threads (an
// arbitrary "activity" route — every non-exempt group gets the same
// middleware, so this one stands in for all of them) and confirm the
// two onboarding endpoints stay reachable without a username.

func TestRequireUsername_ActivityRoute_BlocksAccountWithoutUsername(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUserWithoutUsername(t, testApp.pool)

	resp := testApp.doRequest(t, http.MethodPost, "/threads", dto.CreateThreadRequest{
		Name: "Morning workout",
	}, map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestRequireUsername_ActivityRoute_AllowsAccountWithUsername(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUser(t, testApp.pool)

	resp := testApp.doRequest(t, http.MethodPost, "/threads", dto.CreateThreadRequest{
		Name: "Morning workout",
	}, map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestRequireUsername_UsernameAvailability_ExemptForAccountWithoutUsername(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUserWithoutUsername(t, testApp.pool)

	resp := testApp.doRequest(t, http.MethodGet, "/users/username-availability?username=budisantoso", nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequireUsername_SetUsername_ExemptForAccountWithoutUsername(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUserWithoutUsername(t, testApp.pool)

	resp := testApp.doRequest(t, http.MethodPatch, "/users/me/username", dto.SetUsernameRequest{
		Username: "budisantoso",
	}, map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.UserResponse]](t, resp)
	require.NotNil(t, body.Data.Username)
	require.Equal(t, "budisantoso", *body.Data.Username)
}

func TestRequireUsername_GetMe_BlocksAccountWithoutUsername(t *testing.T) {
	testApp := setupTestApp(t)
	userID := seedTestUserWithoutUsername(t, testApp.pool)

	resp := testApp.doRequest(t, http.MethodGet, "/users/me", nil,
		map[string]string{"Authorization": testApp.authHeader(t, userID)})

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}
