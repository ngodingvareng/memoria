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
	return &AuthHandler{
		usecase:       uc,
		validate:      validator.New(),
		secureCookies: secureCookies,
	}
}

// Register godoc
// @Summary      Register a new account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "Register request"
// @Success      201 {object} dto.WebResponse[dto.UserResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      409 {object} dto.WebResponse[any]
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(req); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	user, err := h.usecase.Register(c, usecase.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c, "user registered", "user_id", user.ID)

	return c.Status(fiber.StatusCreated).JSON(dto.WebResponse[dto.UserResponse]{
		Code:    fiber.StatusCreated,
		Message: "registered",
		Data:    dto.NewUserResponse(user),
	})
}

// Login godoc
// @Summary      Log in with email and password
// @Description  Returns an access token in the response body and sets
// @Description  the refresh token as an httpOnly cookie scoped to /auth.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Login request"
// @Success      200 {object} dto.WebResponse[dto.LoginResponse]
// @Failure      401 {object} dto.WebResponse[any]
// @Router       /auth/login [post]
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

// Refresh godoc
// @Summary      Exchange the refresh token cookie for a new access/refresh token pair
// @Description  Reads the refresh token from the httpOnly cookie, rotates
// @Description  it, and returns a new access token plus a new refresh
// @Description  cookie. Reusing an already-rotated refresh token revokes
// @Description  its entire token family and forces a fresh login.
// @Tags         auth
// @Produce      json
// @Success      200 {object} dto.WebResponse[dto.LoginResponse]
// @Failure      401 {object} dto.WebResponse[any]
// @Router       /auth/refresh [post]
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
// @Summary      Log out of the current session
// @Tags         auth
// @Success      204
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c fiber.Ctx) error {
	// usecase.Logout is idempotent — a missing/already-revoked token
	// resolves to a nil error there, so there's nothing to distinguish
	// here beyond "did it fail".
	token := c.Cookies(middleware.RefreshCookieName)
	if token != "" {
		if err := h.usecase.Logout(c, token); err != nil {
			return err
		}
	}
	h.clearRefreshCookie(c)
	return c.SendStatus(fiber.StatusNoContent)
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
