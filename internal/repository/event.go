package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/cunex-club/quickattend-backend/internal/entity"
)

type EventRepository interface {
	GetManagedEvents(uint64, string, context.Context) (*[]entity.GetEventsQueryResult, error)
	GetAttendedEvents(uint64, int, int, string, context.Context) (*[]entity.GetEventsQueryResult, int64, bool, error)
	GetDiscoveryEvents(uint64, int, int, string, context.Context) (*[]entity.GetEventsQueryResult, int64, bool, error)
}

// No pagination
func (r *repository) GetManagedEvents(refID uint64, search string, ctx context.Context) (*[]entity.GetEventsQueryResult, error) {
	var results []entity.GetEventsQueryResult
	var user entity.User
	tx := r.db.WithContext(ctx)

	errGetUuid := tx.Model(&user).Select("id").Where("ref_id = ?", refID).Scan(&user).Error
	if errGetUuid != nil {
		return nil, errGetUuid
	}

	if search != "" {
		searchQuery := fmt.Sprintf("%%%s%%", search)

		errGetEvents := tx.Table("events").
			Select("events.id", "events.name", "events.organizer", "events.description", "events.date", "events.start_time",
				"events.end_time", "events.location", "event_users.role", "events.evaluation_form").
			Joins(`JOIN event_users ON event_users.user_id = ? 
				AND event_users.event_id = events.id`,
				user.ID).
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
		Select("events.id", "events.name", "events.organizer", "events.description", "events.date", "events.start_time",
			"events.end_time", "events.location", "event_users.role", "events.evaluation_form").
		Joins(`JOIN event_users ON event_users.user_id = ? 
			AND event_users.event_id = events.id`,
			user.ID).
		Order("events.id").
		Scan(&results).Error

	if errGetEvents != nil {
		return nil, errGetEvents
	}
	return &results, nil
}

// returns (result, total, hasNext, error)
func (r *repository) GetAttendedEvents(refID uint64, page int, pageSize int, search string, ctx context.Context) (*[]entity.GetEventsQueryResult, int64, bool, error) {
	var rawResult []entity.GetEventsQueryResult
	tx := r.db.WithContext(ctx)

	var subQuery *gorm.DB
	if search != "" {
		searchQuery := fmt.Sprintf("%%%s%%", search)

		subQuery = tx.Table("events").
			Select("events.id", "events.name", "events.organizer", "events.description", "events.date", "events.start_time",
				"events.end_time", "events.location", "events.evaluation_form").
			Joins(`JOIN event_participants ON event_participants.participant_ref_id = ? 
				AND event_participants.event_id = events.id
				`, refID).
			Where(`(events.name ILIKE ? OR events.organizer ILIKE ? OR events.description ILIKE ? OR events.location ILIKE ?
				OR events.evaluation_form ILIKE ?)
				`, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery)
	} else {
		subQuery = tx.Table("events").
			Select("events.id", "events.name", "events.organizer", "events.description", "events.date", "events.start_time",
				"events.end_time", "events.location", "events.evaluation_form").
			Joins(`JOIN event_participants ON event_participants.participant_ref_id = ? 
			AND event_participants.event_id = events.id
			`, refID)
	}

	var total struct {
		Count int64 `gorm:"column:total"`
	}
	countErr := tx.Raw(`SELECT COUNT(*) AS total FROM (?) AS subQuery`, subQuery).Scan(&total).Error
	if countErr != nil {
		return nil, -1, false, countErr
	}

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

// returns (result, total, hasNext, error)
func (r *repository) GetDiscoveryEvents(refID uint64, page int, pageSize int, search string, ctx context.Context) (*[]entity.GetEventsQueryResult, int64, bool, error) {
	var rawResult []entity.GetEventsQueryResult
	var user entity.User
	tx := r.db.WithContext(ctx)

	errGetUuid := tx.Model(&user).Select("id").Where("ref_id = ?", refID).Scan(&user).Error
	if errGetUuid != nil {
		return nil, -1, false, errGetUuid
	}

	var subQuery *gorm.DB
	if search != "" {
		searchQuery := fmt.Sprintf("%%%s%%", search)

		subQuery = tx.Table("events").Select("events.id", "events.name", "events.organizer", "events.description", "events.date", "events.start_time",
			"events.end_time", "events.location", "events.evaluation_form").
			Where(`NOT EXISTS (
					SELECT 1 FROM event_users WHERE event_users.event_id = events.id
					AND event_users.user_id = ?
				) AND NOT EXISTS (
					SELECT 1 FROM event_participants WHERE event_participants.event_id = events.id
					AND event_participants.participant_ref_id = ?
				)`, user.ID, refID).
			Where(`(events.name ILIKE ? OR events.organizer ILIKE ? OR events.description ILIKE ? OR events.location ILIKE ?
				OR events.evaluation_form ILIKE ?)
				`, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery)
	} else {
		subQuery = tx.Table("events").Select("events.id", "events.name", "events.organizer", "events.description", "events.date", "events.start_time",
			"events.end_time", "events.location", "events.evaluation_form").
			Where(`NOT EXISTS (
				SELECT 1 FROM event_users WHERE event_users.event_id = events.id
				AND event_users.user_id = ?
			) AND NOT EXISTS (
				SELECT 1 FROM event_participants WHERE event_participants.event_id = events.id
				AND event_participants.participant_ref_id = ?
			)`, user.ID, refID)
	}

	var total struct {
		Count int64 `gorm:"column:total"`
	}
	countErr := tx.Raw(`SELECT COUNT(*) AS total FROM (?) AS subQuery`, subQuery).Scan(&total).Error
	if countErr != nil {
		return nil, -1, false, countErr
	}

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
