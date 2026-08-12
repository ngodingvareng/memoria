package handler

import (
	"log/slog"

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

type ReactionHandler struct {
	usecase  usecase.ReactionUsecase
	validate *validator.Validate
}

func NewReactionHandler(usecase usecase.ReactionUsecase) *ReactionHandler {
	return &ReactionHandler{usecase: usecase, validate: validator.New()}
}

// SetReaction godoc
// @Summary      React to a moment
// @Description  Reacting again with a different kind just changes it; there is no separate update endpoint
// @Tags         reactions
// @Accept       json
// @Produce      json
// @Param        id      path string                  true "Moment ID"
// @Param        request body dto.SetReactionRequest true "Reaction kind and audience"
// @Success      200 {object} dto.WebResponse[dto.ReactionResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Failure      403 {object} dto.WebResponse[any]
// @Router       /moments/{id}/reactions [put].
func (h *ReactionHandler) SetReaction(c fiber.Ctx) error {
	momentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	var req dto.SetReactionRequest
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

	reaction, err := h.usecase.SetReaction(c, usecase.SetReactionInput{
		MomentID: momentID, UserID: userID, CircleID: circleID, Kind: enum.ReactionKind(req.Kind),
	})
	if err != nil {
		return err
	}

	slog.InfoContext(c, "reaction set", "reaction_id", reaction.ID, "moment_id", momentID, "user_id", userID)

	return c.JSON(dto.WebResponse[dto.ReactionResponse]{
		Code:    fiber.StatusOK,
		Message: "reaction set",
		Data:    dto.NewReactionResponse(reaction),
	})
}

// RemoveReaction godoc
// @Summary      Remove my reaction from a moment
// @Tags         reactions
// @Param        id       path  string true  "Moment ID"
// @Param        circle_id query string false "Circle audience (omit for the mention context)"
// @Success      204
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /moments/{id}/reactions [delete].
func (h *ReactionHandler) RemoveReaction(c fiber.Ctx) error {
	momentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	var query dto.RemoveReactionQuery
	if err := c.Bind().Query(&query); err != nil {
		return errs.ErrInvalidInput
	}
	circleID, err := parseOptionalUUID(query.CircleID)
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	if err := h.usecase.RemoveReaction(c, usecase.RemoveReactionInput{
		MomentID: momentID, UserID: userID, CircleID: circleID,
	}); err != nil {
		return err
	}

	slog.InfoContext(c, "reaction removed", "moment_id", momentID, "user_id", userID)

	return c.SendStatus(fiber.StatusNoContent)
}

// ListReactions godoc
// @Summary      List a moment's reactions
// @Description  Who reacted, never how many (FEATURES.md, Response)
// @Tags         reactions
// @Produce      json
// @Param        id path string true "Moment ID"
// @Success      200 {object} dto.WebResponse[dto.ListReactionsResponse]
// @Failure      404 {object} dto.WebResponse[any]
// @Router       /moments/{id}/reactions [get].
func (h *ReactionHandler) ListReactions(c fiber.Ctx) error {
	momentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errs.ErrInvalidInput
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	reactions, err := h.usecase.ListReactions(c, momentID, userID)
	if err != nil {
		return err
	}

	responses := make([]dto.ReactionResponse, len(reactions))
	for i, r := range reactions {
		responses[i] = dto.NewReactionResponse(r)
	}

	return c.JSON(dto.WebResponse[dto.ListReactionsResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data:    dto.ListReactionsResponse{Reactions: responses},
	})
}
