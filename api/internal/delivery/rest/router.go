package rest

import (
	"github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	_ "github.com/ngodingvareng/memoria/docs"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/handler"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/middleware"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

// Handlers groups every handler the router needs. Add a field here each
// time a new domain's handler is wired up in app.go.
type Handlers struct {
	Auth          *handler.AuthHandler
	Activity      *handler.ActivityHandler
	ActivityImage *handler.ActivityImageHandler
}

// SetupRoutes wires every route. authRateLimiter is built in app.go
// (where the LOGIN_RATE_LIMIT_* config lives) and applied only to the
// unauthenticated auth endpoints that are actually brute-forceable —
// /refresh and /logout require a valid cookie/session already, so they
// aren't password-guessing targets the way /login and /register are.
func SetupRoutes(app *fiber.App, issuer usecase.AccessTokenIssuer, authRateLimiter fiber.Handler, h Handlers) {
	app.Get("/docs/*", swaggo.New(swaggo.Config{Title: "Book API"}))

	app.Get("/swagger", func(c fiber.Ctx) error {
		return c.Redirect().To("/docs")
	})

	app.Get("/swagger.json", func(c fiber.Ctx) error {
		return c.SendFile("./docs/swagger.json")
	})

	auth := app.Group("/auth")
	auth.Post("/register", authRateLimiter, h.Auth.Register)
	auth.Post("/login", authRateLimiter, h.Auth.Login)
	auth.Post("/refresh", h.Auth.Refresh)
	auth.Post("/logout", h.Auth.Logout)

	activities := app.Group("/activities", middleware.RequireAuth(issuer))
	activities.Post("/", h.Activity.CreateActivity)
	activities.Put("/:id", h.Activity.UpdateActivity)
	activities.Delete("/:id", h.Activity.DeleteActivity)
	activities.Get("/", h.Activity.SearchActivities)
	activities.Get("/:id", h.Activity.GetActivity)

	activities.Post("/:id/images", h.ActivityImage.UploadActivityImage)
	activities.Get("/:id/images", h.ActivityImage.ListActivityImages)
	activities.Delete("/:id/images/:imageId", h.ActivityImage.DeleteActivityImage)
}
