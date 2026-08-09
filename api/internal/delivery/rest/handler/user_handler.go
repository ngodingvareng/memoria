package handler

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/middleware"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
	"github.com/ngodingvareng/memoria/internal/validate"
)

type UserHandler struct {
	usecase  usecase.UserUsecase
	validate *validator.Validate
}

func NewUserHandler(uc usecase.UserUsecase) *UserHandler {
	v := validator.New()
	validate.RegisterUsernameValidator(v)

	return &UserHandler{usecase: uc, validate: v}
}

// CheckUsernameAvailability godoc
// @ID           CheckUsernameAvailability
// @Summary      Check whether a username is available to claim
// @Tags         users
// @Produce      json
// @Param        username query string true "Username to check"
// @Success      200 {object} dto.WebResponse[dto.CheckUsernameAvailabilityResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /users/username-availability [get]
func (h *UserHandler) CheckUsernameAvailability(c fiber.Ctx) error {
	username := c.Query("username")
	if !validate.UsernameFormat.MatchString(username) {
		return errs.ErrInvalidInput
	}

	available, err := h.usecase.CheckUsernameAvailability(c, username)
	if err != nil {
		return err
	}

	return c.JSON(dto.WebResponse[dto.CheckUsernameAvailabilityResponse]{
		Code:    fiber.StatusOK,
		Message: "checked",
		Data:    dto.CheckUsernameAvailabilityResponse{Available: available},
	})
}

// SetUsername godoc
// @ID           SetUsername
// @Summary      Claim a username for the current account
// @Description  Used by the post-register welcome/onboarding step.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body dto.SetUsernameRequest true "Username to claim"
// @Success      200 {object} dto.WebResponse[dto.UserResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      409 {object} dto.WebResponse[any]
// @Router       /users/me/username [patch]
func (h *UserHandler) SetUsername(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	var req dto.SetUsernameRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(req); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	user, err := h.usecase.SetUsername(c, userID, req.Username)
	if err != nil {
		return err
	}

	return c.JSON(dto.WebResponse[dto.UserResponse]{
		Code:    fiber.StatusOK,
		Message: "username set",
		Data:    dto.NewUserResponse(user),
	})
}
