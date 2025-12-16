package handler

import (
	"strings"

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

	res, err := h.Service.Auth.VerifyCUNEXToken(req.Token)
	if err != nil {
		return response.SendError(c, err.Status, err.Code, err.Message)
	}

	return response.OK(c, res)
}

func (h *Handler) AuthUser(c *fiber.Ctx) error {

	header := c.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return response.SendError(c, 401, response.ErrUnauthorized, "missing Authorization header")
	}

	tokenStr := strings.TrimPrefix(header, "Bearer ")
	results, err := h.Service.Auth.GetUserService(tokenStr)

	if err != nil {
		return response.SendError(c, err.Status, err.Code, err.Message)
	}

	return response.OK(c, results)
}
