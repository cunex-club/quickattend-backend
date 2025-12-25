package service

import (
	"context"
	"errors"

	"github.com/cunex-club/quickattend-backend/internal/entity"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventService interface {
	EventDeleteById(EventID string, ctx context.Context) *response.APIError
}

func (s *service) EventDeleteById(EventId string, ctx context.Context) *response.APIError {
	event_id, parseErr := uuid.Parse(EventId)
	if parseErr != nil {
		return &response.APIError{
			Code:    response.ErrInternalError,
			Message: "failed to parse event_id to uuid",
			Status:  500,
		}
	}

	eventDeleteErr := s.repo.Event.DeleteById(event_id, ctx)

	if errors.Is(eventDeleteErr, gorm.ErrRecordNotFound) {
		s.logger.Error().
			Err(eventDeleteErr).
			Str("event_id", EventId).
			Str("action", "delete_event").
			Msg("event not found")
		return &response.APIError{
			Code:    response.ErrNotFound,
			Message: "event not found",
			Status:  404,
		}
	}

	if errors.Is(eventDeleteErr, entity.ErrNilUUID) {
		s.logger.Error().
			Err(eventDeleteErr).
			Str("event_id", EventId).
			Str("action", "delete_event").
			Msg("attempt deleting nil uuid")
		return &response.APIError{
			Code:    response.ErrNotFound,
			Message: "nil uuid not allowed",
			Status:  400,
		}
	}

	if eventDeleteErr != nil {
		s.logger.Error().
			Err(eventDeleteErr).
			Str("event_id", EventId).
			Str("action", "delete_event").
			Msg("service failed to delete event")
		return &response.APIError{
			Code:    response.ErrInternalError,
			Message: "internal db error",
			Status:  500,
		}
	}

	return nil
}
