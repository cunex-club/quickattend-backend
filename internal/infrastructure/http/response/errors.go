package response

import "github.com/gofiber/fiber/v2"

const (
	ErrBadRequest    = "BAD_REQUEST"
	ErrUnauthorized  = "UNAUTHORIZED"
	ErrForbidden     = "FORBIDDEN"
	ErrNotFound      = "NOT_FOUND"
	ErrInternalError = "INTERNAL_SERVER_ERROR"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

type ErrorResponse struct {
	Error APIError `json:"error"`
}

func SendError(c *fiber.Ctx, status int, code string, msg string) error {
	return c.Status(status).JSON(ErrorResponse{
		Error: APIError{
			Code:    code,
			Message: msg,
			Status:  status,
		},
	})
}
