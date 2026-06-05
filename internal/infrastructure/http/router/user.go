package router

import (
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/handler"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/middleware"
	"github.com/gofiber/fiber/v2"
)

func UserRoutes(r fiber.Router, h *handler.AllOfHandler, mw *middleware.Middleware) {
	user := r.Group("/users", mw.AuthRequired())
	user.Get("/:ref_id", h.UserHandler.GetUserByRefId)
}
