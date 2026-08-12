package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ngodingvareng/memoria/internal/app"
	"github.com/ngodingvareng/memoria/internal/config"
)

// @title Memoria API
// @version 1.0
// @BasePath /api/v1.
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	container, err := app.NewContainer(cfg)
	if err != nil {
		slog.Error("failed to initialize app", "error", err)
		os.Exit(1)
	}

	serveErr := make(chan error, 1)
	go func() {
		port := ":" + cfg.ServerPort
		slog.Info("server starting", "port", port)
		if err := container.App.Listen(port); err != nil {
			serveErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serveErr:
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	case <-quit:
		slog.Info("shutting down gracefully")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := container.App.ShutdownWithContext(ctx); err != nil {
		slog.Error("fiber forced to shutdown", "error", err)
	}

	container.DB.Close()
	slog.Info("database connection closed, exiting")
}
