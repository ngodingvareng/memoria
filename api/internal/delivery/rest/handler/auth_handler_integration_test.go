//go:build integration

package handler_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/go-openapi/testify/v2/require"
	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/middleware"
)

const testPassword = "correct horse battery staple"

func testEmail() string {
	return uuid.NewString() + "@example.com"
}

func registerUser(t *testing.T, testApp *testApp, email, username string) *http.Response {
	t.Helper()
	return testApp.doRequest(t, http.MethodPost, "/auth/register", dto.RegisterRequest{
		Name:     "Test User",
		Username: username,
		Email:    email,
		Password: testPassword,
	}, nil)
}

func loginUser(t *testing.T, testApp *testApp, email string) *http.Response {
	t.Helper()
	return testApp.doRequest(t, http.MethodPost, "/auth/login", dto.LoginRequest{
		Email:    email,
		Password: testPassword,
	}, nil)
}

func refreshCookieValue(t *testing.T, resp *http.Response) string {
	t.Helper()
	for _, c := range resp.Cookies() {
		if c.Name == middleware.RefreshCookieName {
			return c.Value
		}
	}
	t.Fatalf("no %s cookie in response", middleware.RefreshCookieName)
	return ""
}

func refreshCookieHeader(token string) map[string]string {
	return map[string]string{"Cookie": fmt.Sprintf("%s=%s", middleware.RefreshCookieName, token)}
}

func TestAuthHandler_Register_Success(t *testing.T) {
	testApp := setupTestApp(t)
	email := testEmail()
	username := testUsername()

	resp := registerUser(t, testApp, email, username)

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.UserResponse]](t, resp)
	require.Equal(t, "Test User", body.Data.Name)
	require.Equal(t, username, body.Data.Username)
	require.Equal(t, email, body.Data.Email)
	require.False(t, body.Data.EmailVerified)
}

func TestAuthHandler_Register_ValidationError_InvalidUsername(t *testing.T) {
	testApp := setupTestApp(t)

	resp := testApp.doRequest(t, http.MethodPost, "/auth/register", dto.RegisterRequest{
		Name:     "Test User",
		Username: "AB", // uppercase and shorter than the min length of 3
		Email:    testEmail(),
		Password: testPassword,
	}, nil)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body := decodeBody[dto.WebResponse[any]](t, resp)
	require.Equal(t, "Validation failed", body.Message)
	require.NotNil(t, body.Errors)
}

func TestAuthHandler_Register_ValidationError_ShortPassword(t *testing.T) {
	testApp := setupTestApp(t)

	resp := testApp.doRequest(t, http.MethodPost, "/auth/register", dto.RegisterRequest{
		Name:     "Test User",
		Username: testUsername(),
		Email:    testEmail(),
		Password: "short", // below min=8
	}, nil)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body := decodeBody[dto.WebResponse[any]](t, resp)
	require.Equal(t, "Validation failed", body.Message)
	require.NotNil(t, body.Errors)
}

func TestAuthHandler_Register_DuplicateEmail_Conflict(t *testing.T) {
	testApp := setupTestApp(t)
	email := testEmail()

	first := registerUser(t, testApp, email, testUsername())
	require.Equal(t, http.StatusCreated, first.StatusCode)

	second := registerUser(t, testApp, email, testUsername())
	require.Equal(t, http.StatusConflict, second.StatusCode)
}

func TestAuthHandler_Register_DuplicateUsername_Conflict(t *testing.T) {
	testApp := setupTestApp(t)
	username := testUsername()

	first := registerUser(t, testApp, testEmail(), username)
	require.Equal(t, http.StatusCreated, first.StatusCode)

	second := registerUser(t, testApp, testEmail(), username)
	require.Equal(t, http.StatusConflict, second.StatusCode)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	testApp := setupTestApp(t)
	email := testEmail()
	require.Equal(t, http.StatusCreated, registerUser(t, testApp, email, testUsername()).StatusCode)

	resp := loginUser(t, testApp, email)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.LoginResponse]](t, resp)
	require.NotEmpty(t, body.Data.AccessToken)
	require.Positive(t, body.Data.ExpiresIn)
	require.Equal(t, email, body.Data.User.Email)
}

func TestAuthHandler_Login_SetsRefreshCookie(t *testing.T) {
	testApp := setupTestApp(t)
	email := testEmail()
	require.Equal(t, http.StatusCreated, registerUser(t, testApp, email, testUsername()).StatusCode)

	resp := loginUser(t, testApp, email)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	var found bool
	for _, c := range resp.Cookies() {
		if c.Name == middleware.RefreshCookieName {
			found = true
			require.NotEmpty(t, c.Value)
		}
	}
	require.True(t, found, "expected a %s cookie to be set", middleware.RefreshCookieName)
}

func TestAuthHandler_Login_WrongPassword_Unauthorized(t *testing.T) {
	testApp := setupTestApp(t)
	email := testEmail()
	require.Equal(t, http.StatusCreated, registerUser(t, testApp, email, testUsername()).StatusCode)

	resp := testApp.doRequest(t, http.MethodPost, "/auth/login", dto.LoginRequest{
		Email:    email,
		Password: "wrong password entirely",
	}, nil)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandler_Login_UnknownEmail_Unauthorized(t *testing.T) {
	testApp := setupTestApp(t)

	resp := loginUser(t, testApp, testEmail())

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandler_Refresh_Success(t *testing.T) {
	testApp := setupTestApp(t)
	email := testEmail()
	require.Equal(t, http.StatusCreated, registerUser(t, testApp, email, testUsername()).StatusCode)

	loginResp := loginUser(t, testApp, email)
	require.Equal(t, http.StatusOK, loginResp.StatusCode)
	loginBody := decodeBody[dto.WebResponse[dto.LoginResponse]](t, loginResp)
	originalToken := refreshCookieValue(t, loginResp)

	resp := testApp.doRequest(t, http.MethodPost, "/auth/refresh", nil, refreshCookieHeader(originalToken))

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.LoginResponse]](t, resp)
	require.NotEmpty(t, body.Data.AccessToken)
	require.NotEqual(t, loginBody.Data.AccessToken, body.Data.AccessToken)

	rotatedToken := refreshCookieValue(t, resp)
	require.NotEmpty(t, rotatedToken)
	require.NotEqual(t, originalToken, rotatedToken)
}

func TestAuthHandler_Refresh_NoCookie_Unauthorized(t *testing.T) {
	testApp := setupTestApp(t)

	resp := testApp.doRequest(t, http.MethodPost, "/auth/refresh", nil, nil)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandler_Refresh_ReusedToken_Unauthorized(t *testing.T) {
	testApp := setupTestApp(t)
	email := testEmail()
	require.Equal(t, http.StatusCreated, registerUser(t, testApp, email, testUsername()).StatusCode)

	loginResp := loginUser(t, testApp, email)
	originalToken := refreshCookieValue(t, loginResp)

	firstRefresh := testApp.doRequest(t, http.MethodPost, "/auth/refresh", nil, refreshCookieHeader(originalToken))
	require.Equal(t, http.StatusOK, firstRefresh.StatusCode)

	// originalToken was rotated away by the refresh above — reusing it
	// now must be treated as potential theft (auth_usecase.go's Refresh:
	// RevokedAt != nil revokes the whole family) rather than silently
	// succeeding again.
	secondUse := testApp.doRequest(t, http.MethodPost, "/auth/refresh", nil, refreshCookieHeader(originalToken))

	require.Equal(t, http.StatusUnauthorized, secondUse.StatusCode)
}

func TestAuthHandler_Logout_WithCookie_NoContent(t *testing.T) {
	testApp := setupTestApp(t)
	email := testEmail()
	require.Equal(t, http.StatusCreated, registerUser(t, testApp, email, testUsername()).StatusCode)

	loginResp := loginUser(t, testApp, email)
	token := refreshCookieValue(t, loginResp)

	resp := testApp.doRequest(t, http.MethodPost, "/auth/logout", nil, refreshCookieHeader(token))

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestAuthHandler_Logout_NoCookie_NoContent(t *testing.T) {
	// AuthHandler.Logout is deliberately idempotent — no refresh cookie at
	// all still resolves to 204, since there's nothing to distinguish
	// from "already logged out".
	testApp := setupTestApp(t)

	resp := testApp.doRequest(t, http.MethodPost, "/auth/logout", nil, nil)

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}
