package handler

import (
	"errors"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/middleware"
	"github.com/ngodingvareng/memoria/internal/usecase"
	"github.com/ngodingvareng/memoria/pkg/errs"
	"github.com/ngodingvareng/memoria/pkg/util"
)

type AuthHandler struct {
	usecase  usecase.AuthUsecase
	validate *validator.Validate
	// secureCookies should be true in production (HTTPS) and false for
	// local dev over plain HTTP — browsers refuse to store a cookie
	// marked Secure over a non-HTTPS connection, so leaving this true
	// during local http://localhost development would make login appear
	// to "succeed" while the cookie is silently never actually set.
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
		return &errs.ValidationError{Errors: util.FormatValidationErrors(err)}
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
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Login request"
// @Success      200 {object} dto.WebResponse[dto.UserResponse]
// @Failure      401 {object} dto.WebResponse[any]
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(req); err != nil {
		return &errs.ValidationError{Errors: util.FormatValidationErrors(err)}
	}

	ip := c.IP()
	userAgent := string(c.Request().Header.UserAgent())

	result, err := h.usecase.Login(c, usecase.LoginInput{
		Email:     req.Email,
		Password:  req.Password,
		IPAddress: &ip,
		UserAgent: &userAgent,
	})
	if err != nil {
		return err
	}

	h.setSessionCookie(c, result.SessionToken, result.ExpiresAt)

	slog.InfoContext(c, "user logged in", "user_id", result.User.ID)

	return c.JSON(dto.WebResponse[dto.UserResponse]{
		Code:    fiber.StatusOK,
		Message: "logged in",
		Data:    dto.NewUserResponse(result.User),
	})
}

// Logout godoc
// @Summary      Log out of the current session
// @Tags         auth
// @Success      204
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c fiber.Ctx) error {
	token := c.Cookies(middleware.SessionCookieName)
	if token != "" {
		if err := h.usecase.Logout(c, token); err != nil && !errors.Is(err, errs.ErrNotFound) {
			return err
		}
	}
	h.clearSessionCookie(c)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AuthHandler) setSessionCookie(c fiber.Ctx, token string, expiresAt time.Time) {
	c.Cookie(&fiber.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    token,
		Expires:  expiresAt,
		HTTPOnly: true,
		Secure:   h.secureCookies,
		SameSite: "Lax",
		Path:     "/",
	})
}

func (h *AuthHandler) clearSessionCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   h.secureCookies,
		SameSite: "Lax",
		Path:     "/",
	})
}
