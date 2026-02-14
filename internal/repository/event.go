package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/cunex-club/quickattend-backend/internal/entity"
)

type EventRepository interface {
	FindById(uuid.UUID, context.Context) (*entity.Event, error)
	DeleteById(uuid.UUID, string, context.Context) error
	Create(*entity.Event, context.Context) (*entity.Event, error)
	Comment(uuid.UUID, time.Time, string, context.Context) error
	IsUserEventOwner(eventID uuid.UUID, userIDStr string, ctx context.Context) (bool, error)

	// For POST participant/:qrcode. Get user info not provided by CU NEX
	GetUserForCheckin(ctx context.Context, refID uint64) (user *entity.CheckinUserQuery, err error)
	// For POST participant/:qrcode. Get necessary event details for checking
	GetEventForCheckin(ctx context.Context, eventId datatypes.UUID, userId datatypes.UUID) (event *entity.CheckinEventQuery, err error)
	// Check if user has already checked in to the event
	CheckEventParticipation(ctx context.Context, eventId datatypes.UUID, participantID datatypes.UUID) (rowId *datatypes.UUID, err error)
	// Check if user is in whitelist / allowed org or faculty of the event
	CheckEventAccess(ctx context.Context, orgCode uint8, refID uint64, attendanceType string, eventId datatypes.UUID) (allow bool, err error)
	InsertScanRecord(ctx context.Context, record *entity.EventParticipants) (rowId *datatypes.UUID, err error)

	GetOneEvent(eventId datatypes.UUID, userId datatypes.UUID, ctx context.Context) (result *entity.GetOneEventQuery, err error)

	GetMyEvents(args *GetEventsArguments) (res *[]entity.GetEventsQueryResult, err error)
	GetPastEvents(args *GetEventsArguments) (res *[]entity.GetEventsQueryResult, total int64, hasNext bool, err error)
	GetDiscoveryEvents(args *GetEventsArguments) (res *[]entity.GetEventsQueryResult, total int64, hasNext bool, err error)
}

type GetEventsArguments struct {
	UserID   datatypes.UUID
	Page     int
	PageSize int
	Search   string
	Ctx      context.Context
}

func (r *repository) Comment(checkInRowId uuid.UUID, timeStamp time.Time, comment string, ctx context.Context) error {
	if checkInRowId == uuid.Nil {
		return entity.ErrNilUUID
	}

	result := r.db.WithContext(ctx).
		Model(&entity.EventParticipants{}).
		// Where("id = ? AND comment_timestamp IS NULL", checkInRowId).
		Where("id = ?", checkInRowId).
		Updates(map[string]any{
			"comment_timestamp": timeStamp,
			"comment":           comment,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		var exists bool
		err := r.db.WithContext(ctx).
			Model(&entity.EventParticipants{}).
			Select("count(1) > 0").
			Where("id = ?", checkInRowId).
			Find(&exists).Error

		if err != nil {
			return err
		}

		if !exists {
			return entity.ErrCheckInTargetNotFound
		}

		return entity.ErrAlreadyCommented
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

func (r *repository) DeleteById(id uuid.UUID, userIdStr string, ctx context.Context) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		var event entity.Event
		if err := tx.Select("id").First(&event, "id = ?", id).Error; err != nil {
			// gorm.ErrRecordNotFound
			return err
		}

		var isOwner int64
		err := tx.Table("event_users").
			Where("event_id = ? AND user_id = ? AND role = ?", id, userIdStr, entity.OWNER).
			Count(&isOwner).Error

		if err != nil {
			return err
		}

		if isOwner == 0 {
			return entity.ErrInsufficientPermissions
		}

		return tx.Delete(&event).Error
	})
}

func (r *repository) GetOneEvent(eventId datatypes.UUID, userId datatypes.UUID, ctx context.Context) (*entity.GetOneEventQuery, error) {
	withCtx := r.db.WithContext(ctx)

	var result entity.GetOneEventQuery
	err := withCtx.
		Model(&entity.Event{}).
		Preload("EventUser.User").
		Preload("EventAgenda", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("start_time")
		}).
		Select("events.*",
			"eu.role AS role",
			"COUNT(DISTINCT ep.participant_id) AS total_registered",
		).
		Joins("LEFT JOIN event_participants ep ON events.id = ep.event_id").
		Joins("LEFT JOIN event_users eu ON events.id = eu.event_id AND eu.user_id = ?", userId).
		Where("events.id = ?", eventId).
		Group("events.id, eu.role").
		First(&result).
		Error

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *repository) GetMyEvents(args *GetEventsArguments) (*[]entity.GetEventsQueryResult, error) {
	withCtx := r.db.WithContext(args.Ctx)

	var results []entity.GetEventsQueryResult
	if args.Search != "" {
		searchQuery := fmt.Sprintf("%%%s%%", args.Search)

		errGetEvents := withCtx.Table("events e").
			Select("e.id", "e.name", "e.organizer", "e.description", "e.start_time",
				"e.end_time", "e.location", "eu.role", "e.evaluation_form").
			Joins(`JOIN event_users eu ON eu.user_id = ? 
				AND eu.event_id = e.id`,
				args.UserID).
			Where(`NOW() <= e.end_time`).
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

	errGetEvents := withCtx.Table("events e").
		Select("e.id", "e.name", "e.organizer", "e.description", "e.start_time",
			"e.end_time", "e.location", "eu.role", "e.evaluation_form").
		Joins(`JOIN event_users eu ON eu.user_id = ? 
			AND eu.event_id = e.id`,
			args.UserID).
		Where(`NOW() <= e.end_time`).
		Order("e.id").
		Scan(&results).Error

	if errGetEvents != nil {
		return nil, errGetEvents
	}
	return &results, nil
}

func (r *repository) GetPastEvents(args *GetEventsArguments) (*[]entity.GetEventsQueryResult, int64, bool, error) {
	withCtx := r.db.WithContext(args.Ctx)

	eventUsers := withCtx.Model(&entity.EventUser{}).
		Select("event_id", "role").
		Where("user_id = ?", args.UserID)

	eventParticipants := withCtx.Model(&entity.EventParticipants{}).
		Select("event_id", "NULL AS role").
		Where("participant_id = ?", args.UserID)

	var subQuery *gorm.DB
	if args.Search != "" {
		searchQuery := fmt.Sprintf("%%%s%%", args.Search)
		subQuery = withCtx.
			Joins("(? UNION ?) AS filter", eventUsers, eventParticipants).
			Joins("JOIN events e ON e.id = filter.event_id").
			Where(`NOW() > e.end_time`).
			Where(`(e.name ILIKE ? OR e.organizer ILIKE ? OR e.description ILIKE ? OR e.location ILIKE ?
				OR e.evaluation_form ILIKE ?)
				`, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery)
	} else {
		subQuery = withCtx.
			Joins("(? UNION ?) AS filter", eventUsers, eventParticipants).
			Joins("JOIN events e ON e.id = filter.event_id").
			Where(`NOW() > e.end_time`)
	}

	var count int64
	countErr := withCtx.Raw(`SELECT COUNT(*) AS total FROM (?) AS subQuery`, subQuery).Scan(&count).Error
	if countErr != nil {
		return nil, -1, false, countErr
	}

	var rawResult []entity.GetEventsQueryResult
	getEventsErr := withCtx.Raw(`SELECT subQuery.* FROM (?) AS subQuery
		ORDER BY subQuery.id
		OFFSET ?
		LIMIT ?
	`, subQuery, args.Page*args.PageSize, args.PageSize+1).Scan(&rawResult).Error
	if getEventsErr != nil {
		return nil, -1, false, getEventsErr
	}

	if len(rawResult) <= args.PageSize {
		return &rawResult, count, false, nil
	}
	clipped := rawResult[:args.PageSize]
	return &clipped, count, true, nil
}

func (r *repository) GetDiscoveryEvents(args *GetEventsArguments) (*[]entity.GetEventsQueryResult, int64, bool, error) {
	tx := r.db.WithContext(args.Ctx)

	var subQuery *gorm.DB
	if args.Search != "" {
		searchQuery := fmt.Sprintf("%%%s%%", args.Search)

		subQuery = tx.Table("events e").Select("e.id", "e.name", "e.organizer", "e.description", "e.start_time",
			"e.end_time", "e.location", "e.evaluation_form").
			Where(`NOT EXISTS (
					SELECT 1 FROM event_users eu WHERE eu.event_id = e.id
					AND eu.user_id = ?
				) AND NOT EXISTS (
					SELECT 1 FROM event_participants ep WHERE ep.event_id = e.id
					AND ep.participant_id = ?
				)`, args.UserID, args.UserID).
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
			)`, args.UserID, args.UserID)
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
	`, subQuery, args.Page*args.PageSize, args.PageSize+1).Scan(&rawResult).Error
	if getEventsErr != nil {
		return nil, -1, false, getEventsErr
	}

	if len(rawResult) <= args.PageSize {
		return &rawResult, count, false, nil
	}
	clipped := rawResult[:args.PageSize]
	return &clipped, count, true, nil
}

func (r *repository) Create(event *entity.Event, ctx context.Context) (*entity.Event, error) {
	if err := r.db.WithContext(ctx).Create(event).Error; err != nil {
		return nil, err
	}
	return event, nil
}

func (r *repository) GetUserForCheckin(ctx context.Context, refID uint64) (*entity.CheckinUserQuery, error) {
	withCtx := r.db.WithContext(ctx)

	var user entity.CheckinUserQuery
	getUserErr := withCtx.Model(&entity.User{}).Select("title_th", "title_en").
		First(&user, &entity.User{RefID: refID}).Error
	if getUserErr != nil {
		return nil, getUserErr
	}

	return &user, nil
}

func (r *repository) GetEventForCheckin(ctx context.Context, eventId datatypes.UUID, userId datatypes.UUID) (*entity.CheckinEventQuery, error) {
	withCtx := r.db.WithContext(ctx)

	var event entity.CheckinEventQuery
	getEventErr := withCtx.Raw(`
			SELECT e.end_time, e.attendence_type, e.allow_all_to_scan, e.revealed_fields, 
				(
					SELECT (
						EXISTS
						(SELECT 1 FROM event_users WHERE event_id = ? AND user_id = ?) 
						OR EXISTS
						(SELECT 1 FROM events WHERE id = ? AND allow_all_to_scan = true)
					)
				) AS this_user_can_scan
			FROM events e
			WHERE e.id = ?
		`, eventId, userId, eventId, eventId).
		Scan(&event).Error
	if getEventErr != nil {
		return nil, getEventErr
	}

	return &event, nil
}

func (r *repository) CheckEventParticipation(ctx context.Context, eventId datatypes.UUID, participantID datatypes.UUID) (*datatypes.UUID, error) {
	withCtx := r.db.WithContext(ctx)

	var rowIdStr string
	tx := withCtx.Model(&entity.EventParticipants{}).Select("id").
		Where("event_id = ?", eventId).
		Where("participant_id = ?", participantID).
		Scan(&rowIdStr)
	if tx.Error != nil {
		return nil, tx.Error
	}

	if tx.RowsAffected == 0 {
		return nil, nil
	}

	// parse string to datatypes.UUID
	parsed, err := uuid.Parse(rowIdStr)
	if err != nil {
		return nil, err
	}
	b := make([]byte, 16)
	copy(b, parsed[:])
	rowId := datatypes.UUID(b)
	return &rowId, nil
}

func (r *repository) CheckEventAccess(ctx context.Context, orgCode uint8, refID uint64, attendanceType string, eventId datatypes.UUID) (bool, error) {
	withCtx := r.db.WithContext(ctx)
	var found bool

	switch attendanceType {
	case string(entity.FACULTIES):
		checkErr := withCtx.Raw(`SELECT EXISTS (
			SELECT 1 FROM event_allowed_faculties
			WHERE event_id = ? AND faculty_no = ?
		) AS subQuery`, eventId, orgCode).Scan(&found).Error
		if checkErr != nil {
			return false, checkErr
		}

		return found, nil

	case string(entity.WHITELIST):
		checkErr := withCtx.Raw(`SELECT EXISTS (
			SELECT 1 FROM event_whitelists
			WHERE event_id = ? AND attendee_ref_id = ?
		) AS subQuery`, eventId, refID).Scan(&found).Error
		if checkErr != nil {
			return false, checkErr
		}

		return found, nil

	default:
		// should not happen
		return false, errors.New("attendanceType is neither 'FACULTIES' nor 'WHITELISTS'")
	}
}

func (r *repository) InsertScanRecord(ctx context.Context, record *entity.EventParticipants) (*datatypes.UUID, error) {
	withCtx := r.db.WithContext(ctx)

	insertErr := withCtx.Model(&entity.EventParticipants{}).Create(record).Error
	if insertErr != nil {
		return nil, insertErr
	}

	return &record.ID, nil
}

// for duplicating events
func (r *repository) IsUserEventOwner(eventID uuid.UUID, userIDStr string, ctx context.Context) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("event_users").
		Where("event_id = ? AND user_id = ? AND role = ?", eventID, userIDStr, entity.OWNER).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
