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
	CreateWithRelations(*entity.Event, context.Context) (*entity.Event, error)
}

func (r *repository) FindById(id uuid.UUID, ctx context.Context) (*entity.Event, error) {
	var event entity.Event
	err := r.db.WithContext(ctx).First(&event, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return &event, nil
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

func (r *repository) CreateWithRelations(event *entity.Event, ctx context.Context) (*entity.Event, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Create(event).Error
	})

	if err != nil {
		return nil, err
	}

	return event, nil
}
