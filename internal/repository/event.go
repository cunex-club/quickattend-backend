package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
	"github.com/cunex-club/quickattend-backend/internal/entity"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EventRepository interface {
	FindById(uuid.UUID, context.Context) (*entity.Event, error)
	DeleteById(uuid.UUID, string, context.Context) error
	Create(*entity.Event, context.Context) (*entity.Event, error)
	Comment(uuid.UUID, time.Time, string, context.Context) error
	// GetEventIDForCheckInRow looks up which event a check-in row belongs to,
	// so callers can authorize a comment write before touching the row.
	GetEventIDForCheckInRow(checkInRowId uuid.UUID, ctx context.Context) (uuid.UUID, error)
	IsUserEventOwner(eventID uuid.UUID, userIDStr string, ctx context.Context) (bool, error)
	GetUserRoleInEvent(eventID uuid.UUID, userID uuid.UUID, ctx context.Context) (*string, error)
	GetEventOwnerRefID(eventID uuid.UUID, ctx context.Context) (uint64, error)
	GetEventEndTime(eventID uuid.UUID, ctx context.Context) (time.Time, error)

	// For POST participant/:qrcode. Get user info not provided by CU NEX
	GetUserForCheckin(ctx context.Context, refID uint64) (user *entity.CheckinUserQuery, err error)
	// For POST participant/:qrcode. Get necessary event details for checking
	GetEventForCheckin(ctx context.Context, eventId datatypes.UUID, userId datatypes.UUID) (event *entity.CheckinEventQuery, err error)
	// Check if user has already checked in to the event
	CheckEventParticipation(ctx context.Context, eventId datatypes.UUID, participantID datatypes.UUID) (rowId *datatypes.UUID, err error)
	// Check if user is in whitelist / allowed org or faculty of the event
	CheckEventAccess(ctx context.Context, orgCode *uint8, refID uint64, attendanceType string, eventId datatypes.UUID) (allow bool, err error)
	InsertScanRecord(ctx context.Context, record *entity.EventParticipants) (rowId *datatypes.UUID, err error)

	GetOneEvent(eventId datatypes.UUID, userId datatypes.UUID, ctx context.Context) (result *entity.GetOneEventQuery, err error)

	GetMyEvents(args *GetEventsArguments) (res *[]entity.GetEventsQueryResult, err error)
	GetPastEvents(args *GetEventsArguments) (res *[]entity.GetEventsQueryResult, total int64, hasNext bool, err error)
	GetDiscoveryEvents(args *GetEventsArguments) (res *[]entity.GetDiscoveryEvents, total int64, hasNext bool, err error)

	CreateEvent(ctx context.Context, payload entity.CreateEventPayload) (*dtoRes.CreateEventRes, error)
	UpdateEvent(ctx context.Context, id string, payload entity.CreateEventPayload) (*dtoRes.UpdateEventRes, error)
	GetScannedParticipantsForExport(ctx context.Context, eventID datatypes.UUID) (*entity.EventParticipantExportData, error)
}

func (r *repository) GetScannedParticipantsForExport(ctx context.Context, eventID datatypes.UUID) (*entity.EventParticipantExportData, error) {
	db := r.db.WithContext(ctx)

	var event entity.Event
	if err := db.Select("name", "start_time").First(&event, "id = ?", eventID).Error; err != nil {
		return nil, err
	}

	rows := make([]entity.EventParticipantExportRow, 0)
	if err := db.Table("event_participants AS ep").
		Select(`ep.scanned_timestamp,
			u.user_type, u.ref_id,
			u.firstname_th, u.surname_th,
			u.firstname_en, u.surname_en,
			ep.organization,
			scanner.ref_id AS scanner_ref_id,
			ep.comment`).
		Joins("JOIN users u ON u.id = ep.participant_id").
		Joins("LEFT JOIN users scanner ON scanner.id = ep.scanner_id").
		Where("ep.event_id = ?", eventID).
		Order("ep.scanned_timestamp ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	return &entity.EventParticipantExportData{
		EventName: event.Name,
		StartTime: event.StartTime,
		Rows:      rows,
	}, nil
}

type GetEventsArguments struct {
	UserID   datatypes.UUID
	Page     int
	PageSize int
	Search   string
	Roles []string
	// RoleFilterProvided distinguishes "no role param sent" (no filtering)
	// from "role param sent but empty" (filter down to role IS NULL only) —
	// len(Roles) alone can't tell those apart since both are empty slices.
	RoleFilterProvided bool
	Date               *time.Time
	Sort               string
	Ctx                context.Context
}

func (r *repository) GetEventIDForCheckInRow(checkInRowId uuid.UUID, ctx context.Context) (uuid.UUID, error) {
	// Scan into a string rather than straight into a UUID type: datatypes.UUID
	// is a [16]byte array, and GORM treats an array destination as a
	// multi-row scan (see CheckEventParticipation, which does the same).
	var eventIDStr string
	err := r.db.WithContext(ctx).
		Model(&entity.EventParticipants{}).
		Select("event_id").
		Where("id = ?", checkInRowId).
		Take(&eventIDStr).Error
	if err != nil {
		return uuid.Nil, err
	}

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return uuid.Nil, err
	}
	return eventID, nil
}

func (r *repository) Comment(checkInRowId uuid.UUID, timeStamp time.Time, comment string, ctx context.Context) error {
	if checkInRowId == uuid.Nil {
		return entity.ErrNilUUID
	}

	result := r.db.WithContext(ctx).
		Model(&entity.EventParticipants{}).
		Where("id = ? AND comment_timestamp IS NULL", checkInRowId).
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

	// Only preload what's necessary for event duplication
	err := r.db.WithContext(ctx).
		Model(&entity.Event{}).
		Preload("EventUser.User").
		Preload("EventUserPending").
		Preload("EventAllowedFaculties").
		Preload("EventWhitelist.User").
		Preload("EventWhitelistPending").
		Where("id = ?", id).
		First(&event).
		Error
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

	var event entity.Event
	errEvent := withCtx.
		Model(&entity.Event{}).
		Preload("EventUser.User", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("ref_id")
		}).
		Preload("EventUserPending", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("user_ref_id")
		}).
		Preload("EventAgenda", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("start_time")
		}).
		Preload("EventAllowedFaculties", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("faculty_no")
		}).
		Preload("EventWhitelist.User", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("ref_id")
		}).
		Preload("EventWhitelistPending", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("attendee_ref_id")
		}).
		Where("id = ?", eventId).
		First(&event).
		Error
	if errEvent != nil {
		return nil, errEvent
	}

	var registerCount uint16
	errCount := withCtx.Model(&entity.EventParticipants{}).
		Select("COUNT(DISTINCT participant_id)").
		Where("event_id = ?", eventId).
		Scan(&registerCount).
		Error
	if errCount != nil {
		return nil, errCount
	}

	var role *string
	errRole := withCtx.Model(&entity.EventUser{}).
		Select("role").
		Where("user_id = ? AND event_id = ?", userId, eventId).
		Scan(&role).
		Error
	if errRole != nil {
		return nil, errRole
	}

	result := entity.GetOneEventQuery{
		Event:           event,
		Role:            role,
		TotalRegistered: registerCount,
	}
	return &result, nil
}

func (r *repository) GetMyEvents(args *GetEventsArguments) (*[]entity.GetEventsQueryResult, error) {
	withCtx := r.db.WithContext(args.Ctx)

	query := withCtx.Table("events e").
		Select("e.id", "e.name", "e.organizer", "e.description", "e.start_time",
			"e.end_time", "e.location", "eu.role", "e.evaluation_form").
		Joins(`JOIN event_users eu ON eu.user_id = ? 
			AND eu.event_id = e.id`,
			args.UserID).
		Where(`NOW() <= e.end_time`)

	if args.Search != "" {
		searchQuery := fmt.Sprintf("%%%s%%", args.Search)
		query = query.
			Where(`(e.name ILIKE ? OR e.organizer ILIKE ? OR e.description ILIKE ? OR e.location ILIKE ?
				OR eu.role::TEXT ILIKE ? OR e.evaluation_form ILIKE ?)`,
				searchQuery, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery)
	}

	var results []entity.GetEventsQueryResult
	errGetEvents := query.
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
		Where("participant_id = ?", args.UserID).
		Where(`NOT EXISTS (
			SELECT 1 FROM event_users eu
			WHERE eu.event_id = event_participants.event_id
			AND eu.user_id = event_participants.participant_id
		)`)

	subQuery := withCtx.
		Joins("(? UNION ALL ?) AS filter", eventUsers, eventParticipants).
		Joins("JOIN events e ON e.id = filter.event_id").
		Where(`NOW() > e.end_time`)

	if args.Search != "" {
		searchQuery := fmt.Sprintf("%%%s%%", args.Search)
		subQuery = subQuery.
			Where(`(e.name ILIKE ? OR e.organizer ILIKE ? OR e.description ILIKE ? OR e.location ILIKE ?
				OR e.evaluation_form ILIKE ?)
				`, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery)
	}

	if args.RoleFilterProvided {
		subQuery = subQuery.Where("(filter.role IS NULL OR filter.role IN (?))", args.Roles)
	}

	if args.Date != nil {
		dayStart := *args.Date
		dayEnd := dayStart.Add(24 * time.Hour)
		subQuery = subQuery.Where("e.start_time >= ? AND e.start_time < ?", dayStart, dayEnd)
	}

	var count int64
	countErr := withCtx.Raw(`SELECT COUNT(*) AS total FROM (?) AS subQuery`, subQuery).Scan(&count).Error
	if countErr != nil {
		return nil, -1, false, countErr
	}

	orderDirection := "DESC"
	if args.Sort == "oldest" {
		orderDirection = "ASC"
	}

	var rawResult []entity.GetEventsQueryResult
	getEventsErr := withCtx.Raw(fmt.Sprintf(`SELECT subQuery.* FROM (?) AS subQuery
		ORDER BY subQuery.start_time %s, subQuery.id
		OFFSET ?
		LIMIT ?
	`, orderDirection), subQuery, args.Page*args.PageSize, args.PageSize+1).Scan(&rawResult).Error
	if getEventsErr != nil {
		return nil, -1, false, getEventsErr
	}

	if len(rawResult) <= args.PageSize {
		return &rawResult, count, false, nil
	}
	clipped := rawResult[:args.PageSize]
	return &clipped, count, true, nil
}

func (r *repository) GetDiscoveryEvents(args *GetEventsArguments) (*[]entity.GetDiscoveryEvents, int64, bool, error) {
	withCtx := r.db.WithContext(args.Ctx)

	subQuery := withCtx.Table("events e").Select("e.id", "e.name", "e.organizer", "e.description", "e.start_time",
		"e.end_time", "e.location", "e.evaluation_form", "e.location_point", "e.allow_all_to_scan").
		Where(`NOT EXISTS (
			SELECT 1 FROM event_users eu WHERE eu.event_id = e.id
			AND eu.user_id = ?
		) AND NOT EXISTS (
			SELECT 1 FROM event_participants ep WHERE ep.event_id = e.id
			AND ep.participant_id = ?
		)`, args.UserID, args.UserID)

	if args.Search != "" {
		searchQuery := fmt.Sprintf("%%%s%%", args.Search)
		subQuery = subQuery.Where(`(e.name ILIKE ? OR e.organizer ILIKE ? OR e.description ILIKE ? OR e.location ILIKE ?
				OR e.evaluation_form ILIKE ?)
				`, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery)
	}

	var count int64
	countErr := withCtx.Raw(`SELECT COUNT(*) FROM (?) AS subQuery`, subQuery).Scan(&count).Error
	if countErr != nil {
		return nil, -1, false, countErr
	}

	var rawResult []entity.GetDiscoveryEvents
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
			SELECT e.start_time, e.end_time, e.attendence_type, e.allow_all_to_scan, e.revealed_fields,
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

func (r *repository) CheckEventAccess(ctx context.Context, orgCode *uint8, refID uint64, attendanceType string, eventId datatypes.UUID) (bool, error) {
	withCtx := r.db.WithContext(ctx)
	var found bool

	switch attendanceType {
	case string(entity.AttendanceFaculties):
		if orgCode == nil {
			// No numeric faculty code to compare (e.g. CU NEX returned a
			// non-numeric code, or the participant has no faculty at all) —
			// there's nothing for this event's allowed-faculty list to match.
			return false, nil
		}
		checkErr := withCtx.Raw(`SELECT EXISTS (
			SELECT 1 FROM event_allowed_faculties
			WHERE event_id = ? AND faculty_no = ?
		) AS subQuery`, eventId, orgCode).Scan(&found).Error
		if checkErr != nil {
			return false, checkErr
		}

		return found, nil

	case string(entity.AttendanceWhitelist):
		checkErr := withCtx.Raw(`SELECT (
            SELECT (
                EXISTS (
                    SELECT 1 FROM event_whitelists 
                    WHERE event_id = ? AND attendee_ref_id = ?
                )
                OR EXISTS (
                    SELECT 1 FROM event_whitelist_pendings
                    WHERE event_id = ? AND attendee_ref_id = ?
                )
            ) AS innerSubQuery
        ) AS subQuery`, eventId, refID, eventId, refID).Scan(&found).Error
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

func (r *repository) GetUserRoleInEvent(eventID uuid.UUID, userID uuid.UUID, ctx context.Context) (*string, error) {
	var eventUser entity.EventUser
	err := r.db.WithContext(ctx).Table("event_users").
		Select("role").
		Where("user_id = ? AND event_id = ?", userID, eventID).
		First(&eventUser).Error

	if err != nil {
		return nil, err
	}

	v := string(eventUser.Role)
	return &v, nil
}

func (r *repository) GetEventOwnerRefID(eventID uuid.UUID, ctx context.Context) (uint64, error) {
	var refID uint64
	err := r.db.WithContext(ctx).
		Table("event_users eu").
		Select("u.ref_id").
		Joins("JOIN users u ON u.id = eu.user_id").
		Where("eu.event_id = ? AND eu.role = ?", eventID, entity.OWNER).
		Take(&refID).Error

	if err != nil {
		return 0, err
	}
	return refID, nil
}

func (r *repository) GetEventEndTime(eventID uuid.UUID, ctx context.Context) (time.Time, error) {
	var endTime time.Time
	err := r.db.WithContext(ctx).
		Model(&entity.Event{}).
		Select("end_time").
		Where("id = ?", eventID).
		Take(&endTime).Error

	if err != nil {
		return time.Time{}, err
	}
	return endTime, nil
}

func (r *repository) CreateEvent(ctx context.Context, payload entity.CreateEventPayload) (*dtoRes.CreateEventRes, error) {
	var res dtoRes.CreateEventRes

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&payload.Event).Error; err != nil {
			return err
		}

		if len(payload.Agendas) > 0 {
			for i := range payload.Agendas {
				payload.Agendas[i].EventID = payload.Event.ID
			}
			if err := tx.Create(&payload.Agendas).Error; err != nil {
				return err
			}
		}

		if len(payload.Whitelist) > 0 {
			for i := range payload.Whitelist {
				payload.Whitelist[i].EventID = payload.Event.ID
			}

			wlOK, wlPend, err := splitWhitelistAndPending(ctx, tx, payload.Whitelist)
			if err != nil {
				return err
			}

			if len(wlOK) > 0 {
				if err := tx.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "event_id"}, {Name: "attendee_ref_id"}},
					DoNothing: true,
				}).Create(&wlOK).Error; err != nil {
					return err
				}
			}

			if len(wlPend) > 0 {
				if err := tx.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "event_id"}, {Name: "attendee_ref_id"}},
					DoNothing: true,
				}).Create(&wlPend).Error; err != nil {
					return err
				}
			}
		}

		if len(payload.AllowedFaculties) > 0 {
			for i := range payload.AllowedFaculties {
				payload.AllowedFaculties[i].EventID = payload.Event.ID
			}
			if err := tx.Create(&payload.AllowedFaculties).Error; err != nil {
				return err
			}
		}

		eventUsers, eventUsersPend, err := splitEventUserAndPending(ctx, tx, payload.Event.ID, payload.EventUsersInput)
		if err != nil {
			return err
		}
		if len(eventUsers) > 0 {
			if err := tx.Create(&eventUsers).Error; err != nil {
				return err
			}
		}
		if len(eventUsersPend) > 0 {
			if err := tx.Create(&eventUsersPend).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	res = dtoRes.CreateEventRes{
		ID: payload.Event.ID.String(),
	}
	return &res, nil
}

func (r *repository) UpdateEvent(ctx context.Context, id string, payload entity.CreateEventPayload) (*dtoRes.UpdateEventRes, error) {
	var res dtoRes.UpdateEventRes

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing entity.Event
		if err := tx.First(&existing, "id = ?", id).Error; err != nil {
			return err
		}

		if err := tx.Model(&existing).Updates(map[string]any{
			"name":              payload.Event.Name,
			"organizer":         payload.Event.Organizer,
			"description":       payload.Event.Description,
			"start_time":        payload.Event.StartTime,
			"end_time":          payload.Event.EndTime,
			"location":          payload.Event.Location,
			"location_point":    payload.Event.LocationPoint,
			"attendence_type":   payload.Event.AttendenceType,
			"allow_all_to_scan": payload.Event.AllowAllToScan,
			"evaluation_form":   payload.Event.EvaluationForm,
			"revealed_fields":   payload.Event.RevealedFields,
		}).Error; err != nil {
			return err
		}

		if err := tx.Where("event_id = ?", existing.ID).Delete(&entity.EventAgenda{}).Error; err != nil {
			return err
		}
		if err := tx.Where("event_id = ?", existing.ID).Delete(&entity.EventWhitelist{}).Error; err != nil {
			return err
		}
		if err := tx.Where("event_id = ?", existing.ID).Delete(&entity.EventWhitelistPending{}).Error; err != nil {
			return err
		}
		if err := tx.Where("event_id = ?", existing.ID).Delete(&entity.EventAllowedFaculties{}).Error; err != nil {
			return err
		}
		if err := tx.Where("event_id = ?", existing.ID).Delete(&entity.EventUser{}).Error; err != nil {
			return err
		}
		if err := tx.Where("event_id = ?", existing.ID).Delete(&entity.EventUserPending{}).Error; err != nil {
			return err
		}

		if len(payload.Agendas) > 0 {
			for i := range payload.Agendas {
				payload.Agendas[i].EventID = existing.ID
			}
			if err := tx.Create(&payload.Agendas).Error; err != nil {
				return err
			}
		}

		if len(payload.Whitelist) > 0 {
			for i := range payload.Whitelist {
				payload.Whitelist[i].EventID = existing.ID
			}

			wlOK, wlPend, err := splitWhitelistAndPending(ctx, tx, payload.Whitelist)
			if err != nil {
				return err
			}

			if len(wlOK) > 0 {
				if err := tx.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "event_id"}, {Name: "attendee_ref_id"}},
					DoNothing: true,
				}).Create(&wlOK).Error; err != nil {
					return err
				}
			}

			if len(wlPend) > 0 {
				if err := tx.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "event_id"}, {Name: "attendee_ref_id"}},
					DoNothing: true,
				}).Create(&wlPend).Error; err != nil {
					return err
				}
			}
		}

		if len(payload.AllowedFaculties) > 0 {
			for i := range payload.AllowedFaculties {
				payload.AllowedFaculties[i].EventID = existing.ID
			}
			if err := tx.Create(&payload.AllowedFaculties).Error; err != nil {
				return err
			}
		}

		eventUsers, eventUsersPend, err := splitEventUserAndPending(ctx, tx, existing.ID, payload.EventUsersInput)
		if err != nil {
			return err
		}
		if len(eventUsers) > 0 {
			if err := tx.Create(&eventUsers).Error; err != nil {
				return err
			}
		}
		if len(eventUsersPend) > 0 {
			if err := tx.Create(&eventUsersPend).Error; err != nil {
				return err
			}
		}

		res.ID = existing.ID.String()
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &res, nil
}

func splitWhitelistAndPending(ctx context.Context, tx *gorm.DB, wl []entity.EventWhitelist) ([]entity.EventWhitelist, []entity.EventWhitelistPending, error) {
	if len(wl) == 0 {
		return nil, nil, nil
	}

	refIDs := make([]uint64, 0, len(wl))
	seenRef := map[uint64]struct{}{}
	for _, x := range wl {
		if _, ok := seenRef[x.AttendeeRefID]; ok {
			continue
		}
		seenRef[x.AttendeeRefID] = struct{}{}
		refIDs = append(refIDs, x.AttendeeRefID)
	}

	var existing []uint64
	if err := tx.WithContext(ctx).
		Model(&entity.User{}).
		Where("ref_id IN ?", refIDs).
		Pluck("ref_id", &existing).Error; err != nil {
		return nil, nil, err
	}

	existSet := map[uint64]struct{}{}
	for _, id := range existing {
		existSet[id] = struct{}{}
	}

	okOut := make([]entity.EventWhitelist, 0, len(wl))
	pendOut := make([]entity.EventWhitelistPending, 0)

	seenPairOK := map[string]struct{}{}
	seenPairPend := map[string]struct{}{}

	for _, x := range wl {
		key := fmt.Sprintf("%s:%d", x.EventID.String(), x.AttendeeRefID)

		if _, ok := existSet[x.AttendeeRefID]; ok {
			if _, dup := seenPairOK[key]; dup {
				continue
			}
			seenPairOK[key] = struct{}{}
			okOut = append(okOut, x)
			continue
		}

		if _, dup := seenPairPend[key]; dup {
			continue
		}
		seenPairPend[key] = struct{}{}
		pendOut = append(pendOut, entity.EventWhitelistPending{
			EventID:       x.EventID,
			AttendeeRefID: x.AttendeeRefID,
		})
	}

	return okOut, pendOut, nil
}

func splitEventUserAndPending(ctx context.Context, tx *gorm.DB, eventID datatypes.UUID, in []entity.EventUserInput) ([]entity.EventUser, []entity.EventUserPending, error) {
	if len(in) == 0 {
		return nil, nil, nil
	}

	// dedup by ref_id (ถ้า ref_id ซ้ำแต่ role ต่างกัน -> error)
	dedup := make([]entity.EventUserInput, 0, len(in))
	seenRole := make(map[uint64]string, len(in))

	for _, x := range in {
		rs := string(x.Role) // role underlying type = string
		if old, ok := seenRole[x.RefID]; ok {
			if old != rs {
				return nil, nil, fmt.Errorf("duplicate ref_id with different role in managers_and_staff: %d", x.RefID)
			}
			continue
		}
		seenRole[x.RefID] = rs
		dedup = append(dedup, x)
	}

	refIDs := make([]uint64, 0, len(dedup))
	for _, x := range dedup {
		refIDs = append(refIDs, x.RefID)
	}

	var users []entity.User
	if err := tx.WithContext(ctx).
		Select("id", "ref_id").
		Where("ref_id IN ?", refIDs).
		Find(&users).Error; err != nil {
		return nil, nil, err
	}

	userFromDB := make(map[uint64]entity.User, len(users))
	for _, u := range users {
		userFromDB[u.RefID] = u
	}

	outOK := make([]entity.EventUser, 0, len(userFromDB))
	outPend := make([]entity.EventUserPending, 0)
	for _, x := range dedup {
		u, ok := userFromDB[x.RefID]
		if !ok {
			outPend = append(outPend, entity.EventUserPending{
				EventID:   eventID,
				UserRefID: x.RefID,
				Role:      x.Role,
			})
			continue
		}
		outOK = append(outOK, entity.EventUser{
			EventID: eventID,
			UserID:  u.ID,
			Role:    x.Role, // ✅ ใช้ role value เดิมได้เลย
		})
	}

	return outOK, outPend, nil
}
