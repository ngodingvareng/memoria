package handler

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/middleware"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
	"github.com/ngodingvareng/memoria/internal/validate"
)

type SearchHandler struct {
	usecase  usecase.SearchUsecase
	validate *validator.Validate
}

func NewSearchHandler(usecase usecase.SearchUsecase) *SearchHandler {
	return &SearchHandler{
		usecase:  usecase,
		validate: validator.New(),
	}
}

// Suggestions godoc
// @Summary      Live search suggestions
// @Description  Top-5 matching Moments and top-5 matching Threads for the authenticated user, for a live-typing search popover
// @Tags         search
// @Produce      json
// @Param        q query string true "Search text"
// @Success      200 {object} dto.WebResponse[dto.SearchSuggestionsResponse]
// @Failure      400 {object} dto.WebResponse[any]
// @Router       /search/suggestions [get]
func (h *SearchHandler) Suggestions(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return errs.ErrUnauthorized
	}

	var query dto.SearchSuggestionsQuery
	if err := c.Bind().Query(&query); err != nil {
		return errs.ErrInvalidInput
	}
	if err := h.validate.Struct(query); err != nil {
		return &errs.ValidationError{Errors: validate.FormatValidationErrors(err)}
	}

	result, err := h.usecase.SearchSuggestions(c, usecase.SearchSuggestionsInput{
		UserID: userID,
		Query:  query.Query,
	})
	if err != nil {
		return err
	}

	moments := make([]dto.MomentSuggestionResponse, len(result.Moments))
	for i, m := range result.Moments {
		moments[i] = dto.NewMomentSuggestionResponse(m)
	}
	threads := make([]dto.ThreadSuggestionResponse, len(result.Threads))
	for i, t := range result.Threads {
		threads[i] = dto.NewThreadSuggestionResponse(t)
	}

	return c.JSON(dto.WebResponse[dto.SearchSuggestionsResponse]{
		Code:    fiber.StatusOK,
		Message: "success",
		Data: dto.SearchSuggestionsResponse{
			Moments: moments,
			Threads: threads,
		},
	})
}
