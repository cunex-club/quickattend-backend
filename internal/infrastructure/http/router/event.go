package router

import (
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/handler"
	"github.com/gofiber/fiber/v2"
)

func EventRoutes(r fiber.Router, h *handler.AllOfHandler) {
	event := r.Group("/event")

	event.Delete("/:id", h.EventHandler.EventDelete)
}
