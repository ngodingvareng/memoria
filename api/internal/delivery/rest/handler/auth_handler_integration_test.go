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

func registerUser(t *testing.T, testApp *testApp, email string) *http.Response {
	t.Helper()
	return testApp.doRequest(t, http.MethodPost, "/auth/register", dto.RegisterRequest{
		Name:     "Test User",
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

	resp := registerUser(t, testApp, email)

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.LoginResponse]](t, resp)
	require.Equal(t, "Test User", body.Data.User.Name)
	require.Nil(t, body.Data.User.Username, "username is claimed later via PATCH /users/me/username, not at register")
	require.Equal(t, email, body.Data.User.Email)
	require.False(t, body.Data.User.EmailVerified)
}

func TestAuthHandler_Register_IssuesSessionLikeLogin(t *testing.T) {
	// Register doubles as login — no separate POST /auth/login round
	// trip is needed before the welcome/username-claim step.
	testApp := setupTestApp(t)

	resp := registerUser(t, testApp, testEmail())

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.LoginResponse]](t, resp)
	require.NotEmpty(t, body.Data.AccessToken)
	require.Positive(t, body.Data.ExpiresIn)

	var found bool
	for _, c := range resp.Cookies() {
		if c.Name == middleware.RefreshCookieName {
			found = true
			require.NotEmpty(t, c.Value)
		}
	}
	require.True(t, found, "expected a %s cookie to be set", middleware.RefreshCookieName)
}

func TestAuthHandler_Register_ValidationError_ShortPassword(t *testing.T) {
	testApp := setupTestApp(t)

	resp := testApp.doRequest(t, http.MethodPost, "/auth/register", dto.RegisterRequest{
		Name:     "Test User",
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

	first := registerUser(t, testApp, email)
	require.Equal(t, http.StatusCreated, first.StatusCode)

	second := registerUser(t, testApp, email)
	require.Equal(t, http.StatusConflict, second.StatusCode)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	testApp := setupTestApp(t)
	email := testEmail()
	require.Equal(t, http.StatusCreated, registerUser(t, testApp, email).StatusCode)

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
	require.Equal(t, http.StatusCreated, registerUser(t, testApp, email).StatusCode)

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
	require.Equal(t, http.StatusCreated, registerUser(t, testApp, email).StatusCode)

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
	require.Equal(t, http.StatusCreated, registerUser(t, testApp, email).StatusCode)

	loginResp := loginUser(t, testApp, email)
	require.Equal(t, http.StatusOK, loginResp.StatusCode)
	originalToken := refreshCookieValue(t, loginResp)

	resp := testApp.doRequest(t, http.MethodPost, "/auth/refresh", nil, refreshCookieHeader(originalToken))

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeBody[dto.WebResponse[dto.LoginResponse]](t, resp)
	require.NotEmpty(t, body.Data.AccessToken)
	// Not asserting body.Data.AccessToken != loginBody.Data.AccessToken:
	// access token claims (sub, iss, iat, exp) are second-granularity
	// (jwt.NewNumericDate) and HS256 signing is deterministic, so a login
	// immediately followed by a refresh within the same wall-clock second
	// legitimately produces a byte-identical JWT. The refresh token cookie
	// below is the actual rotation guarantee this endpoint makes.

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
	require.Equal(t, http.StatusCreated, registerUser(t, testApp, email).StatusCode)

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
	require.Equal(t, http.StatusCreated, registerUser(t, testApp, email).StatusCode)

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
