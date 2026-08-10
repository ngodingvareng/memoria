package middleware

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
	"github.com/ngodingvareng/memoria/internal/errs"
)

// NewErrorHandler builds the fiber.ErrorHandler used for every response.
// debug gates whether the raw internal error text (err.Error(), which can
// include wrapped internal call-path and DB driver details) is included in
// the JSON body — leave it false in production and opt in via DEBUG_MODE
// only for local/dev debugging.
func NewErrorHandler(debug bool) fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
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
		} else if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
			code = fiberErr.Code
			message = fiberErr.Message
		} else {
			// This is the one and only place an unhandled error gets logged.
			// usecase and repository deliberately don't log — they just
			// return (optionally wrapped) errors that bubble up to here.
			// Logging at every layer would print the same failure multiple
			// times as it climbs back up the call stack.
			slog.ErrorContext(c, "unhandled error",
				"error", err,
				"method", c.Method(),
				"path", c.Path(),
			)
		}

		var details any
		if debug {
			details = err.Error()
		}

		return c.Status(code).JSON(dto.WebResponse[any]{
			Code:    code,
			Message: message,
			Errors:  details,
		})
	}
}
