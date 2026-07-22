package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ngodingvareng/memoria/internal/app"
	"github.com/ngodingvareng/memoria/internal/config"
)

// @title Book API
// @version 1.0
// @BasePath /api/v1
func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	container, err := app.NewContainer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	go func() {
		port := ":" + cfg.ServerPort
		log.Printf("Server starting on port %s", port)
		if err := container.App.Listen(port); err != nil {
			log.Printf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down gracefully...")

	if err := container.App.Shutdown(); err != nil {
		log.Printf("Fiber forced to shutdown: %v", err)
	}

	container.DB.Close()
	log.Println("Database connection closed. Exiting.")
}
