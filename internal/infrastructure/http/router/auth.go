package router

import (
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/handler"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/middleware"
	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(r fiber.Router, h *handler.AllOfHandler, mw *middleware.Middleware) {
	// login by CU NEX callback
	r.Post("/callback", h.AuthHandler.AuthCallback)

	auth := r.Group("/auth")

	public := auth.Group("")
	// direct call from LLE
	public.Post("/cunex", h.AuthHandler.AuthCunex)

	protected := auth.Group("", mw.AuthRequired())
	protected.Get("/user", h.AuthHandler.AuthUser)
}
