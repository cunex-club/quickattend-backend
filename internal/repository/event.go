package repository

import (
	"context"
	"github.com/cunex-club/quickattend-backend/internal/entity"
	"github.com/google/uuid"
)

type EventRepository interface {
	DeleteById(uuid.UUID, context.Context) error
}

func (r *repository) DeleteById(id uuid.UUID, ctx context.Context) error {
	return r.db.WithContext(ctx).Delete(&entity.Event{}, id).Error
}
