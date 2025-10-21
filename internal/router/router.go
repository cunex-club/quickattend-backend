package router

import (
	"github.com/cunex-club/quickattend-backend/internal/handler"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, h *handler.AllOfHandler) {
	api := app.Group("/api")

	HealthCheckRoutes(api, h)
}
