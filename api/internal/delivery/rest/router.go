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

func SetupRoutes(app *fiber.App, issuer usecase.AccessTokenIssuer, h Handlers) {
	app.Get("/docs/*", swaggo.New(swaggo.Config{Title: "Book API"}))

	app.Get("/swagger", func(c fiber.Ctx) error {
		return c.Redirect().To("/docs")
	})

	app.Get("/swagger.json", func(c fiber.Ctx) error {
		return c.SendFile("./docs/swagger.json")
	})

	auth := app.Group("/auth")
	auth.Post("/register", h.Auth.Register)
	auth.Post("/login", h.Auth.Login)
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
