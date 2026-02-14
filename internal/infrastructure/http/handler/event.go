package handler

import (
	"github.com/gofiber/fiber/v2"

	dtoReq "github.com/cunex-club/quickattend-backend/internal/dto/request"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/cunex-club/quickattend-backend/internal/service"
)

type EventHandler interface {
	Delete(c *fiber.Ctx) error
	Duplicate(c *fiber.Ctx) error
	Comment(c *fiber.Ctx) error
	PostParticipantHandler(c *fiber.Ctx) error
	GetOneEventHandler(*fiber.Ctx) error
	GetEvents(*fiber.Ctx) error
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	EventID := c.Params("id")
	userIDStr := c.Locals("user_id").(string)

	err := h.Service.Event.DeleteById(EventID, userIDStr, c.UserContext())
	if err != nil {
		return response.SendError(c, err.Status, err.Code, err.Message)
	}
	return response.Deleted(c, nil)
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

func (h *Handler) Duplicate(c *fiber.Ctx) error {
	EventID := c.Params("id")
	userIDStr := c.Locals("user_id").(string)

	res, err := h.Service.Event.DuplicateById(EventID, userIDStr, c.UserContext())
	if err != nil {
		return response.SendError(c, err.Status, err.Code, err.Message)
	}

	return response.Created(c, res)
}

func (h *Handler) Comment(c *fiber.Ctx) error {
	var req dtoReq.CommentReq

	if err := c.BodyParser(&req); err != nil {
		return response.SendError(c, 400, response.ErrBadRequest, "invalid JSON body")
	}

	err := h.Service.Event.Comment(req, c.UserContext())
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

	validated, validateErr := h.Service.Event.GetEventsValidateArgs(userIDStr, params)
	if validateErr != nil {
		return response.SendError(c, validateErr.Status, validateErr.Code, validateErr.Message)
	}

	switch validated.MyEvents {
	case service.Discovery:
		args := service.GetEventsWithPaginationArgs{
			UserID:   validated.UserID,
			Page:     validated.Page,
			PageSize: validated.PageSize,
			Search:   validated.Search,
			Ctx:      c.UserContext(),
		}

		res, pag, err := h.Service.Event.GetDiscoveryEventsService(&args)
		if err != nil {
			return response.SendError(c, err.Status, err.Code, err.Message)
		}
		return response.Paginated(c, res, *pag)

	case service.PastEvents:
		args := service.GetEventsWithPaginationArgs{
			UserID:   validated.UserID,
			Page:     validated.Page,
			PageSize: validated.PageSize,
			Search:   validated.Search,
			Ctx:      c.UserContext(),
		}

		res, pag, err := h.Service.Event.GetPastEventsService(&args)
		if err != nil {
			return response.SendError(c, err.Status, err.Code, err.Message)
		}
		return response.Paginated(c, res, *pag)

	case service.MyEvents:
		res, err := h.Service.Event.GetMyEventsService(validated.UserID, validated.Search, c.UserContext())
		if err != nil {
			return response.SendError(c, err.Status, err.Code, err.Message)
		}
		return response.OK(c, res)

	default:
		// Should not happen
		return response.SendError(c, 500, response.ErrInternalError, "Unknown GetEventsMode from Event service")
	}
}
