package handler

import (
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/middleware"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
	"github.com/ngodingvareng/memoria/internal/validate"
)

type CircleInviteHandler struct {
	usecase  usecase.CircleInviteUsecase
	validate *validator.Validate
}

func NewCircleInviteHandler(usecase usecase.CircleInviteUsecase) *CircleInviteHandler {
	return &CircleInviteHandler{usecase: usecase, validate: validator.New()}
}

// InviteDirect godoc
// @Summary      Directly invite users to a circle
// @Description  Add users by username; joins immediately unless the recipient's own privacy setting holds it back as a pending invite
// @Tags         circle-invites
// @Accept       json
// @Produce      json
// @Param        id      path string                     true "Circle ID"
// @Param        request body dto.InviteDirectRequest true "Usernames to invite"
// @Success      200 {object} dto.WebResponse[dto.InviteDirectResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /circles/{id}/members/direct [post]
func (h *CircleInviteHandler) InviteDirect(c fiber.Ctx) error {
	circleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	var req dto.InviteDirectRequest
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

	result, err := h.usecase.InviteDirect(c, usecase.InviteDirectInput{
		CircleID: circleID, InvitedByUserID: userID, Usernames: req.Usernames,
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c, "circle direct invite processed", "circle_id", circleID, "user_id", userID)

	return c.JSON(dto.WebResponse[dto.InviteDirectResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    dto.NewInviteDirectResponse(result),
	})
}

// ListMyPendingInvites godoc
// @Summary      List my pending circle invites
// @Tags         circle-invites
// @Produce      json
// @Success      200 {object} dto.WebResponse[dto.ListCircleInvitesResponse]
// @Router       /circle-invites [get]
func (h *CircleInviteHandler) ListMyPendingInvites(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	invites, err := h.usecase.ListMyPendingInvites(c, userID)
	if err != nil {
		return err
	}

	responses := make([]dto.CircleInviteResponse, len(invites))
	for i, inv := range invites {
		responses[i] = dto.NewCircleInviteResponse(inv)
	}

	return c.JSON(dto.WebResponse[dto.ListCircleInvitesResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    dto.ListCircleInvitesResponse{Invites: responses},
	})
}

// AcceptInvite godoc
// @Summary      Accept a circle invite
// @Tags         circle-invites
// @Produce      json
// @Param        id path string true "Invite ID"
// @Success      200 {object} dto.WebResponse[dto.CircleMemberResponse]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /circle-invites/{id}/accept [post]
func (h *CircleInviteHandler) AcceptInvite(c fiber.Ctx) error {
	inviteID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	member, err := h.usecase.AcceptInvite(c, inviteID, userID)
	if err != nil {
		return err
	}

	slog.InfoContext(c, "circle invite accepted", "invite_id", inviteID, "user_id", userID)

	return c.JSON(dto.WebResponse[dto.CircleMemberResponse]{
		Code:    fiber.StatusOK,
		Message: "circle invite accepted",
		Data:    dto.NewCircleMemberResponse(member),
	})
}

// DeclineInvite godoc
// @Summary      Decline a circle invite
// @Tags         circle-invites
// @Produce      json
// @Param        id path string true "Invite ID"
// @Success      200 {object} dto.WebResponse[dto.CircleInviteResponse]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /circle-invites/{id}/decline [post]
func (h *CircleInviteHandler) DeclineInvite(c fiber.Ctx) error {
	inviteID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	invite, err := h.usecase.DeclineInvite(c, inviteID, userID)
	if err != nil {
		return err
	}

	slog.InfoContext(c, "circle invite declined", "invite_id", inviteID, "user_id", userID)

	return c.JSON(dto.WebResponse[dto.CircleInviteResponse]{
		Code:    fiber.StatusOK,
		Message: "circle invite declined",
		Data:    dto.NewCircleInviteResponse(invite),
	})
}

// RevokeInvite godoc
// @Summary      Revoke a circle invite
// @Tags         circle-invites
// @Produce      json
// @Param        id path string true "Circle ID"
// @Param        inviteId path string true "Invite ID"
// @Success      200 {object} dto.WebResponse[dto.CircleInviteResponse]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /circles/{id}/invites/{inviteId}/revoke [post]
func (h *CircleInviteHandler) RevokeInvite(c fiber.Ctx) error {
	circleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}
	inviteID, err := uuid.Parse(c.Params("inviteId"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	invite, err := h.usecase.RevokeInvite(c, inviteID, circleID, userID)
	if err != nil {
		return err
	}

	slog.InfoContext(c, "circle invite revoked", "invite_id", inviteID, "circle_id", circleID, "user_id", userID)

	return c.JSON(dto.WebResponse[dto.CircleInviteResponse]{
		Code:    fiber.StatusOK,
		Message: "circle invite revoked",
		Data:    dto.NewCircleInviteResponse(invite),
	})
}

// GetInviteLink godoc
// @Summary      Get a circle's active invite link
// @Tags         circle-invites
// @Produce      json
// @Param        id path string true "Circle ID"
// @Success      200 {object} dto.WebResponse[dto.CircleInviteResponse]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /circles/{id}/invite-link [get]
func (h *CircleInviteHandler) GetInviteLink(c fiber.Ctx) error {
	circleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	invite, err := h.usecase.GetInviteLink(c, circleID, userID)
	if err != nil {
		return err
	}

	return c.JSON(dto.WebResponse[dto.CircleInviteResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    dto.NewCircleInviteResponse(invite),
	})
}

// CreateOrRotateInviteLink godoc
// @Summary      Create or rotate a circle's invite link
// @Description  The old link (if any) stops working the moment the new one exists; the raw token is only ever returned here
// @Tags         circle-invites
// @Accept       json
// @Produce      json
// @Param        id      path string                                 true "Circle ID"
// @Param        request body dto.CreateOrRotateInviteLinkRequest true "Link options"
// @Success      201 {object} dto.WebResponse[dto.CircleInviteLinkResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /circles/{id}/invite-link [post]
func (h *CircleInviteHandler) CreateOrRotateInviteLink(c fiber.Ctx) error {
	circleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	var req dto.CreateOrRotateInviteLinkRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	result, err := h.usecase.CreateOrRotateInviteLink(c, usecase.CreateOrRotateInviteLinkInput{
		CircleID: circleID, UserID: userID, RequiresApproval: req.RequiresApproval,
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c, "circle invite link rotated", "circle_id", circleID, "user_id", userID)

	return c.Status(fiber.StatusCreated).JSON(dto.WebResponse[dto.CircleInviteLinkResponse]{
		Code:    fiber.StatusCreated,
		Message: "circle invite link created",
		Data: dto.CircleInviteLinkResponse{
			Invite: dto.NewCircleInviteResponse(result.Invite),
			Token:  result.RawToken,
		},
	})
}

// SetInviteLinkRequiresApproval godoc
// @Summary      Toggle whether a circle's invite link requires approval
// @Tags         circle-invites
// @Accept       json
// @Produce      json
// @Param        id      path string                                    true "Circle ID"
// @Param        request body dto.SetInviteLinkRequiresApprovalRequest true "New approval requirement"
// @Success      200 {object} dto.WebResponse[dto.CircleInviteResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /circles/{id}/invite-link/approval [patch]
func (h *CircleInviteHandler) SetInviteLinkRequiresApproval(c fiber.Ctx) error {
	circleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	var req dto.SetInviteLinkRequiresApprovalRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	invite, err := h.usecase.SetInviteLinkRequiresApproval(c, circleID, userID, req.RequiresApproval)
	if err != nil {
		return err
	}

	slog.InfoContext(c, "circle invite link approval requirement updated", "circle_id", circleID, "user_id", userID)

	return c.JSON(dto.WebResponse[dto.CircleInviteResponse]{
		Code:    fiber.StatusOK,
		Message: "circle invite link updated",
		Data:    dto.NewCircleInviteResponse(invite),
	})
}
