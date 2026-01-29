package router

import (
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/handler"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/middleware"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, h *handler.AllOfHandler, mw *middleware.Middleware) {
	api := app.Group("/api")

	HealthCheckRoutes(api, h)
	AuthRoutes(api, h, mw)
	EventRoutes(api, h, mw)
}
