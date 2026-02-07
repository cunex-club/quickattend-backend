package router

import (
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/handler"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/middleware"
	"github.com/gofiber/fiber/v2"
)

func EventRoutes(r fiber.Router, h *handler.AllOfHandler, mw *middleware.Middleware) {
	event := r.Group("/events", mw.AuthRequired())
	event.Delete("/:id", h.EventHandler.Delete)
	event.Post("/:id/duplicate", h.EventHandler.Duplicate)

	participant := r.Group("/participant", mw.AuthRequired())
	participant.Put("/comment", h.EventHandler.Comment)
	participant.Post("/:qrcode", h.EventHandler.PostParticipantHandler)
	event.Get("/:id", h.EventHandler.GetOneEventHandler)
}
