package handler

import (
	"strconv"

	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/gofiber/fiber/v2"

	dtoReq "github.com/cunex-club/quickattend-backend/internal/dto/request"
)

type AuthHandler interface {
	AuthCunex(c *fiber.Ctx) error
	AuthUser(c *fiber.Ctx) error
}

func (h *Handler) AuthCunex(c *fiber.Ctx) error {

	var req dtoReq.VerifyTokenReq
	if err := c.BodyParser(&req); err != nil {
		return response.SendError(c, 400, response.ErrBadRequest, "invalid JSON body")
	}

	ctx := c.UserContext()
	res, err := h.Service.Auth.VerifyCUNEXToken(req.Token, ctx)
	if err != nil {
		return response.SendError(c, err.Status, err.Code, err.Message)
	}

	return response.OK(c, res)
}

func (h *Handler) AuthUser(c *fiber.Ctx) error {

	refIDStr, ok := c.Locals("ref_id").(string)
	if !ok {
		return response.SendError(c, 500, response.ErrInternalError, "Failed to assert ref_id as string")
	}

	refID, err := strconv.ParseUint(refIDStr, 10, 64)
	if err != nil {
		return response.SendError(c, 500, response.ErrInternalError, "Could not convert ref_id from string to uint64")
	}

	ctx := c.UserContext()
	results, getUserErr := h.Service.Auth.GetUserService(refID, ctx)
	if getUserErr != nil {
		return response.SendError(c, getUserErr.Status, getUserErr.Code, getUserErr.Message)
	}

	return response.OK(c, results)
}
