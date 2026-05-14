package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
)

type UserHandler interface {
	GetUserByRefId(c *fiber.Ctx) error
}

func (h *Handler) GetUserByRefId(c *fiber.Ctx) error {
	refIdStr := c.Params("ref_id")
	if refIdStr == "" {
		return response.SendError(c, 400, response.ErrBadRequest, "Empty path parameter 'ref_id'")
	}

	user, err := h.Service.User.GetUserByRefId(refIdStr, c.UserContext())
	if err != nil {
		return response.SendError(c, err.Status, err.Code, err.Message)
	}
	return response.OK(c, user)
}
