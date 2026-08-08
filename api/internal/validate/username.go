package validate

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// UsernameFormat mirrors users.chk_users_username_format exactly —
// lowercase letters, digits, underscore, and dot, 3-30 characters.
var UsernameFormat = regexp.MustCompile(`^[a-z0-9_.]{3,30}$`)

// RegisterUsernameValidator registers the "username" tag on v, shared by
// every handler that validates a username field (AuthHandler,
// UserHandler) so the format regex lives in exactly one place.
func RegisterUsernameValidator(v *validator.Validate) {
	_ = v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		return UsernameFormat.MatchString(fl.Field().String())
	})
}
