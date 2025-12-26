package repository

import (
	"context"

	"github.com/cunex-club/quickattend-backend/internal/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventRepository interface {
	FindById(uuid.UUID, context.Context) (*entity.Event, error)
	DeleteById(uuid.UUID, context.Context) error
}

func (r *repository) FindById(id uuid.UUID, ctx context.Context) (*entity.Event, error) {
	return nil, nil
}

func (r *repository) DeleteById(id uuid.UUID, ctx context.Context) error {
	if id == uuid.Nil {
		return entity.ErrNilUUID
	}

	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Event{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
