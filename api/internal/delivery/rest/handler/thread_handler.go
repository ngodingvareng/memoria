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

type ThreadHandler struct {
	usecase  usecase.ThreadUsecase
	validate *validator.Validate
}

func NewThreadHandler(usecase usecase.ThreadUsecase) *ThreadHandler {
	return &ThreadHandler{
		usecase:  usecase,
		validate: validator.New(),
	}
}

// CreateThread godoc
// @Summary      Create a new thread
// @Description  Create a new thread for the authenticated user
// @Tags         threads
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateThreadRequest true "Create thread request"
// @Success      201 {object} dto.WebResponse[dto.ThreadResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      500 {object} dto.WebResponse[any]
// @Router       /threads [post]
func (h *ThreadHandler) CreateThread(c fiber.Ctx) error {
	var req dto.CreateThreadRequest
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

	var circleID *uuid.UUID
	if req.CircleID != nil {
		parsed, err := uuid.Parse(*req.CircleID)
		if err != nil {
			return errs.ErrInvalidInput
		}
		circleID = &parsed
	}

	response, err := h.usecase.CreateThread(c, usecase.CreateThreadInput{
		UserID:      userID,
		CircleID:    circleID,
		Name:        req.Name,
		Description: req.Description,
		ColorHex:    req.ColorHex,
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c, "thread created",
		"thread_id", response.ID,
		"user_id", response.UserID,
	)

	return c.Status(fiber.StatusCreated).JSON(dto.WebResponse[dto.ThreadResponse]{
		Code:    fiber.StatusCreated,
		Message: "thread created",
		Data:    dto.NewThreadResponse(response),
	})

}

// UpdateThread godoc
// @Summary      Update an thread
// @Description  Full-representation update of an thread owned by the authenticated user
// @Tags         threads
// @Accept       json
// @Produce      json
// @Param        id      path string                     true "Thread ID"
// @Param        request body dto.UpdateThreadRequest true "Update thread request"
// @Success      200 {object} dto.WebResponse[dto.ThreadResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      404 {object} dto.WebResponse[any]
// @Failure      500 {object} dto.WebResponse[any]
// @Router       /threads/{id} [put]
func (h *ThreadHandler) UpdateThread(c fiber.Ctx) error {
	threadID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	var req dto.UpdateThreadRequest
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

	response, err := h.usecase.UpdateThread(c, usecase.UpdateThreadInput{
		ID:          threadID,
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		ColorHex:    req.ColorHex,
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c, "thread updated",
		"thread_id", response.ID,
		"user_id", response.UserID,
	)

	return c.JSON(dto.WebResponse[dto.ThreadResponse]{
		Code:    fiber.StatusOK,
		Message: "thread updated",
		Data:    dto.NewThreadResponse(response),
	})
}

// DeleteThread godoc
// @Summary      Delete an thread
// @Description  Soft-deletes an thread owned by the authenticated user
// @Tags         threads
// @Param        id path string true "Thread ID"
// @Success      204
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      500 {object} dto.WebResponse[any]
// @Router       /threads/{id} [delete]
func (h *ThreadHandler) DeleteThread(c fiber.Ctx) error {
	threadID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	if err := h.usecase.SoftDeleteThread(c, threadID, userID); err != nil {
		return err
	}

	slog.InfoContext(c, "thread soft-deleted",
		"thread_id", threadID,
		"user_id", userID,
	)

	return c.SendStatus(fiber.StatusNoContent)
}

// GetThread godoc
// @Summary      Get an thread by ID
// @Description  Get a single thread belonging to the authenticated user
// @Tags         threads
// @Produce      json
// @Param        id path string true "Thread ID"
// @Success      200 {object} dto.WebResponse[dto.ThreadResponse]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /threads/{id} [get]
func (h *ThreadHandler) GetThread(c fiber.Ctx) error {
	threadID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	thread, err := h.usecase.GetThread(c, userID, threadID)
	if err != nil {
		return err
	}

	return c.JSON(dto.WebResponse[dto.ThreadResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    dto.NewThreadResponse(thread),
	})
}

// SearchThreads godoc
// @Summary      Search / list threads
// @Description  Search and filter the authenticated user's threads, with pagination
// @Tags         threads
// @Produce      json
// @Param        name           query string false "Fuzzy/typo-tolerant match against the thread's name and description"
// @Param        archived       query bool   false "Filter by whether the thread is archived"
// @Param        sort           query string false "Sort order; only 'updated_at' is supported (switches to most-recently-updated first)"
// @Param        page           query int    false "Page number (default 1)"
// @Param        page_size      query int    false "Threads per page (default 20, max 100)"
// @Success      200 {object} dto.WebResponse[dto.SearchThreadsResponse]
// @Router       /threads [get]
func (h *ThreadHandler) SearchThreads(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	var query dto.SearchThreadsQuery
	if err := c.Bind().Query(&query); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(query); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	result, err := h.usecase.SearchThreads(c, usecase.SearchThreadsInput{
		UserID:        userID,
		Name:          query.Name,
		Archived:      query.Archived,
		SortByRecency: query.Sort != nil && *query.Sort == "updated_at",
		Page:          query.Page,
		PageSize:      query.PageSize,
	})
	if err != nil {
		return err
	}

	responses := make([]dto.ThreadResponse, len(result.Threads))
	for i, a := range result.Threads {
		responses[i] = dto.NewThreadResponse(a)
	}

	return c.JSON(dto.WebResponse[dto.SearchThreadsResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data: dto.SearchThreadsResponse{
			Threads: responses,
			Pagination: dto.PaginationResponse{
				Page:     result.Page,
				PageSize: result.PageSize,
				Total:    result.Total,
			},
		},
	})
}

// ListCircleThreads godoc
// @Summary      List a Circle's collaborative threads
// @Description  List every Thread owned by a Circle; the caller must be an active member
// @Tags         threads
// @Produce      json
// @Param        id path string true "Circle ID"
// @Success      200 {object} dto.WebResponse[dto.ListThreadsResponse]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /circles/{id}/threads [get]
func (h *ThreadHandler) ListCircleThreads(c fiber.Ctx) error {
	circleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	threads, err := h.usecase.ListCircleThreads(c, circleID, userID)
	if err != nil {
		return err
	}

	responses := make([]dto.ThreadResponse, len(threads))
	for i, a := range threads {
		responses[i] = dto.NewThreadResponse(a)
	}

	return c.JSON(dto.WebResponse[dto.ListThreadsResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    dto.ListThreadsResponse{Threads: responses},
	})
}
