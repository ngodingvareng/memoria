package handler

import (
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/middleware"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
	"github.com/ngodingvareng/memoria/internal/validate"
)

type AuthHandler struct {
	usecase  usecase.AuthUsecase
	validate *validator.Validate
	// secureCookies should be true in production (HTTPS) and false for
	// local dev over plain HTTP — browsers refuse to store a cookie
	// marked Secure over a non-HTTPS connection, so leaving this true
	// during local http://localhost development would make the refresh
	// cookie silently never actually get set.
	secureCookies bool
}

func NewAuthHandler(uc usecase.AuthUsecase, secureCookies bool) *AuthHandler {
	v := validator.New()

	return &AuthHandler{
		usecase:       uc,
		validate:      v,
		secureCookies: secureCookies,
	}
}

// Register godoc
// @Summary      Register a new account
// @Description  Creates the account (without a username — see
// @Description  PATCH /users/me/username for claiming one afterward) and
// @Description  immediately starts a session, exactly like /auth/login:
// @Description  returns an access token and sets the refresh token as an
// @Description  httpOnly cookie scoped to /auth.
// @ID           Register
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "Register request"
// @Success      201 {object} dto.WebResponse[dto.LoginResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      409 {object} dto.WebResponse[any]
// @Router       /auth/register [post].
func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(req); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	ip := c.IP()
	userAgent := string(c.Request().Header.UserAgent())

	tokens, err := h.usecase.Register(c, usecase.RegisterInput{
		Name:      req.Name,
		Email:     req.Email,
		Password:  req.Password,
		IPAddress: &ip,
		UserAgent: &userAgent,
	})
	if err != nil {
		return err
	}

	h.setRefreshCookie(c, tokens.RefreshToken, tokens.RefreshTokenExpiresAt)

	slog.InfoContext(c, "user registered", "user_id", tokens.User.ID)

	return c.Status(fiber.StatusCreated).JSON(dto.WebResponse[dto.LoginResponse]{
		Code:    fiber.StatusCreated,
		Message: "registered",
		Data:    newLoginResponse(tokens),
	})
}

// Login godoc
// @Summary      Log in with email and password
// @Description  Returns an access token in the response body and sets
// @Description  the refresh token as an httpOnly cookie scoped to /auth.
// @ID           Login
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Login request"
// @Success      200 {object} dto.WebResponse[dto.LoginResponse]
// @Failure      401 {object} dto.WebResponse[any]
// @Router       /auth/login [post].
func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(req); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	ip := c.IP()
	userAgent := string(c.Request().Header.UserAgent())

	tokens, err := h.usecase.Login(c, usecase.LoginInput{
		Email:     req.Email,
		Password:  req.Password,
		IPAddress: &ip,
		UserAgent: &userAgent,
	})
	if err != nil {
		return err
	}

	h.setRefreshCookie(c, tokens.RefreshToken, tokens.RefreshTokenExpiresAt)

	slog.InfoContext(c, "user logged in", "user_id", tokens.User.ID)

	return c.JSON(dto.WebResponse[dto.LoginResponse]{
		Code:    fiber.StatusOK,
		Message: "logged in",
		Data:    newLoginResponse(tokens),
	})
}

// GoogleLogin godoc
// @Summary      Log in (or sign up) with a Google ID token
// @Description  Verifies a Google Identity Services ID token server-side
// @Description  and either logs into the account already linked to it,
// @Description  or creates a new one. If the token's email already
// @Description  belongs to a different, non-Google account, this returns
// @Description  409 rather than silently linking — the user should log
// @Description  in with their existing method instead.
// @ID           GoogleLogin
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.GoogleLoginRequest true "Google ID token"
// @Success      200 {object} dto.WebResponse[dto.LoginResponse]
// @Failure      401 {object} dto.WebResponse[any]
// @Failure      409 {object} dto.WebResponse[any]
// @Router       /auth/google [post].
func (h *AuthHandler) GoogleLogin(c fiber.Ctx) error {
	var req dto.GoogleLoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(req); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	ip := c.IP()
	userAgent := string(c.Request().Header.UserAgent())

	tokens, err := h.usecase.LoginWithGoogle(c, usecase.GoogleLoginInput{
		IDToken:   req.IDToken,
		IPAddress: &ip,
		UserAgent: &userAgent,
	})
	if err != nil {
		return err
	}

	h.setRefreshCookie(c, tokens.RefreshToken, tokens.RefreshTokenExpiresAt)

	slog.InfoContext(c, "user logged in via google", "user_id", tokens.User.ID)

	return c.JSON(dto.WebResponse[dto.LoginResponse]{
		Code:    fiber.StatusOK,
		Message: "logged in",
		Data:    newLoginResponse(tokens),
	})
}

// Refresh godoc
// @Summary      Exchange the refresh token cookie for a new access/refresh token pair
// @Description  Reads the refresh token from the httpOnly cookie, rotates
// @Description  it, and returns a new access token plus a new refresh
// @Description  cookie. Reusing an already-rotated refresh token revokes
// @Description  its entire token family and forces a fresh login.
// @ID           RefreshToken
// @Tags         auth
// @Produce      json
// @Success      200 {object} dto.WebResponse[dto.LoginResponse]
// @Failure      401 {object} dto.WebResponse[any]
// @Router       /auth/refresh [post].
func (h *AuthHandler) Refresh(c fiber.Ctx) error {
	rawToken := c.Cookies(middleware.RefreshCookieName)
	if rawToken == "" {
		return errs.ErrUnauthorized
	}

	ip := c.IP()
	userAgent := string(c.Request().Header.UserAgent())

	tokens, err := h.usecase.Refresh(c, usecase.RefreshInput{
		RefreshToken: rawToken,
		IPAddress:    &ip,
		UserAgent:    &userAgent,
	})
	if err != nil {
		// Not found, expired, or reuse detected — either way the cookie
		// the client is holding is no longer good for anything, so clear
		// it rather than leaving a dead cookie around.
		h.clearRefreshCookie(c)
		return err
	}

	h.setRefreshCookie(c, tokens.RefreshToken, tokens.RefreshTokenExpiresAt)

	slog.InfoContext(c, "session refreshed", "user_id", tokens.User.ID)

	return c.JSON(dto.WebResponse[dto.LoginResponse]{
		Code:    fiber.StatusOK,
		Message: "refreshed",
		Data:    newLoginResponse(tokens),
	})
}

// Logout godoc
// @ID           Logout
// @Summary      Log out of the current session
// @Tags         auth
// @Success      204
// @Router       /auth/logout [post].
func (h *AuthHandler) Logout(c fiber.Ctx) error {
	// usecase.Logout is idempotent — a missing/already-revoked token
	// resolves to a nil error there, so there's nothing to distinguish
	// here beyond "did it fail".
	token := c.Cookies(middleware.RefreshCookieName)
	if token != "" {
		if err := h.usecase.Logout(c, token); err != nil {
			return err
		}
		slog.InfoContext(c, "user logged out")
	}
	h.clearRefreshCookie(c)
	return c.SendStatus(fiber.StatusNoContent)
}

// ForgotPassword godoc
// @Summary      Request a password reset link
// @Description  Always responds the same way whether or not email belongs to an account, so this can never be used to check which emails are registered.
// @ID           ForgotPassword
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.ForgotPasswordRequest true "Email to send the reset link to"
// @Success      200 {object} dto.WebResponse[any]
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /auth/forgot-password [post].
func (h *AuthHandler) ForgotPassword(c fiber.Ctx) error {
	var req dto.ForgotPasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(req); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	if err := h.usecase.ForgotPassword(c, req.Email); err != nil {
		return err
	}

	slog.InfoContext(c, "password reset requested", "email", req.Email)

	return c.JSON(dto.WebResponse[any]{
		Code:    fiber.StatusOK,
		Message: "if that email is registered, a reset link has been sent",
	})
}

// ResetPassword godoc
// @Summary      Reset a password using the emailed token
// @Description  Also revokes every existing session for the account, so a stolen session stops working the moment the password changes.
// @ID           ResetPassword
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.ResetPasswordRequest true "Email, reset token, and new password"
// @Success      200 {object} dto.WebResponse[any]
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      401 {object} dto.WebResponse[any]
// @Router       /auth/reset-password [post].
func (h *AuthHandler) ResetPassword(c fiber.Ctx) error {
	var req dto.ResetPasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(req); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	if err := h.usecase.ResetPassword(c, usecase.ResetPasswordInput{
		Email:       req.Email,
		Token:       req.Token,
		NewPassword: req.NewPassword,
	}); err != nil {
		return err
	}

	slog.InfoContext(c, "password reset", "email", req.Email)

	return c.JSON(dto.WebResponse[any]{
		Code:    fiber.StatusOK,
		Message: "password reset",
	})
}

// newLoginResponse builds the shared success body for both Login and
// Refresh — the two endpoints return the exact same shape since Refresh
// is really just "log in again using the refresh cookie instead of a
// password".
func newLoginResponse(tokens *usecase.AuthTokens) dto.LoginResponse {
	return dto.LoginResponse{
		User:        dto.NewUserResponse(tokens.User),
		AccessToken: tokens.AccessToken,
		ExpiresIn:   int(time.Until(tokens.AccessTokenExpiresAt).Seconds()),
	}
}

func (h *AuthHandler) setRefreshCookie(c fiber.Ctx, token string, expiresAt time.Time) {
	c.Cookie(&fiber.Cookie{
		Name:     middleware.RefreshCookieName,
		Value:    token,
		Expires:  expiresAt,
		HTTPOnly: true,
		Secure:   h.secureCookies,
		SameSite: "Lax",
		// Scoped to /auth only — this cookie has no business riding
		// along on every API request, unlike the old session cookie.
		Path: "/auth",
	})
}

func (h *AuthHandler) clearRefreshCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     middleware.RefreshCookieName,
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   h.secureCookies,
		SameSite: "Lax",
		Path:     "/auth",
	})
}
