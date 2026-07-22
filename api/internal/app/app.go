package app

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/config"
	"github.com/ngodingvareng/memoria/internal/delivery/rest"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/handler"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/middleware"
	"github.com/ngodingvareng/memoria/internal/repository"
	"github.com/ngodingvareng/memoria/internal/usecase"
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

	// pgxpool.New is lazy (doesn't actually connect), so ping here to
	// fail fast at startup instead of on the first real request.
	if err := conn.Ping(context.Background()); err != nil {
		return nil, err
	}

	// 2. Wiring: repository -> usecase -> handler
	activityRepo := repository.NewActivityRepository(conn)
	activityUsecase := usecase.NewActivityUsecase(activityRepo)
	activityHandler := handler.NewActivityHandler(activityUsecase)

	// 3. Router
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.CustomErrorHandler,
	})

	rest.SetupRoutes(app, rest.Handlers{
		Activity: activityHandler,
	})

	return &Container{
		App: app,
		DB:  conn,
	}, nil
}
