package repository

import (
	"context"

	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
	"github.com/cunex-club/quickattend-backend/internal/entity"
	"gorm.io/datatypes"
)

type EventRepository interface {
	GetParticipantRepository(ctx context.Context, eventId datatypes.UUID, refID uint64, orgCode int64) (*entity.User, string, error)
}

// returns (user, status, error)
func (r *repository) GetParticipantRepository(ctx context.Context, eventId datatypes.UUID, refID uint64, orgCode int64) (*entity.User, string, error) {
	var user entity.User
	withCtx := r.db.WithContext(ctx)

	// need to check user status on the event too
	// First, get attendence_type of the event
	// IF attendence_type = ALL: join "users" on "event_participants" -> if present = "duplicate"
	// ELSE IF attendence_type = WHITELISTS: join "users" on "event_whitelists" -> if refID not in table = "fail"
	// ELSE IF attendence_type = FACULTIES: join "users" on "event_allowed_faculties" -> if convertToFacultyCode(refID) not in any of faculty_row table = "fail"
	//     - need user type from CU NEX API (student or staff) to accurately convert to faculty/org code?
	// ELSE: throw error (unknown attendence_type, shouldn't happen but just in case)
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

	checkParticipantSubquery := withCtx.Select(`EXISTS (
				SELECT 1 FROM event_participants
				WHERE event_participants.participant_ref_id = ?
				AND event_participants.event_id = ?
			)`, refID, eventId)
	var foundParticipant struct {
		found bool `gorm:"column:found"`
	}

	switch event.AttendenceType {
	case entity.ALL:
		checkParticipantErr := withCtx.Raw(`SELECT * AS found FROM (?) AS subQuery`, checkParticipantSubquery).
			Scan(&foundParticipant.found).Error
		if checkParticipantErr != nil {
			return nil, "", checkParticipantErr
		}

		if !foundParticipant.found {
			return &user, string(dtoRes.SUCCESS), nil
		}
		return &user, string(dtoRes.DUPLICATE), nil

	case entity.FACULTIES:
		var facultyNoList []int64
		getListErr := withCtx.Select("event_allowed_faculties.faculty_no").
			Where("event_allowed_faculties.event_id = ?", eventId).
			Scan(&facultyNoList).Error
		if getListErr != nil {
			return nil, "", getListErr
		}

		if len(facultyNoList) == 0 {
			return &user, string(dtoRes.FAIL), nil
		} else {
			allowed := false
			for _, no := range facultyNoList {
				if orgCode == no {
					allowed = true
				}
			}
			if !allowed {
				return &user, string(dtoRes.FAIL), nil
			}
		}

		checkParticipantErr := withCtx.Raw(`SELECT * AS found FROM (?) AS subQuery`, checkParticipantSubquery).
			Scan(&foundParticipant.found).Error
		if checkParticipantErr != nil {
			return nil, "", checkParticipantErr
		}

		if !foundParticipant.found {
			return &user, string(dtoRes.SUCCESS), nil
		}
		return &user, string(dtoRes.DUPLICATE), nil

	case entity.WHITELIST:
		var whiteList []uint64
		getListErr := withCtx.Select("event_whitelists.attendee_ref_id").
			Where("event_whitelists.event_id = ?", eventId).
			Scan(&whiteList).Error
		if getListErr != nil {
			return nil, "", getListErr
		}

		if len(whiteList) == 0 {
			return &user, string(dtoRes.FAIL), nil
		} else {
			allowed := false
			for _, no := range whiteList {
				if orgCode == int64(no) {
					allowed = true
				}
			}
			if !allowed {
				return &user, string(dtoRes.FAIL), nil
			}
		}

		checkParticipantErr := withCtx.Raw(`SELECT * AS found FROM (?) AS subQuery`, checkParticipantSubquery).
			Scan(&foundParticipant.found).Error
		if checkParticipantErr != nil {
			return nil, "", checkParticipantErr
		}

		if !foundParticipant.found {
			return &user, string(dtoRes.SUCCESS), nil
		}
		return &user, string(dtoRes.DUPLICATE), nil

	default:
		return nil, "", nil
	}
}
