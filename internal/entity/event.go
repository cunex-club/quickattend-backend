package entity

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"gorm.io/datatypes"
)

// ====================================================

type AttendanceType string

const (
	AttendanceAll       AttendanceType = "all"
	AttendanceWhitelist AttendanceType = "whitelist"
	AttendanceFaculties AttendanceType = "faculties"
)

func (at *AttendanceType) Scan(value any) error {
	if value == nil {
		*at = ""
		return nil
	}

	switch v := value.(type) {
	case string:
		*at = AttendanceType(v)
		return nil
	case []byte:
		*at = AttendanceType(string(v))
		return nil
	default:
		return fmt.Errorf("cannot scan %T into AttendanceType", value)
	}
}

func (at AttendanceType) Value() (driver.Value, error) {
	return string(at), nil
}

func ParseAttendanceType(s string) (AttendanceType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "all":
		return AttendanceAll, nil
	case "whitelist":
		return AttendanceWhitelist, nil
	case "faculties":
		return AttendanceFaculties, nil
	default:
		return "", fmt.Errorf("invalid attendance_type")
	}
}

func (at AttendanceType) Valid() bool {
	switch at {
	case AttendanceAll, AttendanceWhitelist, AttendanceFaculties:
		return true
	default:
		return false
	}
}

// ====================================================

type ParticipantData string

const (
	ParticipantName         ParticipantData = "name"
	ParticipantOrganization ParticipantData = "organization"
	ParticipantRefID        ParticipantData = "refid"
	ParticipantPhoto        ParticipantData = "photo"
)

type ParticipantField []ParticipantData

func (pf *ParticipantField) Scan(value any) error {
	if value == nil {
		*pf = nil
		return nil
	}

	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("Error scanning participant_field")
	}

	str = strings.TrimSpace(str)
	str = strings.TrimPrefix(str, "{")
	str = strings.TrimSuffix(str, "}")

	if str == "" {
		*pf = nil
		return nil
	}

	items := strings.Split(str, ",")
	out := make([]ParticipantData, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		out = append(out, ParticipantData(item))
	}

	*pf = out
	return nil
}

func (pf ParticipantField) Value() (driver.Value, error) {
	if len(pf) == 0 {
		return "{}", nil
	}

	items := make([]string, 0, len(pf))
	for _, p := range pf {
		items = append(items, string(p))
	}

	return fmt.Sprintf("{%s}", strings.Join(items, ",")), nil
}

func ParseParticipantData(s string) (ParticipantData, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "name":
		return ParticipantName, nil
	case "organization":
		return ParticipantOrganization, nil
	case "refid":
		return ParticipantRefID, nil
	case "photo":
		return ParticipantPhoto, nil
	default:
		return "", fmt.Errorf("invalid participant field: %s", s)
	}
}

func ParseParticipantFields(fields []string) (ParticipantField, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("revealed_fields is required")
	}

	out := make(ParticipantField, 0, len(fields))
	for _, f := range fields {
		p, err := ParseParticipantData(f)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}

	return out, nil
}

// ====================================================

type Point struct {
	X float64
	Y float64
}

func (p *Point) Scan(val any) error {
	var point string
	switch v := val.(type) {
	case []byte:
		point = string(v)
	case string:
		point = v
	default:
		return fmt.Errorf("cannot convert %T to Point", val)
	}

	_, err := fmt.Sscanf(point, "(%f,%f)", &p.X, &p.Y)
	return err
}

func (p Point) Value() (driver.Value, error) {
	return fmt.Sprintf("(%f,%f)", p.X, p.Y), nil
}

// ====================================================
const ThaiTZ = "Asia/Bangkok"

type Event struct {
	ID             datatypes.UUID    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name           string            `gorm:"type:text;not null;index:idx_events_name_trgm,type:gin" json:"name"`
	Organizer      string            `gorm:"type:text;not null;index:idx_events_organizer_trgm,type:gin" json:"organizer"`
	Description    *string           `gorm:"type:text;index:idx_events_description_trgm,type:gin" json:"description"`
	StartTime      time.Time         `gorm:"type:timestamptz;not null" json:"start_time"`
	EndTime        time.Time         `gorm:"type:timestamptz;not null" json:"end_time"`
	Location       string            `gorm:"type:text;not null;index:idx_events_location_trgm,type:gin" json:"location"`
	AttendenceType AttendanceType   `gorm:"type:attendence_type;not null" json:"attendance_type"`
	AllowAllToScan bool              `gorm:"type:bool;not null" json:"allow_all_to_scan"`
	EvaluationForm *string           `gorm:"type:text;index:idx_events_evaluation_form_trgm,type:gin" json:"evaluation_form"`
	RevealedFields ParticipantField `gorm:"type:participant_data[];not null" json:"revealed_fields"`
}

type EventWhitelist struct {
	ID            datatypes.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	EventID       datatypes.UUID `gorm:"type:uuid;not null;index:unique_event_and_ref_id,unique" json:"event_id"`
	AttendeeRefID uint64         `gorm:"type:bigint;not null;index:unique_event_and_ref_id,unique" json:"attendee_ref_id"`

	Event Event `gorm:"foreignKey:EventID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	User  User  `gorm:"foreignKey:AttendeeRefID;references:RefID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type EventWhitelistPending struct {
	ID            datatypes.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	EventID       datatypes.UUID `gorm:"type:uuid;not null;index;uniqueIndex:uniq_event_pending_attendee" json:"event_id"`
	AttendeeRefID uint64         `gorm:"type:bigint;not null;index;uniqueIndex:uniq_event_pending_attendee" json:"attendee_ref_id"`

	Event Event `gorm:"foreignKey:EventID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type EventAllowedFaculties struct {
	ID        datatypes.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	EventID   datatypes.UUID `gorm:"type:uuid;not null;index:unique_event_and_faculty_no,unique" json:"event_id"`
	FacultyNO uint8          `gorm:"type:int8;not null;index:unique_event_and_faculty_no,unique" json:"faculty_no"`

	Event Event `gorm:"foreignKey:EventID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type EventAgenda struct {
	ID           datatypes.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	EventID      datatypes.UUID `gorm:"type:uuid;not null;index:unique_event_start_end,unique" json:"event_id"`
	ActivityName string         `gorm:"type:text;not null" json:"activity_name"`
	StartTime    time.Time      `gorm:"type:timestamptz;not null;index:unique_event_start_end,unique" json:"start_time"`
	EndTime      time.Time      `gorm:"type:timestamptz;not null;index:unique_event_start_end,unique" json:"end_time"`

	Event Event `gorm:"foreignKey:EventID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type EventParticipants struct {
	ID               datatypes.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	EventID          datatypes.UUID  `gorm:"type:uuid;not null;index:unique_event_and_participant,unique" json:"event_id"`
	CheckinTimestamp *time.Time      `gorm:"type:timestamptz" json:"checkin_timestamp"`
	ScannedTimestamp time.Time       `gorm:"type:timestamptz;not null" json:"scanned_timestamp"`
	Comment          *string         `gorm:"type:text" json:"comment"`
	ParticipantID    datatypes.UUID  `gorm:"type:uuid;not null;index:unique_event_and_participant,unique" json:"participant_id"`
	Organization     string          `gorm:"type:text;not null" json:"organization"`
	ScannedLocation  Point           `gorm:"type:point;not null" json:"scanned_location"`
	ScannerID        *datatypes.UUID `gorm:"type:uuid" json:"scanner_id"`

	Event                      Event `gorm:"foreignKey:EventID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ParticipantRefIDForeignKey User  `gorm:"foreignKey:ParticipantRefID;references:RefID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ScannerIDForeignKey        User  `gorm:"foreignKey:ScannerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

// ====================================================

// for retrieving agenda query result in GET /events/:id
type GetOneEventAgenda struct {
	ActivityName string    `gorm:"column:activity_name"`
	StartTime    time.Time `gorm:"column:start_time"`
	EndTime      time.Time `gorm:"column:end_time"`
}

// for retrieving event details and total participant count in GET /events/:id
type GetOneEventWithTotalCount struct {
	Name            string    `gorm:"column:name"`
	Organizer       string    `gorm:"column:organizer"`
	Description     *string   `gorm:"column:description"`
	StartTime       time.Time `gorm:"column:start_time"`
	EndTime         time.Time `gorm:"column:end_time"`
	Location        string    `gorm:"column:location"`
	TotalRegistered uint16    `gorm:"column:total_registered"`
	EvaluationForm  *string   `gorm:"column:evaluation_form"`
	Role            *string   `gorm:"column:role"`
}

// ====================================================

// for retrieving raw result from DB in GET /events
type GetEventsQueryResult struct {
	ID             datatypes.UUID `gorm:"column:id"`
	Name           string         `gorm:"column:name"`
	Organizer      string         `gorm:"column:organizer"`
	Description    *string        `gorm:"column:description"`
	StartTime      time.Time      `gorm:"column:start_time"`
	EndTime        time.Time      `gorm:"column:end_time"`
	Location       string         `gorm:"column:location"`
	Role           *string        `gorm:"column:role"`
	EvaluationForm *string        `gorm:"column:evaluation_form"`
}

type CreateEventPayload struct {
	Event            Event
	Agendas          []EventAgenda
	Whitelist        []EventWhitelist
	AllowedFaculties []EventAllowedFaculties
	EventUsersInput  []EventUserInput
}
