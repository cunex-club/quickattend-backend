package repository

import (
	"context"

	"github.com/cunex-club/quickattend-backend/internal/entity"
	"gorm.io/datatypes"
)

type EventRepository interface {
	GetOneEvent(eventId datatypes.UUID, ctx context.Context) (eventWithCount *entity.GetOneEventWithTotalCount, agenda *[]entity.GetOneEventAgenda, err error)
}

func (r *repository) GetOneEvent(eventId datatypes.UUID, ctx context.Context) (*entity.GetOneEventWithTotalCount, *[]entity.GetOneEventAgenda, error) {
	withCtx := r.db.WithContext(ctx)

	var agenda []entity.GetOneEventAgenda
	agendaErr := withCtx.Model(&entity.EventAgenda{}).Select("activity_name", "start_time", "end_time").
		Where("event_id = ?", eventId).
		Order("start_time").
		Scan(&agenda).Error
	if agendaErr != nil {
		return nil, nil, agendaErr
	}

	var eventWithCount entity.GetOneEventWithTotalCount
	eventErr := withCtx.Model(&entity.Event{}).Select("events.name", "events.organizer", "events.description", "events.start_time",
		"events.end_time", "events.location", "events.evaluation_form", "COUNT(event_participants.id) AS total_registered").
		Joins("LEFT JOIN event_participants ON events.id = event_participants.event_id").
		Where("events.id = ?", eventId).
		Group("events.id").
		Scan(&eventWithCount).Error
	if eventErr != nil {
		return nil, nil, eventErr
	}

	return &eventWithCount, &agenda, nil
}
