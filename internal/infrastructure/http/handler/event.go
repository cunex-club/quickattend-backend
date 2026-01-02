package handler

import (
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/gofiber/fiber/v2"
)

type EventHandler interface {
	GetEvents(*fiber.Ctx) error
}

func (h *Handler) GetEvents(c *fiber.Ctx) error {
	params := c.Queries()
	refID := c.Locals("ref_id").(uint64)

	res, pagination, err := h.Service.Event.GetEventsService(refID, params, c.UserContext())
	if err != nil {
		return response.SendError(c, err.Status, err.Code, err.Message)
	}

	if pagination != nil {
		return response.Paginated(c, res, *pagination)
	}
	return response.OK(c, res)
}
