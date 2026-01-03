package handler

import (
	"strconv"

	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/gofiber/fiber/v2"
)

type EventHandler interface {
	GetEvents(*fiber.Ctx) error
}

func (h *Handler) GetEvents(c *fiber.Ctx) error {
	params := c.Queries()
	refIDFloat, errStrToFloat := strconv.ParseFloat(c.Locals("ref_id").(string), 64)
	if errStrToFloat != nil {
		return response.SendError(c, 400, response.ErrBadRequest, "ref_id should be able to convert to float64")
	}
	refID := uint64(refIDFloat)

	res, pagination, err := h.Service.Event.GetEventsService(refID, params, c.UserContext())
	if err != nil {
		return response.SendError(c, err.Status, err.Code, err.Message)
	}

	if pagination != nil {
		return response.Paginated(c, res, *pagination)
	}
	return response.OK(c, res)
}
