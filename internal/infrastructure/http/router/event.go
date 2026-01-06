package router

import (
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/handler"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/middleware"
	"github.com/gofiber/fiber/v2"
)

func EventRoutes(r fiber.Router, h *handler.AllOfHandler, mw *middleware.Middleware) {
	participant := r.Group("/participant", mw.AuthRequired())
	participant.Get("/:qrcode", h.EventHandler.GetParticipantHandler)
}
