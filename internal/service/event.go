package service

import (
	"context"

	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/google/uuid"
)

type EventService interface {
	EventDeleteById(EventID string, ctx context.Context) *response.APIError
}

func (s *service) EventDeleteById(EventId string, ctx context.Context) *response.APIError {
	id, parseErr := uuid.Parse(EventId)
	if parseErr != nil {
		return &response.APIError{
			Code:    response.ErrInternalError,
			Message: "failed to parse event_id to uuid",
			Status:  500,
		}
	}

	eventDeleteErr := s.repo.Event.DeleteById(id, ctx)
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
