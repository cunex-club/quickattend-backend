package repository

import (
	"context"

	"github.com/cunex-club/quickattend-backend/internal/entity"
	"gorm.io/datatypes"
)

type EventRepository interface {
	GetOneEvent(eventId datatypes.UUID, userId datatypes.UUID, ctx context.Context) (eventWithCount *entity.GetOneEventWithTotalCount, agenda *[]entity.GetOneEventAgenda, err error)
}

func (r *repository) GetOneEvent(eventId datatypes.UUID, userId datatypes.UUID, ctx context.Context) (*entity.GetOneEventWithTotalCount, *[]entity.GetOneEventAgenda, error) {
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
	eventErr := withCtx.Table("events e").
		Select("e.name", "e.organizer", "e.description", "e.start_time",
			"e.end_time", "e.location", "e.evaluation_form", "eu.role",
			"COUNT(ep.id) AS total_registered").
		Joins("LEFT JOIN event_participants ep ON e.id = ep.event_id").
		Joins("LEFT JOIN event_users eu ON e.id = eu.event_id").
		Where("COALESCE(eu.user_id = ?, true)", userId).
		Where("e.id = ?", eventId).
		Group("e.id").
		Group("eu.role").
		Scan(&eventWithCount).Error
	if eventErr != nil {
		return nil, nil, eventErr
	}

	return &eventWithCount, &agenda, nil
}
