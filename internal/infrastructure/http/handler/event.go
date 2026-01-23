package handler

import (
	"github.com/gofiber/fiber/v2"

	dtoReq "github.com/cunex-club/quickattend-backend/internal/dto/request"
	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
)

type EventHandler interface {
	Delete(c *fiber.Ctx) error
	Duplicate(c *fiber.Ctx) error
	CheckIn(c *fiber.Ctx) error
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	EventID := c.Params("id")

	err := h.Service.Event.DeleteById(EventID, c.UserContext())
	if err != nil {
		return response.SendError(c, err.Status, err.Code, err.Message)
	}

	return response.Deleted(c, nil)
}

func (h *Handler) Duplicate(c *fiber.Ctx) error {
	EventID := c.Params("id")

	createdEvent, err := h.Service.Event.DuplicateById(EventID, c.UserContext())
	if err != nil {
		return response.SendError(c, err.Status, err.Code, err.Message)
	}

	return response.Created(c, dtoRes.DuplicateEventRes{
		DuplicatedEventId: createdEvent.ID.String(),
	})
}

func (h *Handler) CheckIn(c *fiber.Ctx) error {
	var req dtoReq.CheckInReq

	if err := c.BodyParser(&req); err != nil {
		return response.SendError(c, 400, response.ErrBadRequest, "invalid JSON body")
	}

	err := h.Service.Event.CheckIn(req, c.UserContext())
	if err != nil {
		return response.SendError(c, err.Status, err.Code, err.Message)
	}

	return nil
}
