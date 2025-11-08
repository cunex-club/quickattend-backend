package handler

import (
	"time"

	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler interface {
	AuthCunex(c *fiber.Ctx) error
}

func (h *Handler) AuthCunex(c *fiber.Ctx) error {
	start := time.Now()

	status, err := h.Service.HealthCheck.CheckSystem()
	status.ResponseTime = float64(time.Since(start).Milliseconds())
	status.ResponseTime = 999999
	print("Hello AuthCunex in AuthHandler!\n")

	if err != nil {
		h.Logger.Error().Err(err).Msg("Database health check failed")
		return response.SendError(
			c,
			fiber.StatusInternalServerError,
			response.ErrInternalError,
			"Database connection failed",
		)
	}
	return response.OK(c, status)
}
