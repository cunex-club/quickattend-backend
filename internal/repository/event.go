package repository

import (
	"context"
	"errors"

	"github.com/cunex-club/quickattend-backend/internal/entity"
	"gorm.io/datatypes"
)

type EventRepository interface {
	// Get user info not provided by CU NEX, and check attendance type of the event
	GetParticipantUserInfoAndAttendanceType(ctx context.Context, eventId datatypes.UUID, refID uint64) (user *entity.User, attendanceType string, err error)
	// Check if user has already check in to the event
	GetParticipantCheckParticipation(ctx context.Context, eventId datatypes.UUID, refID uint64) (found bool, err error)
	// Check if user is in whitelist / allowed org or faculty of the event
	GetParticipantCheckAccess(ctx context.Context, orgCode int64, refID uint64, attendanceType string, eventId datatypes.UUID) (allow bool, err error)
}

func (r *repository) GetParticipantUserInfoAndAttendanceType(ctx context.Context, eventId datatypes.UUID, refID uint64) (*entity.User, string, error) {
	withCtx := r.db.WithContext(ctx)

	var user entity.User
	getUserErr := withCtx.Model(&user).Select("firstname_th", "surname_th", "title_th", "title_en").
		First(&user, &entity.User{RefID: refID}).Error
	if getUserErr != nil {
		return nil, "", getUserErr
	}

	var event entity.Event
	getAttendanceTypeErr := withCtx.Model(&event).Select("attendence_type").First(&event, &entity.Event{ID: eventId}).Error
	if getAttendanceTypeErr != nil {
		return nil, "", getAttendanceTypeErr
	}

	return &user, string(event.AttendenceType), nil
}

func (r *repository) GetParticipantCheckParticipation(ctx context.Context, eventId datatypes.UUID, refID uint64) (bool, error) {
	withCtx := r.db.WithContext(ctx)

	var found bool
	checkParticipantErr := withCtx.Raw(`SELECT EXISTS (
				SELECT 1 FROM event_participants
				WHERE participant_ref_id = ? AND event_id = ?
			) AS subQuery`, refID, eventId).Scan(&found).Error
	if checkParticipantErr != nil {
		return false, checkParticipantErr
	}

	return found, nil
}

func (r *repository) GetParticipantCheckAccess(ctx context.Context, orgCode int64, refID uint64, attendanceType string, eventId datatypes.UUID) (bool, error) {
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
