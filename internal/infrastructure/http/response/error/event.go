package error

import (
	"errors"

	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

var (
	ErrInvalidEventID          = errors.New("invalid event_id format")
	ErrInvalidUserID           = errors.New("invalid user_id format")
	ErrEventNotFound           = errors.New("event not found")
	ErrExcelGeneration         = errors.New("failed to generate excel")
	ErrInternalDB              = errors.New("internal db error")
	ErrNilUUID                 = errors.New("uuid cannot be nil")
	ErrAlreadyCommented        = errors.New("already commented")
	ErrCheckInTargetNotFound   = errors.New("check in target not found")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
)

func EventError(err error) (int, string, string) {
	switch {
	case err == nil:
		return fiber.StatusInternalServerError, response.ErrInternalError, "internal server error"

	case errors.Is(err, ErrInvalidEventID),
		errors.Is(err, ErrInvalidUserID):
		return fiber.StatusBadRequest, response.ErrBadRequest, err.Error()

	case errors.Is(err, ErrInsufficientPermissions):
		return fiber.StatusForbidden, response.ErrForbidden, err.Error()

	case errors.Is(err, ErrEventNotFound),
		errors.Is(err, gorm.ErrRecordNotFound):
		return fiber.StatusNotFound, response.ErrNotFound, "event not found"

	case errors.Is(err, ErrExcelGeneration):
		return fiber.StatusInternalServerError, response.ErrInternalError, err.Error()

	default:
		return fiber.StatusInternalServerError, response.ErrInternalError, "internal db error"
	}
}
