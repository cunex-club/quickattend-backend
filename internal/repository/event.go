package repository

import (
	"context"
	"fmt"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/cunex-club/quickattend-backend/internal/entity"
)

type EventRepository interface {
	GetManagedEvents(userID datatypes.UUID, search string, ctx context.Context) (res *[]entity.GetEventsQueryResult, err error)
	GetAttendedEvents(userID datatypes.UUID, page int, pageSize int, search string, ctx context.Context) (res *[]entity.GetEventsQueryResult, total int64, hasNext bool, err error)
	GetDiscoveryEvents(userID datatypes.UUID, page int, pageSize int, search string, ctx context.Context) (res *[]entity.GetEventsQueryResult, total int64, hasNext bool, err error)
}

func (r *repository) GetManagedEvents(userID datatypes.UUID, search string, ctx context.Context) (*[]entity.GetEventsQueryResult, error) {
	tx := r.db.WithContext(ctx)

	var results []entity.GetEventsQueryResult
	if search != "" {
		searchQuery := fmt.Sprintf("%%%s%%", search)

		errGetEvents := tx.Table("events").
			Select("events.id", "events.name", "events.organizer", "events.description", "events.start_time",
				"events.end_time", "events.location", "event_users.role", "events.evaluation_form").
			Joins(`JOIN event_users ON event_users.user_id = ? 
				AND event_users.event_id = events.id`,
				userID).
			Where(`(events.name ILIKE ? OR events.organizer ILIKE ? OR events.description ILIKE ? OR events.location ILIKE ?
				OR event_users.role::TEXT ILIKE ? OR events.evaluation_form ILIKE ?)`,
				searchQuery, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery).
			Order("events.id").
			Scan(&results).Error

		if errGetEvents != nil {
			return nil, errGetEvents
		}
		return &results, nil
	}

	errGetEvents := tx.Table("events").
		Select("events.id", "events.name", "events.organizer", "events.description", "events.start_time",
			"events.end_time", "events.location", "event_users.role", "events.evaluation_form").
		Joins(`JOIN event_users ON event_users.user_id = ? 
			AND event_users.event_id = events.id`,
			userID).
		Order("events.id").
		Scan(&results).Error

	if errGetEvents != nil {
		return nil, errGetEvents
	}
	return &results, nil
}

func (r *repository) GetAttendedEvents(userID datatypes.UUID, page int, pageSize int, search string, ctx context.Context) (*[]entity.GetEventsQueryResult, int64, bool, error) {
	tx := r.db.WithContext(ctx)

	var user entity.User
	errGetRefId := tx.Model(&user).Select("ref_id").Where("id = ?", userID).Scan(&user).Error
	if errGetRefId != nil {
		return nil, -1, false, errGetRefId
	}

	var subQuery *gorm.DB
	if search != "" {
		searchQuery := fmt.Sprintf("%%%s%%", search)

		subQuery = tx.Table("events").
			Select("events.id", "events.name", "events.organizer", "events.description", "events.start_time",
				"events.end_time", "events.location", "events.evaluation_form").
			Joins(`JOIN event_participants ON event_participants.participant_ref_id = ? 
				AND event_participants.event_id = events.id
				`, user.RefID).
			Where(`(events.name ILIKE ? OR events.organizer ILIKE ? OR events.description ILIKE ? OR events.location ILIKE ?
				OR events.evaluation_form ILIKE ?)
				`, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery)
	} else {
		subQuery = tx.Table("events").
			Select("events.id", "events.name", "events.organizer", "events.description", "events.start_time",
				"events.end_time", "events.location", "events.evaluation_form").
			Joins(`JOIN event_participants ON event_participants.participant_ref_id = ? 
			AND event_participants.event_id = events.id
			`, user.RefID)
	}

	var total struct {
		Count int64 `gorm:"column:total"`
	}
	countErr := tx.Raw(`SELECT COUNT(*) AS total FROM (?) AS subQuery`, subQuery).Scan(&total).Error
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
		return &rawResult, total.Count, false, nil
	}
	clipped := rawResult[:pageSize]
	return &clipped, total.Count, true, nil
}

func (r *repository) GetDiscoveryEvents(userID datatypes.UUID, page int, pageSize int, search string, ctx context.Context) (*[]entity.GetEventsQueryResult, int64, bool, error) {
	tx := r.db.WithContext(ctx)

	var user entity.User
	errGetRefId := tx.Model(&user).Select("ref_id").Where("id = ?", userID).Scan(&user).Error
	if errGetRefId != nil {
		return nil, -1, false, errGetRefId
	}

	var subQuery *gorm.DB
	if search != "" {
		searchQuery := fmt.Sprintf("%%%s%%", search)

		subQuery = tx.Table("events").Select("events.id", "events.name", "events.organizer", "events.description", "events.start_time",
			"events.end_time", "events.location", "events.evaluation_form").
			Where(`NOT EXISTS (
					SELECT 1 FROM event_users WHERE event_users.event_id = events.id
					AND event_users.user_id = ?
				) AND NOT EXISTS (
					SELECT 1 FROM event_participants WHERE event_participants.event_id = events.id
					AND event_participants.participant_ref_id = ?
				)`, userID, user.RefID).
			Where(`(events.name ILIKE ? OR events.organizer ILIKE ? OR events.description ILIKE ? OR events.location ILIKE ?
				OR events.evaluation_form ILIKE ?)
				`, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery)
	} else {
		subQuery = tx.Table("events").Select("events.id", "events.name", "events.organizer", "events.description", "events.start_time",
			"events.end_time", "events.location", "events.evaluation_form").
			Where(`NOT EXISTS (
				SELECT 1 FROM event_users WHERE event_users.event_id = events.id
				AND event_users.user_id = ?
			) AND NOT EXISTS (
				SELECT 1 FROM event_participants WHERE event_participants.event_id = events.id
				AND event_participants.participant_ref_id = ?
			)`, userID, user.RefID)
	}

	var total struct {
		Count int64 `gorm:"column:total"`
	}
	countErr := tx.Raw(`SELECT COUNT(*) AS total FROM (?) AS subQuery`, subQuery).Scan(&total).Error
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
		return &rawResult, total.Count, false, nil
	}
	clipped := rawResult[:pageSize]
	return &clipped, total.Count, true, nil
}
