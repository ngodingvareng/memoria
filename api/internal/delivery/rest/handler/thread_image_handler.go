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

// maxImageUploadSize caps a single image upload independently of the
// global fiber.Config{BodyLimit: ...} — set that global limit too (see
// note in router.go / app.go), this is a second, per-endpoint check.
const maxImageUploadSize = 10 * 1024 * 1024 // 10 MB

type ThreadImageHandler struct {
	usecase usecase.ThreadImageUsecase
}

func NewThreadImageHandler(uc usecase.ThreadImageUsecase) *ThreadImageHandler {
	return &ThreadImageHandler{usecase: uc}
}

// UploadThreadImage godoc
// @Summary      Upload an image for an thread
// @Tags         threads
// @Accept       multipart/form-data
// @Produce      json
// @Param        id     path     string true  "Thread ID"
// @Param        image  formData file   true  "Image file"
// @Success      201 {object} dto.WebResponse[dto.ThreadImageResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      500 {object} dto.WebResponse[any]
// @Router       /threads/{id}/images [post]
func (h *ThreadImageHandler) UploadThreadImage(c fiber.Ctx) error {
	threadID, err := uuid.Parse(c.Params("id"))
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

	image, err := h.usecase.UploadThreadImage(c, usecase.UploadThreadImageInput{
		ThreadID:    threadID,
		UserID:      userID,
		FileName:    fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		Size:        fileHeader.Size,
		Body:        file,
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c, "thread image uploaded", "thread_id", threadID, "image_id", image.ID)

	return c.Status(fiber.StatusCreated).JSON(dto.WebResponse[dto.ThreadImageResponse]{
		Code:    fiber.StatusCreated,
		Message: "image uploaded",
		Data:    dto.NewThreadImageResponse(*image),
	})
}

// ListThreadImages godoc
// @Summary      List images for an thread
// @Tags         threads
// @Produce      json
// @Param        id path string true "Thread ID"
// @Success      200 {object} dto.WebResponse[[]dto.ThreadImageResponse]
// @Router       /threads/{id}/images [get]
func (h *ThreadImageHandler) ListThreadImages(c fiber.Ctx) error {
	threadID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	images, err := h.usecase.ListThreadImages(c, threadID, userID)
	if err != nil {
		return err
	}

	responses := make([]dto.ThreadImageResponse, len(images))
	for i, img := range images {
		responses[i] = dto.NewThreadImageResponse(img)
	}

	return c.JSON(dto.WebResponse[[]dto.ThreadImageResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    responses,
	})
}

// DeleteThreadImage godoc
// @Summary      Delete an thread image
// @Tags         threads
// @Param        id      path string true "Thread ID"
// @Param        imageId path string true "Image ID"
// @Success      204
// @Router       /threads/{id}/images/{imageId} [delete]
func (h *ThreadImageHandler) DeleteThreadImage(c fiber.Ctx) error {
	threadID, err := uuid.Parse(c.Params("id"))
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

	if err := h.usecase.DeleteThreadImage(c, threadID, imageID, userID); err != nil {
		return err
	}

	slog.InfoContext(c, "thread image deleted", "thread_id", threadID, "image_id", imageID)

	return c.SendStatus(fiber.StatusNoContent)
}
