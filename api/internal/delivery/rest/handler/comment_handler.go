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

type CommentHandler struct {
	usecase  usecase.CommentUsecase
	validate *validator.Validate
}

func NewCommentHandler(usecase usecase.CommentUsecase) *CommentHandler {
	return &CommentHandler{usecase: usecase, validate: validator.New()}
}

// CreateComment godoc
// @Summary      Comment on a moment
// @Description  A flat contextual discussion space under a Moment — never threaded
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        id      path string                    true "Moment ID"
// @Param        request body dto.CreateCommentRequest true "Comment body and audience"
// @Success      201 {object} dto.WebResponse[dto.CommentResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      403 {object} dto.WebResponse[any]
// @Router       /moments/{id}/comments [post]
func (h *CommentHandler) CreateComment(c fiber.Ctx) error {
	momentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	var req dto.CreateCommentRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(req); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	circleID, err := parseOptionalUUID(req.CircleID)
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	comment, err := h.usecase.CreateComment(c, usecase.CreateCommentInput{
		MomentID: momentID, UserID: userID, CircleID: circleID, Body: req.Body,
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c, "comment created", "comment_id", comment.ID, "moment_id", momentID, "user_id", userID)

	return c.Status(fiber.StatusCreated).JSON(dto.WebResponse[dto.CommentResponse]{
		Code:    fiber.StatusCreated,
		Message: "comment created",
		Data:    dto.NewCommentResponse(comment),
	})
}

// ListComments godoc
// @Summary      List a moment's comments
// @Description  Every comment visible to the caller across every context they have access to
// @Tags         comments
// @Produce      json
// @Param        id path string true "Moment ID"
// @Success      200 {object} dto.WebResponse[dto.ListCommentsResponse]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /moments/{id}/comments [get]
func (h *CommentHandler) ListComments(c fiber.Ctx) error {
	momentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	comments, err := h.usecase.ListComments(c, momentID, userID)
	if err != nil {
		return err
	}

	responses := make([]dto.CommentResponse, len(comments))
	for i, cm := range comments {
		responses[i] = dto.NewCommentResponse(cm)
	}

	return c.JSON(dto.WebResponse[dto.ListCommentsResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    dto.ListCommentsResponse{Comments: responses},
	})
}

// UpdateComment godoc
// @Summary      Edit a comment
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        id        path string                    true "Moment ID"
// @Param        commentId path string                    true "Comment ID"
// @Param        request   body dto.UpdateCommentRequest true "New body"
// @Success      200 {object} dto.WebResponse[dto.CommentResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /moments/{id}/comments/{commentId} [put]
func (h *CommentHandler) UpdateComment(c fiber.Ctx) error {
	commentID, err := uuid.Parse(c.Params("commentId"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	var req dto.UpdateCommentRequest
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

	comment, err := h.usecase.UpdateComment(c, usecase.UpdateCommentInput{ID: commentID, UserID: userID, Body: req.Body})
	if err != nil {
		return err
	}

	slog.InfoContext(c, "comment updated", "comment_id", comment.ID, "user_id", userID)

	return c.JSON(dto.WebResponse[dto.CommentResponse]{
		Code:    fiber.StatusOK,
		Message: "comment updated",
		Data:    dto.NewCommentResponse(comment),
	})
}

// DeleteComment godoc
// @Summary      Delete a comment
// @Tags         comments
// @Param        id        path string true "Moment ID"
// @Param        commentId path string true "Comment ID"
// @Success      204
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /moments/{id}/comments/{commentId} [delete]
func (h *CommentHandler) DeleteComment(c fiber.Ctx) error {
	commentID, err := uuid.Parse(c.Params("commentId"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	if err := h.usecase.SoftDeleteComment(c, commentID, userID); err != nil {
		return err
	}

	slog.InfoContext(c, "comment deleted", "comment_id", commentID, "user_id", userID)

	return c.SendStatus(fiber.StatusNoContent)
}

// GetAudience godoc
// @Summary      Get which contexts the caller may comment/react in for a moment
// @Description  Whether the mention context is available, plus which Circle audiences (FEATURES.md, Response) — drives the comment/reaction composer
// @Tags         comments
// @Produce      json
// @Param        id path string true "Moment ID"
// @Success      200 {object} dto.WebResponse[dto.MomentAudienceResponse]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /moments/{id}/audience [get]
func (h *CommentHandler) GetAudience(c fiber.Ctx) error {
	momentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	audience, err := h.usecase.GetAudience(c, momentID, userID)
	if err != nil {
		return err
	}

	return c.JSON(dto.WebResponse[dto.MomentAudienceResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    dto.NewMomentAudienceResponse(audience),
	})
}
