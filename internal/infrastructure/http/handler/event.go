package handler

import (
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/gofiber/fiber/v2"
)

type EventHandler interface {
	GetParticipantHandler(c *fiber.Ctx) error
}

func (h *Handler) GetParticipantHandler(c *fiber.Ctx) error {
	code := c.Params("qrcode")
	res, err := h.Service.Event.GetParticipantService(code, c.UserContext())
	if err != nil {
		return response.SendError(c, err.Status, err.Code, err.Message)
	}
	return response.OK(c, res)
}
