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

