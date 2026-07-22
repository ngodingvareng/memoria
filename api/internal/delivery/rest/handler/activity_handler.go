package handler

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/moez-rd/memoria/internal/delivery/rest/dto"
	"github.com/moez-rd/memoria/internal/usecase"
	"github.com/moez-rd/memoria/pkg/errs"
	"github.com/moez-rd/memoria/pkg/util"
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
		return &errs.ValidationError{Errors: util.FormatValidationErrors(err)}
	}

	// TODO: replace with the authenticated user's id once auth middleware
	// is wired in, e.g.:
	//   userID, ok := c.Locals("user_id").(uuid.UUID)
	//   if !ok { return errs.ErrUnauthorized }
	userID := uuid.New()

	response, err := h.usecase.CreateActivity(c, usecase.CreateActivityInput{
		UserID:                     userID,
		Name:                       req.Name,
		Description:                req.Description,
		IsFixedSchedule:            req.IsFixedSchedule,
		ColorPalette:               req.ColorPalette,
		ConfirmationTimeoutMinutes: req.ConfirmationTimeoutMinutes,
	})
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(dto.WebResponse[dto.ActivityResponse]{
		Code:    fiber.StatusCreated,
		Message: "activity created",
		Data:    dto.NewActivityResponse(response),
	})

}
