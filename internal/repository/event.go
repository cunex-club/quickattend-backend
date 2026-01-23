package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/cunex-club/quickattend-backend/internal/entity"
)

type EventRepository interface {
	FindById(uuid.UUID, context.Context) (*entity.Event, error)
	DeleteById(uuid.UUID, context.Context) error
	Create(*entity.Event, context.Context) (*entity.Event, error)
	CheckIn(uuid.UUID, time.Time, string, context.Context) error
}

func (r *repository) CheckIn(checkInRowId uuid.UUID, timeStamp time.Time, comment string, ctx context.Context) error {
	if checkInRowId == uuid.Nil {
		return entity.ErrNilUUID
	}

	result := r.db.WithContext(ctx).Model(&entity.EventParticipants{}).
		Where("id = ? AND checkin_timestamp IS NULL", checkInRowId).
		Updates(map[string]any{"checkin_timestamp": timeStamp, "comment": comment})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return entity.ErrCheckInFailed
	}

	return nil
}

func (r *repository) FindById(id uuid.UUID, ctx context.Context) (*entity.Event, error) {
	var event entity.Event
	err := r.db.WithContext(ctx).
		Preload("EventWhitelist").
		Preload("EventAllowedFaculties").
		Preload("EventAgenda").
		First(&event, "id = ?", id).Error
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

func (r *repository) Create(event *entity.Event, ctx context.Context) (*entity.Event, error) {
	if err := r.db.WithContext(ctx).Create(event).Error; err != nil {
		return nil, err
	}
	return event, nil
}
