package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/handler"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/middleware"
)

func EventRoutes(r fiber.Router, h *handler.AllOfHandler, mw *middleware.Middleware) {
	event := r.Group("/events", mw.AuthRequired())
	event.Delete("/:id", h.EventHandler.Delete)
	event.Post("/:id/duplicate", h.EventHandler.Duplicate)
	event.Get("/:id", h.EventHandler.GetOneEventHandler)
	event.Get("/", h.EventHandler.GetEvents)

	participant := r.Group("/participant", mw.AuthRequired())
	participant.Put("/comment", h.EventHandler.Comment)
	participant.Post("/:qrcode", h.EventHandler.PostParticipantHandler)
}
