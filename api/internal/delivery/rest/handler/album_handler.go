package handler

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/middleware"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
	"github.com/ngodingvareng/memoria/internal/validate"
)

type AlbumHandler struct {
	usecase  usecase.AlbumUsecase
	validate *validator.Validate
}

func NewAlbumHandler(usecase usecase.AlbumUsecase) *AlbumHandler {
	return &AlbumHandler{usecase: usecase, validate: validator.New()}
}

// ListAlbum godoc
// @Summary      List the personal Album
// @Description  Every photo attachment from every Moment the authenticated user owns, newest occurrence first (FEATURES.md, Looking Back)
// @Tags         album
// @Produce      json
// @Param        page      query int false "Page number (default 1)"
// @Param        page_size query int false "Images per page (default 20, max 100)"
// @Success      200 {object} dto.WebResponse[dto.ListAlbumResponse]
// @Router       /album [get].
func (h *AlbumHandler) ListAlbum(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	var query dto.ListAlbumQuery
	if err := c.Bind().Query(&query); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(query); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	result, err := h.usecase.ListAlbum(c, usecase.ListAlbumInput{
		UserID:   userID,
		Page:     query.Page,
		PageSize: query.PageSize,
	})
	if err != nil {
		return err
	}

	return c.JSON(dto.WebResponse[dto.ListAlbumResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    dto.NewListAlbumResponse(result),
	})
}

// ListCircleAlbum godoc
// @Summary      List a circle's Album
// @Description  Every photo attachment from every Moment visible in the Circle — its own collaborative Threads plus Moments shared into it — newest occurrence first, with "from @username" attribution on shared-in images (FEATURES.md, Looking Back)
// @Tags         album
// @Produce      json
// @Param        id        path  string true  "Circle ID"
// @Param        page      query int    false "Page number (default 1)"
// @Param        page_size query int    false "Images per page (default 20, max 100)"
// @Success      200 {object} dto.WebResponse[dto.ListAlbumResponse]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /circles/{id}/album [get].
func (h *AlbumHandler) ListCircleAlbum(c fiber.Ctx) error {
	circleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	var query dto.ListAlbumQuery
	if err := c.Bind().Query(&query); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(query); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	result, err := h.usecase.ListCircleAlbum(c, usecase.ListCircleAlbumInput{
		UserID:   userID,
		CircleID: circleID,
		Page:     query.Page,
		PageSize: query.PageSize,
	})
	if err != nil {
		return err
	}

	return c.JSON(dto.WebResponse[dto.ListAlbumResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    dto.NewListAlbumResponse(result),
	})
}
