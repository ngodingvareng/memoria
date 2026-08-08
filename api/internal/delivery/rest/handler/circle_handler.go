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

type CircleHandler struct {
	usecase  usecase.CircleUsecase
	validate *validator.Validate
}

func NewCircleHandler(usecase usecase.CircleUsecase) *CircleHandler {
	return &CircleHandler{usecase: usecase, validate: validator.New()}
}

// CreateCircle godoc
// @Summary      Create a circle
// @Description  Create a new private Circle; the caller becomes its admin
// @Tags         circles
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateCircleRequest true "Create circle request"
// @Success      201 {object} dto.WebResponse[dto.CircleResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /circles [post]
func (h *CircleHandler) CreateCircle(c fiber.Ctx) error {
	var req dto.CreateCircleRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(req); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	circle, err := h.usecase.CreateCircle(c, usecase.CreateCircleInput{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		ColorHex:    req.ColorHex,
		ImagePath:   req.ImagePath,
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c, "circle created", "circle_id", circle.ID, "user_id", userID)

	return c.Status(fiber.StatusCreated).JSON(dto.WebResponse[dto.CircleResponse]{
		Code:    fiber.StatusCreated,
		Message: "circle created",
		Data:    dto.NewCircleResponse(circle),
	})
}

// UpdateCircle godoc
// @Summary      Update a circle
// @Description  Full-representation update of a Circle; admin-only
// @Tags         circles
// @Accept       json
// @Produce      json
// @Param        id      path string                  true "Circle ID"
// @Param        request body dto.UpdateCircleRequest true "Update circle request"
// @Success      200 {object} dto.WebResponse[dto.CircleResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /circles/{id} [put]
func (h *CircleHandler) UpdateCircle(c fiber.Ctx) error {
	circleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	var req dto.UpdateCircleRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(req); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	circle, err := h.usecase.UpdateCircle(c, usecase.UpdateCircleInput{
		ID:          circleID,
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		ColorHex:    req.ColorHex,
		ImagePath:   req.ImagePath,
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c, "circle updated", "circle_id", circle.ID, "user_id", userID)

	return c.JSON(dto.WebResponse[dto.CircleResponse]{
		Code:    fiber.StatusOK,
		Message: "circle updated",
		Data:    dto.NewCircleResponse(circle),
	})
}

// DissolveCircle godoc
// @Summary      Dissolve a circle
// @Description  Soft-deletes a Circle; admin-only
// @Tags         circles
// @Param        id path string true "Circle ID"
// @Success      204
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /circles/{id} [delete]
func (h *CircleHandler) DissolveCircle(c fiber.Ctx) error {
	circleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	if err := h.usecase.DissolveCircle(c, circleID, userID); err != nil {
		return err
	}

	slog.InfoContext(c, "circle dissolved", "circle_id", circleID, "user_id", userID)

	return c.SendStatus(fiber.StatusNoContent)
}

// GetCircle godoc
// @Summary      Get a circle by ID
// @Description  Get a single Circle; requires the caller to be an active member
// @Tags         circles
// @Produce      json
// @Param        id path string true "Circle ID"
// @Success      200 {object} dto.WebResponse[dto.CircleResponse]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /circles/{id} [get]
func (h *CircleHandler) GetCircle(c fiber.Ctx) error {
	circleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	circle, err := h.usecase.GetCircle(c, circleID, userID)
	if err != nil {
		return err
	}

	return c.JSON(dto.WebResponse[dto.CircleResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    dto.NewCircleResponse(circle),
	})
}

// ListMyCircles godoc
// @Summary      List my circles
// @Description  Every Circle the authenticated user is currently an active member of
// @Tags         circles
// @Produce      json
// @Success      200 {object} dto.WebResponse[dto.ListCirclesResponse]
// @Router       /circles [get]
func (h *CircleHandler) ListMyCircles(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	circles, err := h.usecase.ListMyCircles(c, userID)
	if err != nil {
		return err
	}

	responses := make([]dto.CircleResponse, len(circles))
	for i, circle := range circles {
		responses[i] = dto.NewCircleResponse(circle)
	}

	return c.JSON(dto.WebResponse[dto.ListCirclesResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    dto.ListCirclesResponse{Circles: responses},
	})
}

// ListMembers godoc
// @Summary      List a circle's members
// @Description  The Circle's active roster; requires the caller to be an active member
// @Tags         circles
// @Produce      json
// @Param        id path string true "Circle ID"
// @Success      200 {object} dto.WebResponse[dto.ListCircleMembersResponse]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /circles/{id}/members [get]
func (h *CircleHandler) ListMembers(c fiber.Ctx) error {
	circleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	members, err := h.usecase.ListMembers(c, circleID, userID)
	if err != nil {
		return err
	}

	responses := make([]dto.CircleMemberResponse, len(members))
	for i, m := range members {
		responses[i] = dto.NewCircleMemberResponse(m)
	}

	return c.JSON(dto.WebResponse[dto.ListCircleMembersResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    dto.ListCircleMembersResponse{Members: responses},
	})
}

// RemoveMember godoc
// @Summary      Remove a circle member
// @Description  Self-leave when userId is the caller's own id, otherwise an admin-forced removal
// @Tags         circles
// @Param        id     path string true "Circle ID"
// @Param        userId path string true "Target user ID"
// @Success      204
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /circles/{id}/members/{userId} [delete]
func (h *CircleHandler) RemoveMember(c fiber.Ctx) error {
	circleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}
	targetUserID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	if targetUserID == userID {
		if err := h.usecase.LeaveCircle(c, circleID, userID); err != nil {
			return err
		}
		slog.InfoContext(c, "circle left", "circle_id", circleID, "user_id", userID)
		return c.SendStatus(fiber.StatusNoContent)
	}

	if err := h.usecase.RemoveMember(c, circleID, targetUserID, userID); err != nil {
		return err
	}

	slog.InfoContext(c, "circle member removed", "circle_id", circleID, "user_id", targetUserID, "removed_by_user_id", userID)

	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateMemberRole godoc
// @Summary      Change a circle member's role
// @Description  Admin-only
// @Tags         circles
// @Accept       json
// @Produce      json
// @Param        id      path string                            true "Circle ID"
// @Param        userId  path string                            true "Target user ID"
// @Param        request body dto.UpdateCircleMemberRoleRequest true "Update role request"
// @Success      200 {object} dto.WebResponse[dto.CircleMemberResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /circles/{id}/members/{userId}/role [patch]
func (h *CircleHandler) UpdateMemberRole(c fiber.Ctx) error {
	circleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}
	targetUserID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	var req dto.UpdateCircleMemberRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(req); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	member, err := h.usecase.UpdateMemberRole(c, usecase.UpdateCircleMemberRoleInput{
		CircleID: circleID, UserID: targetUserID, ChangedByUserID: userID, Role: enum.CircleRole(req.Role),
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c, "circle member role updated", "circle_id", circleID, "user_id", targetUserID, "changed_by_user_id", userID)

	return c.JSON(dto.WebResponse[dto.CircleMemberResponse]{
		Code:    fiber.StatusOK,
		Message: "circle member role updated",
		Data:    dto.NewCircleMemberResponse(member),
	})
}

// UpdateMemberPermissions godoc
// @Summary      Change a circle member's permissions
// @Description  Admin-only
// @Tags         circles
// @Accept       json
// @Produce      json
// @Param        id      path string                                   true "Circle ID"
// @Param        userId  path string                                   true "Target user ID"
// @Param        request body dto.UpdateCircleMemberPermissionsRequest true "Update permissions request"
// @Success      200 {object} dto.WebResponse[dto.CircleMemberResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /circles/{id}/members/{userId} [patch]
func (h *CircleHandler) UpdateMemberPermissions(c fiber.Ctx) error {
	circleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}
	targetUserID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	var req dto.UpdateCircleMemberPermissionsRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	member, err := h.usecase.UpdateMemberPermissions(c, usecase.UpdateCircleMemberPermissionsInput{
		CircleID: circleID, UserID: targetUserID, ChangedByUserID: userID,
		CanInvite: req.CanInvite, CanCapture: req.CanCapture,
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c, "circle member permissions updated", "circle_id", circleID, "user_id", targetUserID, "changed_by_user_id", userID)

	return c.JSON(dto.WebResponse[dto.CircleMemberResponse]{
		Code:    fiber.StatusOK,
		Message: "circle member permissions updated",
		Data:    dto.NewCircleMemberResponse(member),
	})
}
