package router

import (
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/handler"
	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(r fiber.Router, h *handler.AllOfHandler) {
	health := r.Group("/auth")

	health.Get("/cunex", h.AuthHandler.AuthCunex)
}
