package handler

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/middleware"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

type MomentImageHandler struct {
	usecase usecase.MomentImageUsecase
}

func NewMomentImageHandler(uc usecase.MomentImageUsecase) *MomentImageHandler {
	return &MomentImageHandler{usecase: uc}
}

// UploadMomentImage godoc
// @Summary      Upload an image for a moment
// @Tags         moments
// @Accept       multipart/form-data
// @Produce      json
// @Param        id     path     string true  "Moment ID"
// @Param        image  formData file   true  "Image file"
// @Success      201 {object} dto.WebResponse[dto.MomentImageResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      500 {object} dto.WebResponse[any]
// @Router       /moments/{id}/images [post]
func (h *MomentImageHandler) UploadMomentImage(c fiber.Ctx) error {
	momentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

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

	// Sniff the real content type from the file bytes instead of
	// trusting the client-supplied header — see detectImageContentType's
	// comment in thread_image_handler.go.
	detectedType, body, err := detectImageContentType(file)
	if err != nil {
		return errs.New(fiber.StatusBadRequest, "could not read uploaded file")
	}
	if !allowedImageContentTypes[detectedType] {
		return errs.New(fiber.StatusBadRequest, "uploaded file is not a supported image type")
	}

	image, err := h.usecase.UploadMomentImage(c, usecase.UploadMomentImageInput{
		MomentID:    momentID,
		UserID:      userID,
		FileName:    fileHeader.Filename,
		ContentType: detectedType,
		Size:        fileHeader.Size,
		Body:        body,
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c, "moment image uploaded", "moment_id", momentID, "image_id", image.ID)

	return c.Status(fiber.StatusCreated).JSON(dto.WebResponse[dto.MomentImageResponse]{
		Code:    fiber.StatusCreated,
		Message: "image uploaded",
		Data:    dto.NewMomentImageResponse(*image),
	})
}

// ListMomentImages godoc
// @Summary      List images for a moment
// @Tags         moments
// @Produce      json
// @Param        id path string true "Moment ID"
// @Success      200 {object} dto.WebResponse[[]dto.MomentImageResponse]
// @Router       /moments/{id}/images [get]
func (h *MomentImageHandler) ListMomentImages(c fiber.Ctx) error {
	momentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	images, err := h.usecase.ListMomentImages(c, momentID, userID)
	if err != nil {
		return err
	}

	responses := make([]dto.MomentImageResponse, len(images))
	for i, img := range images {
		responses[i] = dto.NewMomentImageResponse(img)
	}

	return c.JSON(dto.WebResponse[[]dto.MomentImageResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    responses,
	})
}

// DeleteMomentImage godoc
// @Summary      Delete a moment image
// @Tags         moments
// @Param        id      path string true "Moment ID"
// @Param        imageId path string true "Image ID"
// @Success      204
// @Router       /moments/{id}/images/{imageId} [delete]
func (h *MomentImageHandler) DeleteMomentImage(c fiber.Ctx) error {
	momentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}
	imageID, err := uuid.Parse(c.Params("imageId"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	if err := h.usecase.DeleteMomentImage(c, momentID, imageID, userID); err != nil {
		return err
	}

	slog.InfoContext(c, "moment image deleted", "moment_id", momentID, "image_id", imageID)

	return c.SendStatus(fiber.StatusNoContent)
}
