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
	Auth        *handler.AuthHandler
	Thread      *handler.ThreadHandler
	ThreadImage *handler.ThreadImageHandler
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

	threads := app.Group("/threads", middleware.RequireAuth(issuer))
	threads.Post("/", h.Thread.CreateThread)
	threads.Put("/:id", h.Thread.UpdateThread)
	threads.Delete("/:id", h.Thread.DeleteThread)
	threads.Get("/", h.Thread.SearchThreads)
	threads.Get("/:id", h.Thread.GetThread)

	threads.Post("/:id/images", h.ThreadImage.UploadThreadImage)
	threads.Get("/:id/images", h.ThreadImage.ListThreadImages)
	threads.Delete("/:id/images/:imageId", h.ThreadImage.DeleteThreadImage)
}
