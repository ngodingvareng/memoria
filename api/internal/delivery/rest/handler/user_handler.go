package handler

import (
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/middleware"
	"github.com/ngodingvareng/memoria/internal/enum"
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

// GetUserByID godoc
// @ID           GetUserByID
// @Summary      Get a user's public profile
// @Description  Minimal other-facing fields (name, username, image) — e.g. resolving a Circle member's user_id
// @Tags         users
// @Produce      json
// @Param        id path string true "User ID"
// @Success      200 {object} dto.WebResponse[dto.PublicUserResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /users/{id} [get]
func (h *UserHandler) GetUserByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	user, err := h.usecase.GetPublicProfile(c, id)
	if err != nil {
		return err
	}

	return c.JSON(dto.WebResponse[dto.PublicUserResponse]{
		Code:    fiber.StatusOK,
		Message: "found",
		Data:    dto.NewPublicUserResponse(user),
	})
}

// GetUserByUsername godoc
// @ID           GetUserByUsername
// @Summary      Get a user's public profile by username
// @Description  Minimal other-facing fields (name, username, bio, image) — backs the @username profile page
// @Tags         users
// @Produce      json
// @Param        username path string true "Username"
// @Success      200 {object} dto.WebResponse[dto.PublicUserResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /users/username/{username} [get]
func (h *UserHandler) GetUserByUsername(c fiber.Ctx) error {
	username := c.Params("username")
	if !validate.UsernameFormat.MatchString(username) {
		return errs.ErrInvalidInput
	}

	user, err := h.usecase.GetPublicProfileByUsername(c, username)
	if err != nil {
		return err
	}

	return c.JSON(dto.WebResponse[dto.PublicUserResponse]{
		Code:    fiber.StatusOK,
		Message: "found",
		Data:    dto.NewPublicUserResponse(user),
	})
}

// MarkUserKnown godoc
// @ID           MarkUserKnown
// @Summary      Mark a user as known
// @Description  One-directional, silent, idempotent — grants username's "known" audience tier toward the caller (FEATURES.md, Privacy & Control)
// @Tags         users
// @Param        username path string true "Username to mark known"
// @Success      204
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /users/username/{username}/known [post]
func (h *UserHandler) MarkUserKnown(c fiber.Ctx) error {
	username := c.Params("username")
	if !validate.UsernameFormat.MatchString(username) {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	if err := h.usecase.MarkUserKnown(c, userID, username); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetMe godoc
// @Summary      Get the caller's own full profile
// @Description  Includes the privacy fields GetUserByID/GetUserByUsername deliberately omit
// @Tags         users
// @Produce      json
// @Success      200 {object} dto.WebResponse[dto.PrivateUserResponse]
// @Router       /users/me [get]
func (h *UserHandler) GetMe(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	user, err := h.usecase.GetOwnProfile(c, userID)
	if err != nil {
		return err
	}

	return c.JSON(dto.WebResponse[dto.PrivateUserResponse]{
		Code:    fiber.StatusOK,
		Message: "found",
		Data:    dto.NewPrivateUserResponse(user),
	})
}

// UpdatePrivacySettings godoc
// @Summary      Update the caller's privacy settings
// @Description  Social Interaction + Data Controls toggles (FEATURES.md, Privacy & Control)
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body dto.UpdatePrivacySettingsRequest true "New privacy settings"
// @Success      200 {object} dto.WebResponse[dto.PrivateUserResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /users/me/privacy [put]
func (h *UserHandler) UpdatePrivacySettings(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	var req dto.UpdatePrivacySettingsRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(req); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	updated, err := h.usecase.UpdatePrivacySettings(c, usecase.UpdatePrivacySettingsInput{
		UserID:                 userID,
		MentionPolicy:          enum.AudiencePolicy(req.MentionPolicy),
		CircleInvitePolicy:     enum.AudiencePolicy(req.CircleInvitePolicy),
		DiscoverableByUsername: req.DiscoverableByUsername,
		StripPhotoMetadata:     req.StripPhotoMetadata,
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c, "privacy settings updated", "user_id", userID)

	return c.JSON(dto.WebResponse[dto.PrivateUserResponse]{
		Code:    fiber.StatusOK,
		Message: "privacy settings updated",
		Data:    dto.NewPrivateUserResponse(updated),
	})
}

// BlockUser godoc
// @Summary      Block a user by username
// @Tags         users
// @Accept       json
// @Param        request body dto.BlockUserRequest true "Username to block"
// @Success      204
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /users/me/blocks [post]
func (h *UserHandler) BlockUser(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	var req dto.BlockUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(req); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	if err := h.usecase.BlockUser(c, userID, req.Username); err != nil {
		return err
	}

	slog.InfoContext(c, "user blocked", "user_id", userID)

	return c.SendStatus(fiber.StatusNoContent)
}

// UnblockUser godoc
// @Summary      Unblock a user by username
// @Tags         users
// @Param        username path string true "Username to unblock"
// @Success      204
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /users/me/blocks/{username} [delete]
func (h *UserHandler) UnblockUser(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	username := c.Params("username")
	if !validate.UsernameFormat.MatchString(username) {
		return errs.ErrInvalidInput
	}

	if err := h.usecase.UnblockUser(c, userID, username); err != nil {
		return err
	}

	slog.InfoContext(c, "user unblocked", "user_id", userID)

	return c.SendStatus(fiber.StatusNoContent)
}

// ListBlockedUsers godoc
// @Summary      List users the caller has blocked
// @Tags         users
// @Produce      json
// @Success      200 {object} dto.WebResponse[dto.ListBlockedUsersResponse]
// @Router       /users/me/blocks [get]
func (h *UserHandler) ListBlockedUsers(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	users, err := h.usecase.ListBlockedUsers(c, userID)
	if err != nil {
		return err
	}

	return c.JSON(dto.WebResponse[dto.ListBlockedUsersResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    dto.NewListBlockedUsersResponse(users),
	})
}

// MuteUser godoc
// @Summary      Mute a user by username
// @Tags         users
// @Accept       json
// @Param        request body dto.MuteUserRequest true "Username to mute"
// @Success      204
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /users/me/mutes [post]
func (h *UserHandler) MuteUser(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	var req dto.MuteUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(req); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	if err := h.usecase.MuteUser(c, userID, req.Username); err != nil {
		return err
	}

	slog.InfoContext(c, "user muted", "user_id", userID)

	return c.SendStatus(fiber.StatusNoContent)
}

// UnmuteUser godoc
// @Summary      Unmute a user by username
// @Tags         users
// @Param        username path string true "Username to unmute"
// @Success      204
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /users/me/mutes/{username} [delete]
func (h *UserHandler) UnmuteUser(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	username := c.Params("username")
	if !validate.UsernameFormat.MatchString(username) {
		return errs.ErrInvalidInput
	}

	if err := h.usecase.UnmuteUser(c, userID, username); err != nil {
		return err
	}

	slog.InfoContext(c, "user unmuted", "user_id", userID)

	return c.SendStatus(fiber.StatusNoContent)
}

// ListMutedUsers godoc
// @Summary      List users the caller has muted
// @Tags         users
// @Produce      json
// @Success      200 {object} dto.WebResponse[dto.ListMutedUsersResponse]
// @Router       /users/me/mutes [get]
func (h *UserHandler) ListMutedUsers(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	users, err := h.usecase.ListMutedUsers(c, userID)
	if err != nil {
		return err
	}

	return c.JSON(dto.WebResponse[dto.ListMutedUsersResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    dto.NewListMutedUsersResponse(users),
	})
}

// UploadProfileImage godoc
// @ID           UploadProfileImage
// @Summary      Upload the current user's profile photo
// @Description  Replaces any existing photo. Publicly readable by URL, no signing.
// @Tags         users
// @Accept       multipart/form-data
// @Produce      json
// @Param        image formData file true "Image file"
// @Success      200 {object} dto.WebResponse[dto.PublicUserResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /users/me/image [post]
func (h *UserHandler) UploadProfileImage(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	fileHeader, err := c.FormFile("image")
	if err != nil {
		return errs.New(fiber.StatusBadRequest, "image file is required")
	}
	if fileHeader.Size > maxImageUploadSize {
		return errs.New(fiber.StatusBadRequest, "image exceeds the 10MB size limit")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return errs.New(fiber.StatusBadRequest, "could not read uploaded file")
	}
	defer file.Close()

	detectedType, body, err := detectImageContentType(file)
	if err != nil {
		return errs.New(fiber.StatusBadRequest, "could not read uploaded file")
	}
	if !allowedImageContentTypes[detectedType] {
		return errs.New(fiber.StatusBadRequest, "uploaded file is not a supported image type")
	}

	user, err := h.usecase.UploadProfileImage(c, usecase.UploadProfileImageInput{
		UserID:      userID,
		FileName:    fileHeader.Filename,
		ContentType: detectedType,
		Size:        fileHeader.Size,
		Body:        body,
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c, "profile image uploaded", "user_id", userID)

	return c.JSON(dto.WebResponse[dto.PublicUserResponse]{
		Code:    fiber.StatusOK,
		Message: "profile image uploaded",
		Data:    dto.NewPublicUserResponse(user),
	})
}
