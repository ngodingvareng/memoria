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

type ActivityHandler struct {
	usecase  usecase.ActivityUsecase
	validate *validator.Validate
}

func NewActivityHandler(usecase usecase.ActivityUsecase) *ActivityHandler {
	return &ActivityHandler{
		usecase:  usecase,
		validate: validator.New(),
	}
}

// CreateActivity godoc
// @Summary      Create a new activity
// @Description  Create a new activity for the authenticated user
// @Tags         activities
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateActivityRequest true "Create activity request"
// @Success      201 {object} dto.WebResponse[dto.ActivityResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      500 {object} dto.WebResponse[any]
// @Router       /activities [post]
func (h *ActivityHandler) CreateActivity(c fiber.Ctx) error {
	var req dto.CreateActivityRequest
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

	response, err := h.usecase.CreateActivity(c, usecase.CreateActivityInput{
		UserID:                     userID,
		Name:                       req.Name,
		Description:                req.Description,
		IsFixedSchedule:            req.IsFixedSchedule,
		ColorHex:                   req.ColorHex,
		ConfirmationTimeoutMinutes: req.ConfirmationTimeoutMinutes,
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c, "activity created",
		"activity_id", response.ID,
		"user_id", response.UserID,
	)

	return c.Status(fiber.StatusCreated).JSON(dto.WebResponse[dto.ActivityResponse]{
		Code:    fiber.StatusCreated,
		Message: "activity created",
		Data:    dto.NewActivityResponse(response),
	})

}

// UpdateActivity godoc
// @Summary      Update an activity
// @Description  Full-representation update of an activity owned by the authenticated user
// @Tags         activities
// @Accept       json
// @Produce      json
// @Param        id      path string                     true "Activity ID"
// @Param        request body dto.UpdateActivityRequest true "Update activity request"
// @Success      200 {object} dto.WebResponse[dto.ActivityResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      404 {object} dto.WebResponse[any]
// @Failure      500 {object} dto.WebResponse[any]
// @Router       /activities/{id} [put]
func (h *ActivityHandler) UpdateActivity(c fiber.Ctx) error {
	activityID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	var req dto.UpdateActivityRequest
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

	response, err := h.usecase.UpdateActivity(c, usecase.UpdateActivityInput{
		ID:                         activityID,
		UserID:                     userID,
		Name:                       req.Name,
		Description:                req.Description,
		ColorHex:                   req.ColorHex,
		ConfirmationTimeoutMinutes: req.ConfirmationTimeoutMinutes,
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c, "activity updated",
		"activity_id", response.ID,
		"user_id", response.UserID,
	)

	return c.JSON(dto.WebResponse[dto.ActivityResponse]{
		Code:    fiber.StatusOK,
		Message: "activity updated",
		Data:    dto.NewActivityResponse(response),
	})
}

// DeleteActivity godoc
// @Summary      Delete an activity
// @Description  Soft-deletes an activity owned by the authenticated user
// @Tags         activities
// @Param        id path string true "Activity ID"
// @Success      204
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      500 {object} dto.WebResponse[any]
// @Router       /activities/{id} [delete]
func (h *ActivityHandler) DeleteActivity(c fiber.Ctx) error {
	activityID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	if err := h.usecase.SoftDeleteActivity(c, activityID, userID); err != nil {
		return err
	}

	slog.InfoContext(c, "activity soft-deleted",
		"activity_id", activityID,
		"user_id", userID,
	)

	return c.SendStatus(fiber.StatusNoContent)
}

// GetActivity godoc
// @Summary      Get an activity by ID
// @Description  Get a single activity belonging to the authenticated user
// @Tags         activities
// @Produce      json
// @Param        id path string true "Activity ID"
// @Success      200 {object} dto.WebResponse[dto.ActivityResponse]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /activities/{id} [get]
func (h *ActivityHandler) GetActivity(c fiber.Ctx) error {
	activityID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	activity, err := h.usecase.GetActivity(c, userID, activityID)
	if err != nil {
		return err
	}

	return c.JSON(dto.WebResponse[dto.ActivityResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    dto.NewActivityResponse(activity),
	})
}

// SearchActivities godoc
// @Summary      Search / list activities
// @Description  Search and filter the authenticated user's activities, with pagination
// @Tags         activities
// @Produce      json
// @Param        name              query string false "Filter by name (partial, case-insensitive)"
// @Param        is_fixed_schedule query bool   false "Filter by fixed-schedule flag"
// @Param        page              query int    false "Page number (default 1)"
// @Param        page_size         query int    false "Items per page (default 20, max 100)"
// @Success      200 {object} dto.WebResponse[dto.SearchActivitiesResponse]
// @Router       /activities [get]
func (h *ActivityHandler) SearchActivities(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	var query dto.SearchActivitiesQuery
	if err := c.Bind().Query(&query); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(query); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	result, err := h.usecase.SearchActivities(c, usecase.SearchActivitiesInput{
		UserID:          userID,
		Name:            query.Name,
		IsFixedSchedule: query.IsFixedSchedule,
		Page:            query.Page,
		PageSize:        query.PageSize,
	})
	if err != nil {
		return err
	}

	responses := make([]dto.ActivityResponse, len(result.Activities))
	for i, a := range result.Activities {
		responses[i] = dto.NewActivityResponse(a)
	}

	return c.JSON(dto.WebResponse[dto.SearchActivitiesResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data: dto.SearchActivitiesResponse{
			Activities: responses,
			Pagination: dto.PaginationResponse{
				Page:     result.Page,
				PageSize: result.PageSize,
				Total:    result.Total,
			},
		},
	})
}
