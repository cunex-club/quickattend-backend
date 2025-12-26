package handler

import (
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/gofiber/fiber/v2"

	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
)

type EventHandler interface {
	EventDelete(c *fiber.Ctx) error
	EventDuplicate(c *fiber.Ctx) error
}

func (h *Handler) EventDelete(c *fiber.Ctx) error {
	EventID := c.Params("id")

	err := h.Service.Event.EventDeleteById(EventID, c.Context())
	if err != nil {
		return response.SendError(c, err.Status, err.Code, err.Message)
	}

	return response.Deleted(c, nil)
}

func (h *Handler) EventDuplicate(c *fiber.Ctx) error {
	EventID := c.Params("id")

	createdEvent, err := h.Service.Event.EventDuplicateById(EventID, c.Context())
	if err != nil {
		return response.SendError(c, err.Status, err.Code, err.Message)
	}

	return response.OK(c, dtoRes.DuplicateEventRes{
		DuplicatedEventId: createdEvent.ID,
	})
}
