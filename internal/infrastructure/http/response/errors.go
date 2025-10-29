package response

import "github.com/gofiber/fiber/v2"

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type ErrorResponse struct {
	Error APIError `json:"error"`
}

func BadRequest(c *fiber.Ctx, msg string, details any) error {
	return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
		Error: APIError{Code: "BAD_REQUEST", Message: msg, Details: details},
	})
}

func Unauthorized(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
		Error: APIError{Code: "UNAUTHORIZED", Message: msg},
	})
}
