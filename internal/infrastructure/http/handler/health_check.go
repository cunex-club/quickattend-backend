package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type HealthCheckHandler interface {
	HealthCheck(c *fiber.Ctx) error
}

func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	start := time.Now()

	status, err := h.Service.HealthCheck.CheckSystem()
	status.ResponseTime = float64(time.Since(start).Milliseconds())

	httpStatus := fiber.StatusOK
	if err != nil {
		h.Logger.Error().Err(err).Msg("Database health check failed")
		httpStatus = fiber.StatusInternalServerError
	}

	return c.Status(httpStatus).JSON(status)
}
