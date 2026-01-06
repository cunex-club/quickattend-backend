package repository

import (
	"context"

	"github.com/cunex-club/quickattend-backend/internal/entity"
	"gorm.io/datatypes"
)

type EventRepository interface {
	// Get user info not provided by CU NEX, and check attendance type of the event
	GetParticipantUserInfoAndAttendanceType(ctx context.Context, eventId datatypes.UUID, refID uint64) (user *entity.User, attendanceType string, err error)
	// Check if user has already check in to the event
	GetParticipantCheckParticipation(ctx context.Context, eventId datatypes.UUID, refID uint64) (found bool, err error)
	// Check if user is in whitelist / allowed org or faculty of the event
	GetParticipantCheckAccess(ctx context.Context, orgCode int64, attendanceType string, eventId datatypes.UUID) (allow bool, err error)
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
	checkParticipantErr := withCtx.Select(`EXISTS (
				SELECT 1 FROM event_participants
				WHERE event_participants.participant_ref_id = ?
				AND event_participants.event_id = ?
			)`, refID, eventId).Scan(&found).Error
	if checkParticipantErr != nil {
		return false, checkParticipantErr
	}

	return found, nil
}

func (r *repository) GetParticipantCheckAccess(ctx context.Context, orgCode int64, attendanceType string, eventId datatypes.UUID) (bool, error) {
	withCtx := r.db.WithContext(ctx)

	if attendanceType == string(entity.FACULTIES) {
		var facultyNoList []int64
		getListErr := withCtx.Select("event_allowed_faculties.faculty_no").
			Where("event_allowed_faculties.event_id = ?", eventId).
			Scan(&facultyNoList).Error
		if getListErr != nil {
			return false, getListErr
		}

		if len(facultyNoList) == 0 {
			return false, nil
		} else {
			allowed := false
			for _, no := range facultyNoList {
				if orgCode == no {
					allowed = true
				}
			}
			if !allowed {
				return false, nil
			}
		}

		return true, nil
	}

	var whiteList []uint64
	getListErr := withCtx.Select("event_whitelists.attendee_ref_id").
		Where("event_whitelists.event_id = ?", eventId).
		Scan(&whiteList).Error
	if getListErr != nil {
		return false, getListErr
	}

	if len(whiteList) == 0 {
		return false, nil
	} else {
		allowed := false
		for _, no := range whiteList {
			if orgCode == int64(no) {
				allowed = true
			}
		}
		if !allowed {
			return false, nil
		}
	}

	return true, nil
}
