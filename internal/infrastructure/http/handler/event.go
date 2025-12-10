package handler

import (
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/gofiber/fiber/v2"
)

type EventHandler interface {
	EventDelete(c *fiber.Ctx) error
	EventDuplicate(c *fiber.Ctx) error
}

func (h *Handler) EventDelete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	println("del event ID:", id)
	if err != nil {
		return response.SendError(c, 400, response.ErrBadRequest, "invalid id")
	}
	return nil
}

func (h *Handler) EventDuplicate(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	println("duplicate event ID:", id)
	if err != nil {
		return response.SendError(c, 400, response.ErrBadRequest, "invalid id")
	}
	return nil
}
