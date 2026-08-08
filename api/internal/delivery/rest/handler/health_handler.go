package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/dto"
)

// HealthHandler backs /healthz for container orchestration (Docker
// healthcheck, Kubernetes liveness/readiness probes). It pings the
// database so a container that's up but can't reach Postgres is
// reported unhealthy instead of accepting traffic it can't serve.
type HealthHandler struct {
	db *pgxpool.Pool
}

func NewHealthHandler(db *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{db: db}
}

// Health godoc
// @Summary      Health check
// @Tags         health
// @Produce      json
// @Success      200 {object} dto.WebResponse[any]
// @Failure      503 {object} dto.WebResponse[any]
// @Router       /healthz [get]
func (h *HealthHandler) Health(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c, 2*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(dto.WebResponse[any]{
			Code:    fiber.StatusServiceUnavailable,
			Message: "unhealthy",
			Errors:  err.Error(),
		})
	}

	return c.JSON(dto.WebResponse[any]{
		Code:    fiber.StatusOK,
		Message: "ok",
	})
}
