package handler

import (
	"errors"

	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/gofiber/fiber/v2"

	dtoReq "github.com/cunex-club/quickattend-backend/internal/dto/request"
	"gorm.io/gorm"
)

type EventHandler interface {
	GetOneEventHandler(*fiber.Ctx) error
	GetEvents(*fiber.Ctx) error
	CreateEvent(c *fiber.Ctx) error
	UpdateEvent(c *fiber.Ctx) error
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

func (h *Handler) GetEvents(c *fiber.Ctx) error {
	params := c.Queries()
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok {
		return response.SendError(c, 500, response.ErrInternalError, "Failed to assert user_id as a string")
	}

	res, pagination, err := h.Service.Event.GetEventsService(userIDStr, params, c.UserContext())
	if err != nil {
		return response.SendError(c, err.Status, err.Code, err.Message)
	}

	if pagination != nil {
		return response.Paginated(c, res, *pagination)
	}
	return response.OK(c, res)
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
