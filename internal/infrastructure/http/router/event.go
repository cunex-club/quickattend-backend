package router

import (
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/handler"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/middleware"
	"github.com/gofiber/fiber/v2"
)

func EventRoutes(r fiber.Router, h *handler.AllOfHandler, mw *middleware.Middleware) {
	events := r.Group("/events")

	event := r.Group("/events", mw.AuthRequired())
	event.Get("/:id", h.EventHandler.GetOneEventHandler)
	events.Post("/", h.EventHandler.CreateEvent)
	events.Put("/:id", h.EventHandler.UpdateEvent)
}
