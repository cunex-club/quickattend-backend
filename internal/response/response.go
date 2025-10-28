package response

import "github.com/gofiber/fiber/v2"

type Pagination struct {
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"pageSize,omitempty"`
	Total    int64  `json:"total,omitempty"`
	NextID   string `json:"nextId,omitempty"`
}

type APIResponse struct {
	Data       any         `json:"data,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type ErrorResponse struct {
	Error APIError `json:"error"`
}

func OK(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(APIResponse{Data: data})
}

func Created(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(APIResponse{Data: data})
}

func Paginated(c *fiber.Ctx, data any, pagination Pagination) error {
	return c.Status(fiber.StatusOK).JSON(APIResponse{Data: data, Pagination: &pagination})
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
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
