package handler

import (
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/gofiber/fiber/v2"

	dtoReq "github.com/cunex-club/quickattend-backend/internal/dto/request"
)

type AuthHandler interface {
	AuthCunex(c *fiber.Ctx) error
	AuthUser(c *fiber.Ctx) error
	AuthCallback(c *fiber.Ctx) error
}

func (h *Handler) AuthCallback(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		// might redirect to a frontend error page, send error json for now
		return response.SendError(c, 400, response.ErrBadRequest, "missing token in callback URL")
	}

	ctx := c.UserContext()
	res, err := h.Service.Auth.VerifyCUNEXToken(token, ctx)
	if err != nil {
		// might redirect to a frontend error page, send error json for now
		return response.SendError(c, err.Status, err.Code, err.Message)
	}

	c.Cookie(&fiber.Cookie{
		Name:     "jwt",
		Value:    res.AccessToken,
		Path:     "/",
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Lax",
	})

	frontendHomeURL := "https://quickattend.cunex.club/"
	return c.Redirect(frontendHomeURL, fiber.StatusFound)
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
	userIDStr := c.Locals("user_id").(string)

	results, getUserErr := h.Service.Auth.GetUserService(userIDStr, c.UserContext())
	if getUserErr != nil {
		return response.SendError(c, getUserErr.Status, getUserErr.Code, getUserErr.Message)
	}

	return response.OK(c, results)
}
