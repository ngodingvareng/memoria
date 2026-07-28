package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

// SessionCookieName is shared with the auth handler (which sets/clears
// this same cookie) — defined once here to avoid the two packages
// drifting out of sync on the cookie name.
const SessionCookieName = "session"

const userIDLocalsKey = "user_id"

// RequireAuth validates the session cookie and stores the authenticated
// user's id in c.Locals — this is what replaces the userID := uuid.New()
// placeholder that's been sitting in ActivityHandler since the very
// first version of the create-activity feature.
func RequireAuth(authUsecase usecase.AuthUsecase) fiber.Handler {
	return func(c fiber.Ctx) error {
		token := c.Cookies(SessionCookieName)
		if token == "" {
			return errs.ErrUnauthorized
		}

		user, err := authUsecase.ValidateSession(c, token)
		if err != nil {
			return err
		}

		c.Locals(userIDLocalsKey, user.ID)
		return c.Next()
	}
}

// UserIDFromContext retrieves the user id RequireAuth stored for this
// request. ok is false if RequireAuth wasn't applied to this route (a
// programming error — a protected handler being reachable without the
// middleware) or, in principle, if the stored value was ever the wrong
// type, though RequireAuth itself only ever stores a uuid.UUID.
func UserIDFromContext(c fiber.Ctx) (uuid.UUID, bool) {
	id, ok := c.Locals(userIDLocalsKey).(uuid.UUID)
	return id, ok
}
