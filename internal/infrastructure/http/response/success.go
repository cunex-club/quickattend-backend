package response

import "github.com/gofiber/fiber/v2"

type APIResponse struct {
	Data any   `json:"data,omitempty"`
	Meta *Meta `json:"meta,omitempty"`
}

type Meta struct {
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
	Total    int64 `json:"total"`
	HasNext  bool  `json:"hasNext"`
}

func OK(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Data: data,
	})
}

func Paginated(c *fiber.Ctx, data any, pag Pagination) error {
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Data: data,
		Meta: &Meta{Pagination: &pag},
	})
}
