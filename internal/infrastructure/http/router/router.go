package router

import (
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/handler"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/middleware"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, h *handler.AllOfHandler, mw *middleware.Middleware) {
	api := app.Group("/api")

	// Public routes
	public := api.Group("")
	public.Get("/health-check", h.HealthCheckHandler.HealthCheck)

	protected := api.Group("", mw.AuthRequired())
	protected.Get("/me", nil) // example route
}
