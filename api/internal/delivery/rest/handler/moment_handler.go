package handler

import (
	"log/slog"
	"time"

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

type MomentHandler struct {
	usecase  usecase.MomentUsecase
	validate *validator.Validate
}

func NewMomentHandler(usecase usecase.MomentUsecase) *MomentHandler {
	return &MomentHandler{
		usecase:  usecase,
		validate: validator.New(),
	}
}

// parseOptionalUUID mirrors the c.Params("id") parse pattern used
// throughout, but for an optional *string request field already
// confirmed to be a well-formed UUID by struct validation.
func parseOptionalUUID(s *string) (*uuid.UUID, error) {
	if s == nil {
		return nil, nil
	}
	id, err := uuid.Parse(*s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// CreateMoment godoc
// @Summary      Capture a new moment
// @Description  Create a new manually-captured Moment for the authenticated user
// @Tags         moments
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateMomentRequest true "Create moment request"
// @Success      201 {object} dto.WebResponse[dto.MomentResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      500 {object} dto.WebResponse[any]
// @Router       /moments [post].
func (h *MomentHandler) CreateMoment(c fiber.Ctx) error {
	var req dto.CreateMomentRequest
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

	occurredAt, err := time.Parse(time.RFC3339, req.OccurredAt)
	if err != nil {
		return errs.ErrInvalidInput
	}

	threadID, err := parseOptionalUUID(req.ThreadID)
	if err != nil {
		return errs.ErrInvalidInput
	}

	clientID, err := parseOptionalUUID(req.ClientID)
	if err != nil {
		return errs.ErrInvalidInput
	}

	var recordedAt *time.Time
	if req.RecordedAt != nil {
		t, err := time.Parse(time.RFC3339, *req.RecordedAt)
		if err != nil {
			return errs.ErrInvalidInput
		}
		recordedAt = &t
	}

	origin := enum.MomentOriginManual
	if req.Origin != nil {
		origin = enum.MomentOrigin(*req.Origin)
	}

	response, err := h.usecase.CreateMoment(c, usecase.CreateMomentInput{
		UserID:                   userID,
		ThreadID:                 threadID,
		Origin:                   origin,
		OccurredAt:               occurredAt,
		OccurredUTCOffsetMinutes: req.OccurredUTCOffsetMinutes,
		RecordedAt:               recordedAt,
		Note:                     req.Note,
		ColorHex:                 req.ColorHex,
		PlaceName:                req.PlaceName,
		Latitude:                 req.Latitude,
		Longitude:                req.Longitude,
		ClientID:                 clientID,
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c, "moment created", "moment_id", response.ID, "user_id", response.UserID)

	return c.Status(fiber.StatusCreated).JSON(dto.WebResponse[dto.MomentResponse]{
		Code:    fiber.StatusCreated,
		Message: "moment created",
		Data:    dto.NewMomentResponse(response),
	})
}

// UpdateMoment godoc
// @Summary      Update a moment
// @Description  Full-representation update of a Moment owned by the authenticated user
// @Tags         moments
// @Accept       json
// @Produce      json
// @Param        id      path string                     true "Moment ID"
// @Param        request body dto.UpdateMomentRequest true "Update moment request"
// @Success      200 {object} dto.WebResponse[dto.MomentResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      404 {object} dto.WebResponse[any]
// @Failure      500 {object} dto.WebResponse[any]
// @Router       /moments/{id} [put].
func (h *MomentHandler) UpdateMoment(c fiber.Ctx) error {
	momentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	var req dto.UpdateMomentRequest
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

	occurredAt, err := time.Parse(time.RFC3339, req.OccurredAt)
	if err != nil {
		return errs.ErrInvalidInput
	}

	threadID, err := parseOptionalUUID(req.ThreadID)
	if err != nil {
		return errs.ErrInvalidInput
	}

	response, err := h.usecase.UpdateMoment(c, usecase.UpdateMomentInput{
		ID:                       momentID,
		UserID:                   userID,
		ThreadID:                 threadID,
		OccurredAt:               occurredAt,
		OccurredUTCOffsetMinutes: req.OccurredUTCOffsetMinutes,
		Note:                     req.Note,
		ColorHex:                 req.ColorHex,
		PlaceName:                req.PlaceName,
		Latitude:                 req.Latitude,
		Longitude:                req.Longitude,
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c, "moment updated", "moment_id", response.ID, "user_id", response.UserID)

	return c.JSON(dto.WebResponse[dto.MomentResponse]{
		Code:    fiber.StatusOK,
		Message: "moment updated",
		Data:    dto.NewMomentResponse(response),
	})
}

// DeleteMoment godoc
// @Summary      Delete a moment
// @Description  Soft-deletes a Moment owned by the authenticated user
// @Tags         moments
// @Param        id path string true "Moment ID"
// @Success      204
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      500 {object} dto.WebResponse[any]
// @Router       /moments/{id} [delete].
func (h *MomentHandler) DeleteMoment(c fiber.Ctx) error {
	momentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	if err := h.usecase.SoftDeleteMoment(c, momentID, userID); err != nil {
		return err
	}

	slog.InfoContext(c, "moment soft-deleted", "moment_id", momentID, "user_id", userID)

	return c.SendStatus(fiber.StatusNoContent)
}

// GetMoment godoc
// @Summary      Get a moment by ID
// @Description  Get a single Moment reachable by the caller — its owner, an active mentioned user, or a Circle context it's visible in
// @Tags         moments
// @Produce      json
// @Param        id path string true "Moment ID"
// @Success      200 {object} dto.WebResponse[dto.MomentResponse]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /moments/{id} [get].
func (h *MomentHandler) GetMoment(c fiber.Ctx) error {
	momentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	moment, err := h.usecase.GetMoment(c, userID, momentID)
	if err != nil {
		return err
	}

	return c.JSON(dto.WebResponse[dto.MomentResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    dto.NewMomentResponse(moment),
	})
}

func newListMomentsResponse(result *usecase.MomentListResult) dto.ListMomentsResponse {
	responses := make([]dto.MomentResponse, len(result.Moments))
	for i, m := range result.Moments {
		responses[i] = dto.NewMomentResponse(m)
	}
	return dto.ListMomentsResponse{
		Moments: responses,
		Pagination: dto.PaginationResponse{
			Page:     result.Page,
			PageSize: result.PageSize,
		},
	}
}

// ListMoments godoc
// @Summary      List moments
// @Description  The authenticated user's home feed — their own Moments, every Moment in a Circle they belong to, and every Moment they're mentioned in — newest occurrence first
// @Tags         moments
// @Produce      json
// @Param        page      query int false "Page number (default 1)"
// @Param        page_size query int false "Moments per page (default 20, max 100)"
// @Success      200 {object} dto.WebResponse[dto.ListMomentsResponse]
// @Router       /moments [get].
func (h *MomentHandler) ListMoments(c fiber.Ctx) error {
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

	result, err := h.usecase.ListMoments(c, usecase.ListMomentsInput{
		UserID:   userID,
		Page:     query.Page,
		PageSize: query.PageSize,
	})
	if err != nil {
		return err
	}

	return c.JSON(dto.WebResponse[dto.ListMomentsResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    newListMomentsResponse(result),
	})
}

// SearchMoments godoc
// @Summary      Search moments
// @Description  Global text Search over the authenticated user's own Moments
// @Tags         moments
// @Produce      json
// @Param        q         query string true  "Search text"
// @Param        page      query int    false "Page number (default 1)"
// @Param        page_size query int    false "Moments per page (default 20, max 100)"
// @Success      200 {object} dto.WebResponse[dto.ListMomentsResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /moments/search [get].
func (h *MomentHandler) SearchMoments(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	var query dto.SearchMomentsQuery
	if err := c.Bind().Query(&query); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(query); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	result, err := h.usecase.SearchMoments(c, usecase.SearchMomentsInput{
		UserID:   userID,
		Query:    query.Query,
		Page:     query.Page,
		PageSize: query.PageSize,
	})
	if err != nil {
		return err
	}

	return c.JSON(dto.WebResponse[dto.ListMomentsResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    newListMomentsResponse(result),
	})
}

// ListCircleMoments godoc
// @Summary      List a circle's moments
// @Description  The Circle's Album feed — Moments from its collaborative Threads plus Moments shared into it, newest occurrence first
// @Tags         moments
// @Produce      json
// @Param        id        path  string true  "Circle ID"
// @Param        page      query int    false "Page number (default 1)"
// @Param        page_size query int    false "Moments per page (default 20, max 100)"
// @Success      200 {object} dto.WebResponse[dto.ListMomentsResponse]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /circles/{id}/moments [get].
func (h *MomentHandler) ListCircleMoments(c fiber.Ctx) error {
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
	if err := h.validate.Struct(query); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	result, err := h.usecase.ListCircleMoments(c, usecase.ListCircleMomentsInput{
		UserID:   userID,
		CircleID: circleID,
		Page:     query.Page,
		PageSize: query.PageSize,
	})
	if err != nil {
		return err
	}

	return c.JSON(dto.WebResponse[dto.ListMomentsResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    newListMomentsResponse(result),
	})
}

// ListThreadMoments godoc
// @Summary      List a thread's moments
// @Description  A single Thread's browsable timeline, newest occurrence first
// @Tags         moments
// @Produce      json
// @Param        id        path  string true  "Thread ID"
// @Param        page      query int    false "Page number (default 1)"
// @Param        page_size query int    false "Moments per page (default 20, max 100)"
// @Success      200 {object} dto.WebResponse[dto.ListMomentsResponse]
// @Router       /threads/{id}/moments [get].
func (h *MomentHandler) ListThreadMoments(c fiber.Ctx) error {
	threadID, err := uuid.Parse(c.Params("id"))
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
	if err := h.validate.Struct(query); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	result, err := h.usecase.ListThreadMoments(c, usecase.ListThreadMomentsInput{
		UserID:   userID,
		ThreadID: threadID,
		Page:     query.Page,
		PageSize: query.PageSize,
	})
	if err != nil {
		return err
	}

	return c.JSON(dto.WebResponse[dto.ListMomentsResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    newListMomentsResponse(result),
	})
}
