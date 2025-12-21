package handler

import (
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/gofiber/fiber/v2"
)

type EventHandler interface {
	EventDelete(c *fiber.Ctx) error
}

func (h *Handler) EventDelete(c *fiber.Ctx) error {
	EventID := c.Params("event_id")

	err := h.Service.Event.EventDeleteById(EventID, c.Context())	
	if err != nil {
		return response.SendError(c, 500, "failed to delete event", err.Error())
	}

	return response.Deleted(c, nil)
}

