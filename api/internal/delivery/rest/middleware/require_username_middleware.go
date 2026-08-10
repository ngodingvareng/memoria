package middleware

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
)

// UserGetter is the minimal capability RequireUsername needs — kept
// narrow (rather than depending on the full usecase.UserRepository) so
// this middleware only couples to the one method it actually calls.
// usecase.UserRepository already satisfies this.
type UserGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
}

// RequireUsername rejects the request with errs.ErrUsernameRequired
// unless the authenticated user has already claimed a username. Must
// run after RequireAuth, which is what populates the user id this reads
// from context.
//
// Unlike RequireAuth, this does hit the database every request — a
// username, once claimed, can't be un-claimed, so it doesn't belong in
// the stateless JWT (see AccessTokenIssuer's own reasoning for keeping
// claims minimal), and there's no cache here in front of it.
func RequireUsername(users UserGetter) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, ok := UserIDFromContext(c)
		if !ok {
			return errs.ErrUnauthorized
		}

		user, err := users.GetByID(c, userID)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				return errs.ErrUnauthorized
			}
			return err
		}
		if user.Username == nil {
			return errs.ErrUsernameRequired
		}
		return c.Next()
	}
}
