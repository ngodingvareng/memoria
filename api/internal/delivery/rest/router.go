package rest

import (
	"github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	_ "github.com/ngodingvareng/memoria/docs"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/handler"
)

// Handlers groups every handler the router needs. Add a field here each
// time a new domain's handler is wired up in app.go.
type Handlers struct {
	Activity      *handler.ActivityHandler
	ActivityImage *handler.ActivityImageHandler
}

func SetupRoutes(app *fiber.App, h Handlers) {
	app.Get("/docs/*", swaggo.New(swaggo.Config{Title: "Book API"}))

	app.Get("/swagger", func(c fiber.Ctx) error {
		return c.Redirect().To("/docs")
	})

	app.Get("/swagger.json", func(c fiber.Ctx) error {
		return c.SendFile("./docs/swagger.json")
	})

	activities := app.Group("/activities")
	activities.Post("/", h.Activity.CreateActivity)

	activities.Post("/:id/images", h.ActivityImage.UploadActivityImage)
	activities.Get("/:id/images", h.ActivityImage.ListActivityImages)
	activities.Delete("/:id/images/:imageId", h.ActivityImage.DeleteActivityImage)

}
