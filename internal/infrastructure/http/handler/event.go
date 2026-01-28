package handler

import (
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/gofiber/fiber/v2"
)

type EventHandler interface {
	GetOneEventHandler(*fiber.Ctx) error
}

func (h *Handler) GetOneEventHandler(c *fiber.Ctx) error {
	eventIdStr := c.Params("id")
	userIdStr := c.Locals("user_id").(string)

	res, err := h.Service.Event.GetOneEventService(eventIdStr, userIdStr, c.UserContext())
	if err != nil {
		return response.SendError(c, err.Status, err.Code, err.Message)
	}

	return response.OK(c, res)
}
