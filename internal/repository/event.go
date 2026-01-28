package repository

import (
	"context"
	"fmt"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/cunex-club/quickattend-backend/internal/entity"
)

type EventRepository interface {
	GetOneEvent(eventId datatypes.UUID, userId datatypes.UUID, ctx context.Context) (eventWithCount *entity.GetOneEventWithTotalCount, agenda *[]entity.GetOneEventAgenda, err error)
	GetManagedEvents(userID datatypes.UUID, search string, ctx context.Context) (res *[]entity.GetEventsQueryResult, err error)
	GetAttendedEvents(userID datatypes.UUID, page int, pageSize int, search string, ctx context.Context) (res *[]entity.GetEventsQueryResult, total int64, hasNext bool, err error)
	GetDiscoveryEvents(userID datatypes.UUID, page int, pageSize int, search string, ctx context.Context) (res *[]entity.GetEventsQueryResult, total int64, hasNext bool, err error)
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

func (r *repository) GetManagedEvents(userID datatypes.UUID, search string, ctx context.Context) (*[]entity.GetEventsQueryResult, error) {
	tx := r.db.WithContext(ctx)

	var results []entity.GetEventsQueryResult
	if search != "" {
		searchQuery := fmt.Sprintf("%%%s%%", search)

		errGetEvents := tx.Table("events e").
			Select("e.id", "e.name", "e.organizer", "e.description", "e.start_time",
				"e.end_time", "e.location", "eu.role", "e.evaluation_form").
			Joins(`JOIN event_users eu ON eu.user_id = ? 
				AND eu.event_id = e.id`,
				userID).
			Where(`(e.name ILIKE ? OR e.organizer ILIKE ? OR e.description ILIKE ? OR e.location ILIKE ?
				OR eu.role::TEXT ILIKE ? OR e.evaluation_form ILIKE ?)`,
				searchQuery, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery).
			Order("e.id").
			Scan(&results).Error

		if errGetEvents != nil {
			return nil, errGetEvents
		}
		return &results, nil
	}

	errGetEvents := tx.Table("events e").
		Select("e.id", "e.name", "e.organizer", "e.description", "e.start_time",
			"e.end_time", "e.location", "eu.role", "e.evaluation_form").
		Joins(`JOIN event_users eu ON eu.user_id = ? 
			AND eu.event_id = e.id`,
			userID).
		Order("e.id").
		Scan(&results).Error

	if errGetEvents != nil {
		return nil, errGetEvents
	}
	return &results, nil
}

func (r *repository) GetAttendedEvents(userID datatypes.UUID, page int, pageSize int, search string, ctx context.Context) (*[]entity.GetEventsQueryResult, int64, bool, error) {
	tx := r.db.WithContext(ctx)

	var subQuery *gorm.DB
	if search != "" {
		searchQuery := fmt.Sprintf("%%%s%%", search)

		subQuery = tx.Table("events e").
			Select("e.id", "e.name", "e.organizer", "e.description", "e.start_time",
				"e.end_time", "e.location", "e.evaluation_form").
			Joins(`JOIN event_participants ep ON ep.participant_id = ? 
				AND ep.event_id = e.id
				`, userID).
			Where(`(e.name ILIKE ? OR e.organizer ILIKE ? OR e.description ILIKE ? OR e.location ILIKE ?
				OR e.evaluation_form ILIKE ?)
				`, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery)
	} else {
		subQuery = tx.Table("events e").
			Select("e.id", "e.name", "e.organizer", "e.description", "e.start_time",
				"e.end_time", "e.location", "e.evaluation_form").
			Joins(`JOIN event_participants ep ON ep.participant_id = ? 
			AND ep.event_id = e.id
			`, userID)
	}

	var count int64
	countErr := tx.Raw(`SELECT COUNT(*) AS total FROM (?) AS subQuery`, subQuery).Scan(&count).Error
	if countErr != nil {
		return nil, -1, false, countErr
	}

	var rawResult []entity.GetEventsQueryResult
	getEventsErr := tx.Raw(`SELECT subQuery.* FROM (?) AS subQuery
		ORDER BY subQuery.id
		OFFSET ?
		LIMIT ?
	`, subQuery, page*pageSize, pageSize+1).Scan(&rawResult).Error
	if getEventsErr != nil {
		return nil, -1, false, getEventsErr
	}

	if len(rawResult) <= pageSize {
		return &rawResult, count, false, nil
	}
	clipped := rawResult[:pageSize]
	return &clipped, count, true, nil
}

func (r *repository) GetDiscoveryEvents(userID datatypes.UUID, page int, pageSize int, search string, ctx context.Context) (*[]entity.GetEventsQueryResult, int64, bool, error) {
	tx := r.db.WithContext(ctx)

	var subQuery *gorm.DB
	if search != "" {
		searchQuery := fmt.Sprintf("%%%s%%", search)

		subQuery = tx.Table("events e").Select("e.id", "e.name", "e.organizer", "e.description", "e.start_time",
			"e.end_time", "e.location", "e.evaluation_form").
			Where(`NOT EXISTS (
					SELECT 1 FROM event_users eu WHERE eu.event_id = e.id
					AND eu.user_id = ?
				) AND NOT EXISTS (
					SELECT 1 FROM event_participants ep WHERE ep.event_id = e.id
					AND ep.participant_id = ?
				)`, userID, userID).
			Where(`(e.name ILIKE ? OR e.organizer ILIKE ? OR e.description ILIKE ? OR e.location ILIKE ?
				OR e.evaluation_form ILIKE ?)
				`, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery)
	} else {
		subQuery = tx.Table("events e").Select("e.id", "e.name", "e.organizer", "e.description", "e.start_time",
			"e.end_time", "e.location", "e.evaluation_form").
			Where(`NOT EXISTS (
				SELECT 1 FROM event_users eu WHERE eu.event_id = e.id
				AND eu.user_id = ?
			) AND NOT EXISTS (
				SELECT 1 FROM event_participants ep WHERE ep.event_id = e.id
				AND ep.participant_id = ?
			)`, userID, userID)
	}

	var count int64
	countErr := tx.Raw(`SELECT COUNT(*) FROM (?) AS subQuery`, subQuery).Scan(&count).Error
	if countErr != nil {
		return nil, -1, false, countErr
	}

	var rawResult []entity.GetEventsQueryResult
	getEventsErr := tx.Raw(`SELECT subQuery.* FROM (?) AS subQuery
		ORDER BY subQuery.id
		OFFSET ?
		LIMIT ?
	`, subQuery, page*pageSize, pageSize+1).Scan(&rawResult).Error
	if getEventsErr != nil {
		return nil, -1, false, getEventsErr
	}

	if len(rawResult) <= pageSize {
		return &rawResult, count, false, nil
	}
	clipped := rawResult[:pageSize]
	return &clipped, count, true, nil
}
