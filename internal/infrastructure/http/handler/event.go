package handler

import (
	"errors"

	dtoReq "github.com/cunex-club/quickattend-backend/internal/dto/request"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type EventHandler interface {
	CreateEvent(c *fiber.Ctx) error
	UpdateEvent(c *fiber.Ctx) error
}

func (h *Handler) CreateEvent(c *fiber.Ctx) error {
	var req dtoReq.CreateEventReq
	if err := c.BodyParser(&req); err != nil {
		return response.SendError(c, fiber.StatusBadRequest, response.ErrBadRequest, "invalid json body")
	}

	res, err := h.Service.Event.CreateEvent(c.Context(), req)
	if err != nil {
		return response.SendError(c, fiber.StatusBadRequest, response.ErrValidation, err.Error())
	}

	return response.Created(c, res)
}

func (h *Handler) UpdateEvent(c *fiber.Ctx) error {
	id := c.Params("id")

	var req dtoReq.UpdateEventReq
	if err := c.BodyParser(&req); err != nil {
		return response.SendError(c, fiber.StatusBadRequest, response.ErrBadRequest, "invalid json body")
	}

	res, err := h.Service.Event.UpdateEvent(c.Context(), id, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.SendError(c, fiber.StatusNotFound, response.ErrNotFound, "not found")
		}
		return response.SendError(c, fiber.StatusBadRequest, response.ErrValidation, err.Error())
	}
	return response.OK(c, res)
}
