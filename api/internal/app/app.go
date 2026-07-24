package app

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/config"
	"github.com/ngodingvareng/memoria/internal/delivery/rest"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/handler"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/middleware"
	"github.com/ngodingvareng/memoria/internal/repository"
	"github.com/ngodingvareng/memoria/internal/usecase"
	slogfiber "github.com/samber/slog-fiber"
)

type Container struct {
	App *fiber.App
	DB  *pgxpool.Pool
}

func NewContainer(cfg *config.Config) (*Container, error) {
	// 1. DB
	conn, err := pgxpool.New(context.Background(), cfg.GetDSN())
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(context.Background()); err != nil {
		return nil, err
	}

	// 2. Wiring: repository -> usecase -> handler
	activityRepo := repository.NewActivityRepository(conn)
	activityUsecase := usecase.NewActivityUsecase(activityRepo)
	activityHandler := handler.NewActivityHandler(activityUsecase)

	// 3. Fiber app + global middleware.
	// Renamed the local var from "app" to "fiberApp" — this file's own
	// package is named "app", and reusing that name for a variable here
	// reads confusingly once there are several fiberApp.Use(...) calls.
	fiberApp := fiber.New(fiber.Config{
		ErrorHandler: middleware.CustomErrorHandler,
	})

	// Access log for every request, structured via slog (see main.go for
	// the JSON handler setup). recover.New() should sit right after it so
	// a panic anywhere downstream still gets logged and turned into a
	// normal 500 instead of crashing the whole process.
	fiberApp.Use(slogfiber.New(slog.Default()))
	fiberApp.Use(recover.New())

	// 4. Router
	rest.SetupRoutes(fiberApp, rest.Handlers{
		Activity: activityHandler,
	})

	return &Container{
		App: fiberApp,
		DB:  conn,
	}, nil

}
