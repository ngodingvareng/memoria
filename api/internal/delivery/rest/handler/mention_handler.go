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

type MentionHandler struct {
	usecase  usecase.MentionUsecase
	validate *validator.Validate
}

func NewMentionHandler(usecase usecase.MentionUsecase) *MentionHandler {
	return &MentionHandler{usecase: usecase, validate: validator.New()}
}

// CreateMention godoc
// @Summary      Mention a user on a moment
// @Description  Names another user present in a personal Moment; automatically shares the Moment with them
// @Tags         mentions
// @Accept       json
// @Produce      json
// @Param        id      path string                       true "Moment ID"
// @Param        request body dto.CreateMentionRequest true "Username to mention"
// @Success      201 {object} dto.WebResponse[dto.MentionResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      403 {object} dto.WebResponse[any]
// @Router       /moments/{id}/mentions [post].
func (h *MentionHandler) CreateMention(c fiber.Ctx) error {
	momentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	var req dto.CreateMentionRequest
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

	result, err := h.usecase.CreateMention(c, usecase.CreateMentionInput{
		MomentID: momentID, OwnerUserID: userID, Username: req.Username,
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c, "mention created", "moment_id", momentID, "mention_id", result.Mention.ID, "user_id", userID)

	return c.Status(fiber.StatusCreated).JSON(dto.WebResponse[dto.MentionResponse]{
		Code:    fiber.StatusCreated,
		Message: "mention created",
		Data:    dto.NewCreateMentionResponse(result),
	})
}

// ListMentions godoc
// @Summary      List a moment's mentions
// @Tags         mentions
// @Produce      json
// @Param        id path string true "Moment ID"
// @Success      200 {object} dto.WebResponse[dto.ListMentionsResponse]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /moments/{id}/mentions [get].
func (h *MentionHandler) ListMentions(c fiber.Ctx) error {
	momentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	mentions, err := h.usecase.ListMentions(c, momentID, userID)
	if err != nil {
		return err
	}

	responses := make([]dto.MentionResponse, len(mentions))
	for i, m := range mentions {
		responses[i] = dto.NewMentionResponse(m)
	}

	return c.JSON(dto.WebResponse[dto.ListMentionsResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    dto.ListMentionsResponse{Mentions: responses},
	})
}

// DeleteMention godoc
// @Summary      Delete a mention (owner)
// @Description  The Moment owner removing a mention they added by mistake
// @Tags         mentions
// @Param        id         path string true "Moment ID"
// @Param        mentionId  path string true "Mention ID"
// @Success      204
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /moments/{id}/mentions/{mentionId} [delete].
func (h *MentionHandler) DeleteMention(c fiber.Ctx) error {
	momentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}
	mentionID, err := uuid.Parse(c.Params("mentionId"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	if err := h.usecase.DeleteMention(c, mentionID, momentID, userID); err != nil {
		return err
	}

	slog.InfoContext(c, "mention deleted", "moment_id", momentID, "mention_id", mentionID, "user_id", userID)

	return c.SendStatus(fiber.StatusNoContent)
}

// LeaveMention godoc
// @Summary      Leave a mention (mentioned user)
// @Description  The mentioned user removing themselves; the Moment stays with its owner, unchanged, and no notification is sent
// @Tags         mentions
// @Param        id path string true "Moment ID"
// @Success      204
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /moments/{id}/mentions/leave [post].
func (h *MentionHandler) LeaveMention(c fiber.Ctx) error {
	momentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	if err := h.usecase.LeaveMention(c, momentID, userID); err != nil {
		return err
	}

	slog.InfoContext(c, "mention left", "moment_id", momentID, "user_id", userID)

	return c.SendStatus(fiber.StatusNoContent)
}

// ListMentionedMoments godoc
// @Summary      List moments I'm mentioned in
// @Description  "Mentioned Moments" — every Moment the authenticated user has ever been mentioned in
// @Tags         mentions
// @Produce      json
// @Param        page      query int false "Page number (default 1)"
// @Param        page_size query int false "Moments per page (default 20, max 100)"
// @Success      200 {object} dto.WebResponse[dto.ListMomentsResponse]
// @Router       /mentions [get].
func (h *MentionHandler) ListMentionedMoments(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	var query dto.ListMomentsQuery
	if err := c.Bind().Query(&query); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(query); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	result, err := h.usecase.ListMentionedMoments(c, usecase.ListMentionedMomentsInput{
		UserID: userID, Page: query.Page, PageSize: query.PageSize,
	})
	if err != nil {
		return err
	}

	responses := make([]dto.MomentResponse, len(result.Moments))
	for i, m := range result.Moments {
		responses[i] = dto.NewMomentResponse(m)
	}

	return c.JSON(dto.WebResponse[dto.ListMomentsResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data: dto.ListMomentsResponse{
			Moments:    responses,
			Pagination: dto.PaginationResponse{Page: result.Page, PageSize: result.PageSize},
		},
	})
}

// SearchMentionableUsers godoc
// @Summary      Search users to mention
// @Description  Typeahead search over discoverable usernames the caller is currently allowed to mention
// @Tags         mentions
// @Produce      json
// @Param        q query string true "Username prefix"
// @Success      200 {object} dto.WebResponse[dto.SearchMentionableUsersResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /mentions/users [get].
func (h *MentionHandler) SearchMentionableUsers(c fiber.Ctx) error {
	var query dto.SearchMentionableUsersQuery
	if err := c.Bind().Query(&query); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(query); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	users, err := h.usecase.SearchMentionableUsers(c, userID, query.Query)
	if err != nil {
		return err
	}

	responses := make([]dto.PublicUserResponse, len(users))
	for i, u := range users {
		responses[i] = dto.NewPublicUserResponse(u)
	}

	return c.JSON(dto.WebResponse[dto.SearchMentionableUsersResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    dto.SearchMentionableUsersResponse{Users: responses},
	})
}

// ShareToCircle godoc
// @Summary      Share a moment to a circle
// @Description  The mention flow's optional "Share to circle too?" step — never a standalone share action
// @Tags         mentions
// @Accept       json
// @Produce      json
// @Param        id      path string                        true "Moment ID"
// @Param        request body dto.ShareMomentToCircleRequest true "Target circle"
// @Success      200 {object} dto.WebResponse[dto.MomentCircleResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /moments/{id}/share [post].
func (h *MentionHandler) ShareToCircle(c fiber.Ctx) error {
	momentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	var req dto.ShareMomentToCircleRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(req); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}
	circleID, err := uuid.Parse(req.CircleID)
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	share, err := h.usecase.ShareToCircle(c, usecase.ShareMomentToCircleInput{
		MomentID: momentID, CircleID: circleID, UserID: userID,
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c, "moment shared to circle", "moment_id", momentID, "circle_id", circleID, "user_id", userID)

	response := dto.WebResponse[dto.MomentCircleResponse]{Code: fiber.StatusOK, Message: "moment shared to circle"}
	if share != nil {
		response.Data = dto.NewMomentCircleResponse(share)
	}
	return c.JSON(response)
}

// UnshareFromCircle godoc
// @Summary      Unshare a moment from a circle
// @Tags         mentions
// @Param        id       path string true "Moment ID"
// @Param        circleId path string true "Circle ID"
// @Success      204
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /moments/{id}/share/{circleId} [delete].
func (h *MentionHandler) UnshareFromCircle(c fiber.Ctx) error {
	momentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}
	circleID, err := uuid.Parse(c.Params("circleId"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	if err := h.usecase.UnshareFromCircle(c, momentID, circleID, userID); err != nil {
		return err
	}

	slog.InfoContext(c, "moment unshared from circle", "moment_id", momentID, "circle_id", circleID, "user_id", userID)

	return c.SendStatus(fiber.StatusNoContent)
}

// ListShares godoc
// @Summary      List which circles a moment is shared to
// @Description  The "manage sharing" surface — every Circle this personal Moment is currently shared into
// @Tags         mentions
// @Produce      json
// @Param        id path string true "Moment ID"
// @Success      200 {object} dto.WebResponse[dto.ListMomentSharesResponse]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /moments/{id}/shares [get].
func (h *MentionHandler) ListShares(c fiber.Ctx) error {
	momentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	circleIDs, err := h.usecase.ListSharedCircles(c, momentID, userID)
	if err != nil {
		return err
	}

	ids := make([]string, len(circleIDs))
	for i, id := range circleIDs {
		ids[i] = id.String()
	}

	return c.JSON(dto.WebResponse[dto.ListMomentSharesResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    dto.ListMomentSharesResponse{CircleIDs: ids},
	})
}
