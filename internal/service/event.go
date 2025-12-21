package service

import (
	"context"

	"github.com/google/uuid"
)

type EventService interface {
	EventDeleteById(EventID string, ctx context.Context) error
}

func (s *service) EventDeleteById(EventId string, ctx context.Context) error {
	id, err := uuid.Parse(EventId)
	if err != nil {
		return err
	}

	return s.repo.Event.DeleteById(id, ctx)
}
