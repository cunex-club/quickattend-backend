package handler

import (
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/gofiber/fiber/v2"
)

type EventHandler interface {
	GetOneEventHandler(*fiber.Ctx) error
	GetEvents(*fiber.Ctx) error
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
