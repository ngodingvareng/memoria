package handler

import (
	"bytes"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"

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

// sniffLen mirrors the number of bytes net/http.DetectContentType
// actually inspects — reading more would just waste memory.
const sniffLen = 512

// allowedImageContentTypes is the sniffed-content allowlist. Client-
// supplied Content-Type headers are never trusted (see below); this is
// checked against the type net/http.DetectContentType derives from the
// file's actual magic bytes.
var allowedImageContentTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// detectImageContentType sniffs the real content type from the file's
// magic bytes rather than trusting the client-supplied Content-Type
// header (which is attacker-controlled and easily spoofed — e.g. a
// ".jpg" upload that's actually an HTML/SVG payload). It returns a
// reader that replays the sniffed prefix followed by the rest of the
// file, so the sniff doesn't consume bytes the caller still needs to
// read.
func detectImageContentType(file multipart.File) (contentType string, body io.Reader, err error) {
	buf := make([]byte, sniffLen)
	n, err := io.ReadFull(file, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return "", nil, err
	}
	buf = buf[:n]
	return http.DetectContentType(buf), io.MultiReader(bytes.NewReader(buf), file), nil
}

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

	// Sniff the real content type from the file bytes instead of
	// trusting the client-supplied header (fileHeader.Header.Get
	// ("Content-Type") is attacker-controlled and easily spoofed).
	detectedType, body, err := detectImageContentType(file)
	if err != nil {
		return errs.New(fiber.StatusBadRequest, "could not read uploaded file")
	}
	if !allowedImageContentTypes[detectedType] {
		return errs.New(fiber.StatusBadRequest, "uploaded file is not a supported image type")
	}

	image, err := h.usecase.UploadThreadImage(c, usecase.UploadThreadImageInput{
		ThreadID:    threadID,
		UserID:      userID,
		FileName:    fileHeader.Filename,
		ContentType: detectedType,
		Size:        fileHeader.Size,
		Body:        body,
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
