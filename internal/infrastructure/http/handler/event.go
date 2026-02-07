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
	PostParticipantHandler(c *fiber.Ctx) error
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
