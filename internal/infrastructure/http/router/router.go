package router

import (
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/handler"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/middleware"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, h *handler.AllOfHandler, mw *middleware.Middleware) {
	api := app.Group("/api")

	AuthRoutes(api, h, mw)
	EventRoutes(api, h, mw)
	HealthCheckRoutes(api, h)
	EventRoutes(api, h, mw)
}
