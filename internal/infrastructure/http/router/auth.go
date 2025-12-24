package router

import (
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/handler"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/middleware"
	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(r fiber.Router, h *handler.AllOfHandler, mw *middleware.Middleware) {
	auth := r.Group("/auth")

	public := auth.Group("")
	public.Post("/cunex", h.AuthHandler.AuthCunex)

	protected := auth.Group("", mw.AuthRequired())
	protected.Get("/user", h.AuthHandler.AuthUser)
}
