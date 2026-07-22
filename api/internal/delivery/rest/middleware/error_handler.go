package middleware

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
	"github.com/ngodingvareng/memoria/pkg/errs"
)

func CustomErrorHandler(c fiber.Ctx, err error) error {
	// Default to 500 Internal Server Error
	code := fiber.StatusInternalServerError
	message := "Internal server error"

	if valErr, ok := errors.AsType[*errs.ValidationError](err); ok {
		return c.Status(fiber.StatusBadRequest).JSON(dto.WebResponse[any]{
			Code:    fiber.StatusBadRequest,
			Message: "Validation failed",
			Errors:  valErr.Errors,
		})
	}

	if appErr, ok := errors.AsType[*errs.AppError](err); ok {
		code = appErr.Code
		message = appErr.Message
	} else {
		if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
			code = fiberErr.Code
			message = fiberErr.Message
		} else {
			log.Printf("[SERVER ERROR]: %v\n", err)
		}
	}

	// Return the standardized JSON response
	return c.Status(code).JSON(dto.WebResponse[any]{
		Code:    code,
		Message: message,
		Errors:  err.Error(), // Optional: show error details during development (remove for production)
	})
}
