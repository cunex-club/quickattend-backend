package handler

import (
	dtoReq "github.com/cunex-club/quickattend-backend/internal/dto/request"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/gofiber/fiber/v2"
)

type EventHandler interface {
	PostParticipantHandler(c *fiber.Ctx) error
}

func (h *Handler) PostParticipantHandler(c *fiber.Ctx) error {
	code := c.Params("qrcode")
	userId := c.Locals("user_id").(string)

	var reqBody dtoReq.PostParticipantReqBody
	parseBodyErr := c.BodyParser(&reqBody)
	if parseBodyErr != nil {
		return response.SendError(c, 400, response.ErrBadRequest, "Invalid request body")
	}

	res, serviceErr := h.Service.Event.PostParticipantService(code, reqBody.EventId, userId, reqBody.ScannedLocationX, reqBody.ScannedLocationY, c.UserContext())
	if serviceErr != nil {
		return response.SendError(c, serviceErr.Status, serviceErr.Code, serviceErr.Message)
	}
	return response.OK(c, res)
}
