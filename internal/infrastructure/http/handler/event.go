package handler

import (
	"github.com/gofiber/fiber/v2"

	dtoReq "github.com/cunex-club/quickattend-backend/internal/dto/request"
	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
)

type EventHandler interface {
	EventDelete(c *fiber.Ctx) error
	EventDuplicate(c *fiber.Ctx) error
	EventCheckIn(c *fiber.Ctx) error
}

func (h *Handler) EventDelete(c *fiber.Ctx) error {
	EventID := c.Params("id")

	err := h.Service.Event.EventDeleteById(EventID, c.UserContext())
	if err != nil {
		return response.SendError(c, err.Status, err.Code, err.Message)
	}

	return response.Deleted(c, nil)
}

func (h *Handler) EventDuplicate(c *fiber.Ctx) error {
	EventID := c.Params("id")

	createdEvent, err := h.Service.Event.EventDuplicateById(EventID, c.UserContext())
	if err != nil {
		return response.SendError(c, err.Status, err.Code, err.Message)
	}

	return response.Created(c, dtoRes.DuplicateEventRes{
		DuplicatedEventId: createdEvent.ID.String(),
	})
}

func (h *Handler) EventCheckIn(c *fiber.Ctx) error {
	checkInReq := new(dtoReq.CheckInReq)
	if err := c.BodyParser(checkInReq); err != nil {
		return response.SendError(c, 400, response.ErrBadRequest, "bad request body")
	}

	err := h.Service.Event.EventCheckIn(checkInReq, c.UserContext())
	if err != nil {
		return response.SendError(c, err.Status, err.Code, err.Message)
	}

	return nil
}
