package handler

import (
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/middleware"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
	"github.com/ngodingvareng/memoria/internal/validate"
)

type NotificationHandler struct {
	usecase  usecase.NotificationUsecase
	validate *validator.Validate
}

func NewNotificationHandler(uc usecase.NotificationUsecase) *NotificationHandler {
	return &NotificationHandler{usecase: uc, validate: validator.New()}
}

// ListNotifications godoc
// @Summary      List the caller's notifications
// @Description  The in-app feed, newest first, plus the header badge's unread count
// @Tags         notifications
// @Produce      json
// @Param        page      query int false "Page number"
// @Param        page_size query int false "Page size"
// @Success      200 {object} dto.WebResponse[dto.ListNotificationsResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /notifications [get].
func (h *NotificationHandler) ListNotifications(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	var query dto.ListNotificationsQuery
	if err := c.Bind().Query(&query); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(query); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	result, err := h.usecase.ListNotifications(c, usecase.ListNotificationsInput{
		UserID:   userID,
		Page:     query.Page,
		PageSize: query.PageSize,
	})
	if err != nil {
		return err
	}

	return c.JSON(dto.WebResponse[dto.ListNotificationsResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    dto.NewListNotificationsResponse(result),
	})
}

// MarkNotificationRead godoc
// @Summary      Mark one notification read
// @Tags         notifications
// @Param        id path string true "Notification ID"
// @Success      204
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /notifications/{id}/read [patch].
func (h *NotificationHandler) MarkNotificationRead(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	if err := h.usecase.MarkRead(c, id, userID); err != nil {
		return err
	}

	slog.InfoContext(c, "notification marked read", "notification_id", id, "user_id", userID)

	return c.SendStatus(fiber.StatusNoContent)
}

// MarkAllNotificationsRead godoc
// @Summary      Mark every visible notification read
// @Tags         notifications
// @Success      204
// @Router       /notifications/read-all [post].
func (h *NotificationHandler) MarkAllNotificationsRead(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	if err := h.usecase.MarkAllRead(c, userID); err != nil {
		return err
	}

	slog.InfoContext(c, "all notifications marked read", "user_id", userID)

	return c.SendStatus(fiber.StatusNoContent)
}

// GetNotificationPreferences godoc
// @Summary      Get the caller's notification preferences
// @Tags         notifications
// @Produce      json
// @Success      200 {object} dto.WebResponse[dto.NotificationPreferencesResponse]
// @Router       /notifications/preferences [get].
func (h *NotificationHandler) GetNotificationPreferences(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	pref, err := h.usecase.GetPreferences(c, userID)
	if err != nil {
		return err
	}

	return c.JSON(dto.WebResponse[dto.NotificationPreferencesResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    dto.NewNotificationPreferencesResponse(pref),
	})
}

// UpdateNotificationPreferences godoc
// @Summary      Update the caller's notification preferences
// @Description  Every type is togglable independently, plus quiet hours (FEATURES.md, Notification Preferences)
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        request body dto.UpdateNotificationPreferencesRequest true "New preferences"
// @Success      200 {object} dto.WebResponse[dto.NotificationPreferencesResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /notifications/preferences [put].
func (h *NotificationHandler) UpdateNotificationPreferences(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	var req dto.UpdateNotificationPreferencesRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(req); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	quietHoursStart, err := parseWallClock(req.QuietHoursStart)
	if err != nil {
		return errs.ErrInvalidInput
	}
	quietHoursEnd, err := parseWallClock(req.QuietHoursEnd)
	if err != nil {
		return errs.ErrInvalidInput
	}

	updated, err := h.usecase.UpdatePreferences(c, usecase.UpdateNotificationPreferencesInput{
		UserID:                           userID,
		CommitmentUpcomingEnabled:        req.CommitmentUpcomingEnabled,
		MomentDueEnabled:                 req.MomentDueEnabled,
		MomentMissedEnabled:              req.MomentMissedEnabled,
		EchoReadyEnabled:                 req.EchoReadyEnabled,
		RecapGeneratedEnabled:            req.RecapGeneratedEnabled,
		CircleInviteReceivedEnabled:      req.CircleInviteReceivedEnabled,
		CircleJoinRequestReceivedEnabled: req.CircleJoinRequestReceivedEnabled,
		MentionedInMomentEnabled:         req.MentionedInMomentEnabled,
		AddedToCircleEnabled:             req.AddedToCircleEnabled,
		CircleJoinRequestApprovedEnabled: req.CircleJoinRequestApprovedEnabled,
		ResponseDigestEnabled:            req.ResponseDigestEnabled,
		ResponseDigestHour:               req.ResponseDigestHour,
		QuietHoursStart:                  quietHoursStart,
		QuietHoursEnd:                    quietHoursEnd,
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c, "notification preferences updated", "user_id", userID)

	return c.JSON(dto.WebResponse[dto.NotificationPreferencesResponse]{
		Code:    fiber.StatusOK,
		Message: "notification preferences updated",
		Data:    dto.NewNotificationPreferencesResponse(updated),
	})
}

// parseWallClock parses an "HH:MM:SS" string into a *time.Time carrying
// only that wall-clock component (see pgTimeToPtr's mirror on the
// repository side). Returns nil, nil for a nil input.
func parseWallClock(s *string) (*time.Time, error) {
	if s == nil {
		return nil, nil
	}
	parsed, err := time.Parse("15:04:05", *s)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
