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

type CircleJoinRequestHandler struct {
	usecase  usecase.CircleJoinRequestUsecase
	validate *validator.Validate
}

func NewCircleJoinRequestHandler(usecase usecase.CircleJoinRequestUsecase) *CircleJoinRequestHandler {
	return &CircleJoinRequestHandler{usecase: usecase, validate: validator.New()}
}

// FollowInviteLink godoc
// @Summary      Follow a circle invite link
// @Description  Joins immediately if the link doesn't require approval, otherwise opens a join request
// @Tags         circle-join-requests
// @Accept       json
// @Produce      json
// @Param        request body dto.FollowInviteLinkRequest true "Invite link token"
// @Success      200 {object} dto.WebResponse[dto.FollowInviteLinkResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /circle-join-requests [post]
func (h *CircleJoinRequestHandler) FollowInviteLink(c fiber.Ctx) error {
	var req dto.FollowInviteLinkRequest
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

	result, err := h.usecase.FollowInviteLink(c, req.Token, userID)
	if err != nil {
		return err
	}

	slog.InfoContext(c, "circle invite link followed", "user_id", userID)

	response := dto.FollowInviteLinkResponse{}
	if result.Joined != nil {
		memberResponse := dto.NewCircleMemberResponse(result.Joined)
		response.Member = &memberResponse
	}
	if result.JoinRequest != nil {
		joinRequestResponse := dto.NewCircleJoinRequestResponse(result.JoinRequest)
		response.JoinRequest = &joinRequestResponse
	}

	return c.JSON(dto.WebResponse[dto.FollowInviteLinkResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    response,
	})
}

// ListPending godoc
// @Summary      List a circle's pending join requests
// @Description  The approval queue; requires the caller to have invite privileges
// @Tags         circle-join-requests
// @Produce      json
// @Param        id        path  string true  "Circle ID"
// @Param        page      query int    false "Page number (default 1)"
// @Param        page_size query int    false "Requests per page (default 20, max 100)"
// @Success      200 {object} dto.WebResponse[dto.ListCircleJoinRequestsResponse]
// @Failure      403 {object} dto.WebResponse[any]
// @Router       /circles/{id}/join-requests [get]
func (h *CircleJoinRequestHandler) ListPending(c fiber.Ctx) error {
	circleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	var query dto.ListMomentsQuery
	if err := c.Bind().Query(&query); err != nil {
		return errs.ErrInvalidInput
	}

	requests, err := h.usecase.ListPending(c, circleID, userID, query.Page, query.PageSize)
	if err != nil {
		return err
	}

	responses := make([]dto.CircleJoinRequestResponse, len(requests))
	for i, r := range requests {
		responses[i] = dto.NewCircleJoinRequestResponse(r)
	}

	return c.JSON(dto.WebResponse[dto.ListCircleJoinRequestsResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    dto.ListCircleJoinRequestsResponse{JoinRequests: responses},
	})
}

// Approve godoc
// @Summary      Approve a circle join request
// @Tags         circle-join-requests
// @Produce      json
// @Param        id        path string true "Circle ID"
// @Param        requestId path string true "Join request ID"
// @Success      200 {object} dto.WebResponse[dto.CircleMemberResponse]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /circles/{id}/join-requests/{requestId}/approve [post]
func (h *CircleJoinRequestHandler) Approve(c fiber.Ctx) error {
	circleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}
	requestID, err := uuid.Parse(c.Params("requestId"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	member, err := h.usecase.Approve(c, requestID, circleID, userID)
	if err != nil {
		return err
	}

	slog.InfoContext(c, "circle join request approved", "request_id", requestID, "circle_id", circleID, "decided_by_user_id", userID)

	return c.JSON(dto.WebResponse[dto.CircleMemberResponse]{
		Code:    fiber.StatusOK,
		Message: "circle join request approved",
		Data:    dto.NewCircleMemberResponse(member),
	})
}

// Reject godoc
// @Summary      Reject a circle join request
// @Tags         circle-join-requests
// @Produce      json
// @Param        id        path string true "Circle ID"
// @Param        requestId path string true "Join request ID"
// @Success      200 {object} dto.WebResponse[dto.CircleJoinRequestResponse]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /circles/{id}/join-requests/{requestId}/reject [post]
func (h *CircleJoinRequestHandler) Reject(c fiber.Ctx) error {
	circleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}
	requestID, err := uuid.Parse(c.Params("requestId"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	request, err := h.usecase.Reject(c, requestID, circleID, userID)
	if err != nil {
		return err
	}

	slog.InfoContext(c, "circle join request rejected", "request_id", requestID, "circle_id", circleID, "decided_by_user_id", userID)

	return c.JSON(dto.WebResponse[dto.CircleJoinRequestResponse]{
		Code:    fiber.StatusOK,
		Message: "circle join request rejected",
		Data:    dto.NewCircleJoinRequestResponse(request),
	})
}

// Cancel godoc
// @Summary      Cancel my own circle join request
// @Tags         circle-join-requests
// @Produce      json
// @Param        id path string true "Join request ID"
// @Success      200 {object} dto.WebResponse[dto.CircleJoinRequestResponse]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /circle-join-requests/{id}/cancel [post]
func (h *CircleJoinRequestHandler) Cancel(c fiber.Ctx) error {
	requestID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	request, err := h.usecase.Cancel(c, requestID, userID)
	if err != nil {
		return err
	}

	slog.InfoContext(c, "circle join request cancelled", "request_id", requestID, "user_id", userID)

	return c.JSON(dto.WebResponse[dto.CircleJoinRequestResponse]{
		Code:    fiber.StatusOK,
		Message: "circle join request cancelled",
		Data:    dto.NewCircleJoinRequestResponse(request),
	})
}
