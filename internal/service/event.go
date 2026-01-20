package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	dtoReq "github.com/cunex-club/quickattend-backend/internal/dto/request"
	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
	"github.com/cunex-club/quickattend-backend/internal/entity"
)

var thaiLoc = time.FixedZone(entity.ThaiTZ, 7*3600)

type EventService interface {
	CreateEvent(ctx context.Context, req dtoReq.CreateEventReq) (*dtoRes.CreateEventRes, error)
	UpdateEvent(ctx context.Context, id string, updates dtoReq.UpdateEventReq) (*dtoRes.UpdateEventRes, error)
}

func (s *service) CreateEvent(ctx context.Context, req dtoReq.CreateEventReq) (*dtoRes.CreateEventRes, error) {
	payload, err := buildCreateOrUpdatePayload(req)
	if err != nil {
		return nil, err
	}
	return s.repo.Event.CreateEvent(ctx, payload)
}

func (s *service) UpdateEvent(ctx context.Context, id string, req dtoReq.UpdateEventReq) (*dtoRes.UpdateEventRes, error) {
	payload, err := buildCreateOrUpdatePayload(dtoReq.CreateEventReq(req))
	if err != nil {
		return nil, err
	}
	return s.repo.Event.UpdateEvent(ctx, id, payload)
}

func buildCreateOrUpdatePayload(req dtoReq.CreateEventReq) (entity.CreateEventPayload, error) {
	if err := validateThaiTimezone(req.Timezone); err != nil {
		return entity.CreateEventPayload{}, err
	}

	if req.AllowAllToScan == nil {
		return entity.CreateEventPayload{}, fmt.Errorf("allow_all_to_scan is required")
	}

	at, err := entity.ParseAttendanceType(req.AttendanceType)
	if err != nil {
		return entity.CreateEventPayload{}, err
	}

	// Time parsing and validation
	startTime, err := parseTime(req.StartTime)
	if err != nil {
		return entity.CreateEventPayload{}, err
	}
	endTime, err := parseTime(req.EndTime)
	if err != nil {
		return entity.CreateEventPayload{}, err
	}

	if !endTime.After(startTime) {
		return entity.CreateEventPayload{}, fmt.Errorf("end_time must be after start_time")
	}
	if !isSameDay(startTime, endTime) {
		return entity.CreateEventPayload{}, fmt.Errorf("start_time and end_time must be on the same day in timezone %s", entity.ThaiTZ)
	}

	// attendance_type=all -> attendee must be empty
	if at == entity.AttendanceAll && len(req.Attendee) != 0 {
		return entity.CreateEventPayload{}, fmt.Errorf("attendee must be empty when attendance_type=all")
	}

	revealedFields, err := entity.ParseParticipantFields(req.RevealedFields)
	if err != nil {
		return entity.CreateEventPayload{}, err
	}

	event := entity.Event{
		Name:        req.Name,
		Organizer:   req.Organizer,
		Description: req.Description,

		StartTime: startTime,
		EndTime:   endTime,

		Location:       req.Location,
		AttendanceType: at,
		AllowAllToScan: *req.AllowAllToScan,
		EvaluationForm: req.EvaluationForm,
		RevealedFields: revealedFields,
	}

	agendas, err := buildAgendas(req.Agenda, startTime, endTime)
	if err != nil {
		return entity.CreateEventPayload{}, err
	}

	whitelist, faculties, err := buildAttendanceTargets(at, req.Attendee)
	if err != nil {
		return entity.CreateEventPayload{}, err
	}

	// managers_and_staff -> EventUsersInput (ref_id + parsed role)
	eventUsersInput, err := buildEventUsersInput(req.ManagersAndStaff)
	if err != nil {
		return entity.CreateEventPayload{}, err
	}

	return entity.CreateEventPayload{
		Event:            event,
		Agendas:          agendas,
		Whitelist:        whitelist,
		AllowedFaculties: faculties,
		EventUsersInput:  eventUsersInput,
	}, nil

}

func buildEventUsersInput(in []dtoReq.ManagerStaffReq) ([]entity.EventUserInput, error) {
	if len(in) == 0 {
		return nil, nil
	}

	out := make([]entity.EventUserInput, 0, len(in))
	seen := map[string]struct{}{}

	for _, m := range in {
		r, err := entity.ParseRole(m.Role)
		if err != nil {
			return nil, err
		}

		key := fmt.Sprintf("%d:%s", m.RefID, r)
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}

		out = append(out, entity.EventUserInput{
			RefID: m.RefID,
			Role:  r,
		})
	}
	return out, nil
}

// validateThaiTimezone enforces client timezone to be Asia/Bangkok only.
func validateThaiTimezone(tz string) error {
	if strings.TrimSpace(tz) != entity.ThaiTZ {
		return fmt.Errorf("timezone must be %s", entity.ThaiTZ)
	}
	return nil
}

// parseTime parses RFC3339 and enforces that it is UTC (Z / +00:00).
func parseTime(v string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(v))
	if err != nil {
		return time.Time{}, err
	}
	_, off := t.Zone()
	if off != 0 {
		return time.Time{}, fmt.Errorf("time must be UTC (use Z / +00:00)")
	}
	return t.UTC(), nil
}

// isSameDay checks whether two UTC instants are on the same calendar day in Thai timezone.
func isSameDay(aUTC, bUTC time.Time) bool {
	a := aUTC.In(thaiLoc)
	b := bUTC.In(thaiLoc)
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// isSameDay checks whether two UTC instants are on the same calendar day in Thai timezone.
func buildAgendas(in []dtoReq.CreateAgendaReq, eventStartUTC, eventEndUTC time.Time) ([]entity.EventAgenda, error) {
	if len(in) == 0 {
		return nil, nil
	}

	out := make([]entity.EventAgenda, 0, len(in))
	for _, a := range in {
		st, err := parseTime(a.StartTime)
		if err != nil {
			return nil, err
		}
		et, err := parseTime(a.EndTime)
		if err != nil {
			return nil, err
		}

		if !et.After(st) {
			return nil, fmt.Errorf("agenda end_time must be after start_time")
		}
		if !isSameDay(st, et) {
			return nil, fmt.Errorf("agenda start_time and end_time must be on the same day in timezone %s", entity.ThaiTZ)
		}
		if st.Before(eventStartUTC) || et.After(eventEndUTC) {
			return nil, fmt.Errorf("agenda time must be within event start_time and end_time")
		}

		out = append(out, entity.EventAgenda{
			ActivityName: a.ActivityName,
			StartTime:    st,
			EndTime:      et,
		})
	}
	return out, nil
}

// buildAttendanceTargets converts attendee list into whitelist/faculty rows based on attendance_type.
func buildAttendanceTargets(at entity.AttendanceType, attendee []any) ([]entity.EventWhitelist, []entity.EventAllowedFaculties, error) {
	switch at {
	case entity.AttendanceAll:
		return nil, nil, nil

	case entity.AttendanceWhitelist:
		if len(attendee) == 0 {
			return nil, nil, fmt.Errorf("attendee is required")
		}
		out := make([]entity.EventWhitelist, 0, len(attendee))
		for _, v := range attendee {
			ref, err := anyToUint64(v)
			if err != nil {
				return nil, nil, err
			}
			out = append(out, entity.EventWhitelist{AttendeeRefID: ref})
		}
		return out, nil, nil

	case entity.AttendanceFaculties:
		if len(attendee) == 0 {
			return nil, nil, fmt.Errorf("attendee is required")
		}
		out := make([]entity.EventAllowedFaculties, 0, len(attendee))
		for _, v := range attendee {
			fno, err := anyToUint8(v)
			if err != nil {
				return nil, nil, err
			}
			out = append(out, entity.EventAllowedFaculties{FacultyNO: fno})
		}
		return nil, out, nil

	default:
		return nil, nil, fmt.Errorf("invalid attendance_type")
	}
}

// anyToUint64 converts JSON number/string into uint64.
func anyToUint64(v any) (uint64, error) {
	switch x := v.(type) {
	case uint64:
		return x, nil
	case int:
		if x < 0 {
			return 0, fmt.Errorf("invalid number")
		}
		return uint64(x), nil
	case int64:
		if x < 0 {
			return 0, fmt.Errorf("invalid number")
		}
		return uint64(x), nil
	case float64:
		if x < 0 {
			return 0, fmt.Errorf("invalid number")
		}
		return uint64(x), nil
	case string:
		u, err := strconv.ParseUint(strings.TrimSpace(x), 10, 64)
		if err != nil {
			return 0, err
		}
		return u, nil
	default:
		return 0, fmt.Errorf("invalid type: %T", v)
	}
}

// anyToUint8 converts JSON number/string into uint8.
func anyToUint8(v any) (uint8, error) {
	u, err := anyToUint64(v)
	if err != nil {
		return 0, err
	}
	if u > 255 {
		return 0, fmt.Errorf("out of range")
	}
	return uint8(u), nil
}
