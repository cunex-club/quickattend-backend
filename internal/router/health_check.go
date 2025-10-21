package router

import (
	"github.com/cunex-club/quickattend-backend/internal/handler"
	"github.com/gofiber/fiber/v2"
)

func HealthCheckRoutes(r fiber.Router, h *handler.AllOfHandler) {
	health := r.Group("/health-check")

	health.Get("/", h.HealthCheckHandler.HealthCheck)
}
