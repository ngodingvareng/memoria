package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

// RefreshCookieName is shared with the auth handler (which sets/clears
// this same cookie on /auth/login, /auth/refresh, /auth/logout) —
// defined once here to avoid the two packages drifting out of sync on
// the cookie name. Unlike the old SessionCookieName, this cookie only
// ever travels to /auth/* (see the handler's cookie Path), never to
// protected routes — RequireAuth below doesn't read it at all.
const RefreshCookieName = "refresh_token"

const userIDLocalsKey = "user_id"

// RequireAuth validates the Authorization: Bearer <access_token> header
// against the given issuer — this is what replaces the userID :=
// uuid.New() placeholder that's been sitting in ActivityHandler since
// the very first version of the create-activity feature.
//
// Unlike the old session-cookie RequireAuth, this never touches the
// database: the access token is a stateless, self-contained JWT, so a
// valid signature + unexpired exp claim is enough on its own.
func RequireAuth(issuer usecase.AccessTokenIssuer) fiber.Handler {
	return func(c fiber.Ctx) error {
		const bearerPrefix = "Bearer "

		header := c.Get(fiber.HeaderAuthorization)
		if !strings.HasPrefix(header, bearerPrefix) {
			return errs.ErrUnauthorized
		}

		userID, err := issuer.Parse(strings.TrimPrefix(header, bearerPrefix))
		if err != nil {
			if errors.Is(err, errs.ErrTokenExpired) {
				return errs.ErrTokenExpired
			}
			return errs.ErrUnauthorized
		}

		c.Locals(userIDLocalsKey, userID)
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
