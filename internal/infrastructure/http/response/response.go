package response

import "github.com/gofiber/fiber/v2"

const (
	ErrBadRequest    = "BAD_REQUEST"
	ErrUnauthorized  = "UNAUTHORIZED"
	ErrForbidden     = "FORBIDDEN"
	ErrNotFound      = "NOT_FOUND"
	ErrValidation    = "VALIDATION_ERROR"
	ErrInternalError = "INTERNAL_SERVER_ERROR"
	ErrConflict = "CONFLICT"
)

type APIResponse struct {
	Data  any       `json:"data"`
	Error *APIError `json:"error"`
	Meta  *Meta     `json:"meta"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

type Meta struct {
	Pagination *Pagination `json:"pagination"`
}

type Pagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
	Total    int64 `json:"total"`
	HasNext  bool  `json:"hasNext"`
}

// --- Success Helpers -----------

func OK(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Data:  data,
		Error: nil,
		Meta:  nil,
	})
}

func Created(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(APIResponse{
		Data:  data,
		Error: nil,
		Meta:  nil,
	})
}

func Deleted(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusNoContent).JSON(APIResponse{
		Data:  data,
		Error: nil,
		Meta:  nil,
	})
}

func Paginated(c *fiber.Ctx, data any, pag Pagination) error {
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Data:  data,
		Error: nil,
		Meta:  &Meta{Pagination: &pag},
	})
}

// --- Error Helper ----------

func SendError(c *fiber.Ctx, status int, code string, msg string) error {
	return c.Status(status).JSON(APIResponse{
		Data: nil,
		Meta: nil,
		Error: &APIError{
			Code:    code,
			Message: msg,
			Status:  status,
		},
	})
}
