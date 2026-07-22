package util

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

// ErrorResponse describes a single field-level validation failure.
//
// NOTE: pkg/errs.ValidationError already references `util.ErrorResponse`
// with a comment "assuming this is your util return type" — meaning it
// wasn't finalized on your end either. If a real ErrorResponse type
// already exists elsewhere in pkg/util, delete this struct and keep only
// FormatValidationErrors below, adjusting field names to match it.
type ErrorResponse struct {
	Field   string `json:"field" example:"name"`
	Tag     string `json:"tag" example:"required"`
	Message string `json:"message" example:"name is required"`
}

// FormatValidationErrors converts a github.com/go-playground/validator
// error into a slice suitable for pkg/errs.ValidationError.Errors. It
// returns nil if err isn't a validator.ValidationErrors (e.g. it was some
// other kind of error), so callers should confirm validation actually
// failed before relying on a non-empty result.
func FormatValidationErrors(err error) []*ErrorResponse {
	verrs, ok := errors.AsType[validator.ValidationErrors](err)
	if !ok {
		return nil
	}

	out := make([]*ErrorResponse, 0, len(verrs))
	for _, fe := range verrs {
		out = append(out, &ErrorResponse{
			Field:   fe.Field(),
			Tag:     fe.Tag(),
			Message: formatFieldMessage(fe),
		})
	}
	return out
}

func formatFieldMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "min":
		return fe.Field() + " must be at least " + fe.Param() + " characters"
	case "max":
		return fe.Field() + " must be at most " + fe.Param() + " characters"
	case "gt":
		return fe.Field() + " must be greater than " + fe.Param()
	case "hexcolor":
		return fe.Field() + " must be a valid hex color (e.g. #RRGGBB)"
	default:
		return fe.Field() + " is invalid"
	}
}
