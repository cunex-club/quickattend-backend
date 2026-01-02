package repository

import (
	"context"
	"fmt"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
)

type EventRepository interface {
	GetManagedEvents(uint64, string, context.Context) (*[]dtoRes.GetEventsIndividualEvent, error)
	GetAttendedEvents(uint64, int, int, string, context.Context) (*[]dtoRes.GetEventsIndividualEvent, int64, bool, error)
	GetDiscoveryEvents(uint64, int, int, string, context.Context) (*[]dtoRes.GetEventsIndividualEvent, int64, bool, error)
}

// managed (กิจกรรมของฉัน) section has no pagination
func (r *repository) GetManagedEvents(refID uint64, search string, ctx context.Context) (*[]dtoRes.GetEventsIndividualEvent, error) {
	var results []dtoRes.GetEventsIndividualEvent
	var userUuid datatypes.UUID
	tx := r.db.WithContext(ctx)

	errGetUuid := tx.Table("users").Select("id").Where("ref_id = ?", refID).First(&userUuid).Error
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
				userUuid).
			Where(`events.name ILIKE ? OR events.organizer ILIKE ? OR events.description ILIKE ? OR events.location ILIKE ?
				OR event_users.role ILIKE ? OR events.evaluation_form ILIKE ?`,
				searchQuery, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery).
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
			userUuid).
		Scan(&results).Error

	if errGetEvents != nil {
		return nil, errGetEvents
	}
	return &results, nil
}

// returns (result, total, hasNext, error)
func (r *repository) GetAttendedEvents(refID uint64, page int, pageSize int, search string, ctx context.Context) (*[]dtoRes.GetEventsIndividualEvent, int64, bool, error) {
	var results []dtoRes.GetEventsIndividualEvent
	var total int64
	tx := r.db.WithContext(ctx)

	if search != "" {
		searchQuery := fmt.Sprintf("%%%s%%", search)

		errGetEvents := tx.Table("events").
			Select("events.id", "events.name", "events.organizer", "events.description", "events.date", "events.start_time",
				"events.end_time", "events.location", "events.evaluation_form").
			Joins(`JOIN event_participants ON event_participants.participant_ref_id = ? 
				AND event_participants.event_id = events.id
				`, refID).
			Where(`events.name ILIKE ? OR events.organizer ILIKE ? OR events.description ILIKE ? OR events.location ILIKE ?
				OR events.evaluation_form ILIKE ?
				`, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery).
			Count(&total).
			Offset((page - 1) * pageSize).
			Limit(pageSize + 1).
			Scan(&results).Error

		if errGetEvents != nil {
			return nil, -1, false, errGetEvents
		}

		if len(results) <= pageSize {
			return &results, total, false, nil
		}
		clipped := results[:pageSize]
		return &clipped, total, true, nil
	}

	errGetEvents := tx.Table("events").
		Select("events.id", "events.name", "events.organizer", "events.description", "events.date", "events.start_time",
			"events.end_time", "events.location", "events.evaluation_form").
		Joins(`JOIN event_participants ON event_participants.participant_ref_id = ? 
			AND event_participants.event_id = events.id
			`, refID).
		Count(&total).
		Offset((page - 1) * pageSize).
		Limit(pageSize + 1).
		Scan(&results).Error

	if errGetEvents != nil {
		return nil, -1, false, errGetEvents
	}

	if len(results) <= pageSize {
		return &results, total, false, nil
	}
	clipped := results[:pageSize]
	return &clipped, total, true, nil
}

// returns (result, total, hasNext, error)
func (r *repository) GetDiscoveryEvents(refID uint64, page int, pageSize int, search string, ctx context.Context) (*[]dtoRes.GetEventsIndividualEvent, int64, bool, error) {
	type Results struct {
		rows  []dtoRes.GetEventsIndividualEvent `gorm:"column:rows"`
		total int64                             `gorm:"column:total"`
	}
	var res Results
	var userUuid datatypes.UUID
	tx := r.db.WithContext(ctx)

	errGetUuid := tx.Table("users").Select("id").Where("ref_id = ?", refID).First(&userUuid).Error
	if errGetUuid != nil {
		return nil, -1, false, errGetUuid
	}

	var subQuery *gorm.DB
	if search != "" {
		searchQuery := fmt.Sprintf("%%%s%%", search)

		subQuery = tx.Table("events").Select("events.id", "events.name", "events.organizer", "events.description", "events.date", "events.start_time",
			"events.end_time", "events.location", "events.evaluation_form").
			Distinct("events.id").
			Where(`NOT EXISTS (
					SELECT 1 FROM event_users WHERE event_users.event_id = events.id
					AND event_users.user_id = ?
				) AND NOT EXISTS (
					SELECT 1 FROM event_participants WHERE event_participants.event_id = events.id
					AND event_participants.participant_ref_id = ?
				)`, userUuid, refID).
			Where(`(events.name ILIKE ? OR events.organizer ILIKE ? OR events.description ILIKE ? OR events.location ILIKE ?
				OR events.evaluation_form ILIKE ?)
				`, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery)

		// subQuery = tx.Raw(`
		// 	SELECT (id, name, organizer, description, date, start_time, end_time, location, evaluation_form) FROM events
		// 	WHERE NOT EXISTS (
		// 		SELECT 1 FROM event_users WHERE event_users.event_id = events.id
		// 		AND event_users.user_id = ?
		// 	) AND NOT EXISTS (
		// 		SELECT 1 FROM event_participants WHERE event_participants.event_id = events.id
		// 		AND event_participants.participant_ref_id = ?
		// 	) AND (
		// 	 	events.name ILIKE ? OR events.organizer ILIKE ? OR events.description ILIKE ? OR events.location ILIKE ?
		// 		OR events.evaluation_form ILIKE ?
		// 	)
		// `, userUuid, refID, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery)
	} else {
		subQuery = tx.Table("events").Select("events.id", "events.name", "events.organizer", "events.description", "events.date", "events.start_time",
			"events.end_time", "events.location", "events.evaluation_form").
			Distinct("events.id").
			Where(`NOT EXISTS (
				SELECT 1 FROM event_users WHERE event_users.event_id = events.id
				AND event_users.user_id = ?
			) AND NOT EXISTS (
				SELECT 1 FROM event_participants WHERE event_participants.event_id = events.id
				AND event_participants.participant_ref_id = ?
			)`, userUuid, refID)

		// subQuery = tx.Raw(`
		// 	SELECT (id, name, organizer, description, date, start_time, end_time, location, evaluation_form) FROM events
		// 	WHERE NOT EXISTS (
		// 		SELECT 1 FROM event_users WHERE event_users.event_id = events.id
		// 		AND event_users.user_id = ?
		// 	) AND NOT EXISTS (
		// 		SELECT 1 FROM event_participants WHERE event_participants.event_id = events.id
		// 		AND event_participants.participant_ref_id = ?
		// 	)
		// `, userUuid, refID)
	}

	getEventsErr := tx.Select(`(* AS rows, COUNT(*) OVER() AS total) FROM (?) AS subQuery`, subQuery).
		Order("events.id").
		Offset((page - 1) * pageSize).
		Limit(pageSize + 1).
		Scan(&res).Error
	if getEventsErr != nil {
		return nil, -1, false, getEventsErr
	}

	if len(res.rows) <= pageSize {
		return &(res.rows), res.total, false, nil
	}
	clipped := res.rows[:pageSize]
	return &clipped, res.total, true, nil

	// errCountTotal := baseQuery.Count(&total).Error
	// if errCountTotal != nil {
	// 	return nil, -1, false, errCountTotal
	// }

	// errPagination := baseQuery.
	// 	Order("events.id").
	// 	Offset((page - 1) * pageSize).
	// 	Limit(pageSize + 1).
	// 	Scan(&results).Error
	// if errPagination != nil {
	// 	return nil, -1, false, errPagination
	// }

	// if len(results) <= pageSize {
	// 	return &results, total, false, nil
	// }
	// clipped := results[:pageSize]
	// return &clipped, total, true, nil
}
