package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/moez-rd/memoria/internal/db"
)

func main() {
	app := fiber.New()

	// Runs first on every request
	app.Use(func(c fiber.Ctx) error {
		c.Set("X-Powered-By", "Fiber")
		return c.Next()
	})

	// Serve files from ./public
	app.Use("/", static.New("./public"))

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Get("/users/:id", func(c fiber.Ctx) error {
		return c.SendString("User ID: " + c.Params("id"))
	})

	app.Get("/api/users", func(c fiber.Ctx) error {
		return c.JSON([]fiber.Map{
			{"id": 1, "name": "Alice"},
			{"id": 2, "name": "Bob"},
		})
	})

	app.Get("/admin", func(c fiber.Ctx) error {
		return fiber.ErrForbidden
	})

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, "user=pqgotest dbname=pqgotest sslmode=verify-full")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	queries := db.New(conn)

	_, err = queries.CreateAuthor(ctx, db.CreateAuthorParams{
		Name: "Brian",
		Bio:  pgtype.Text{String: ""},
	})

	log.Fatal(app.Listen(":3000"))
}
