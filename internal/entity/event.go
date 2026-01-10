package entity

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"gorm.io/datatypes"
)

// ====================================================

type attendence_type string

const (
	WHITELIST attendence_type = "WHITELIST"
	FACULTIES attendence_type = "FACULTIES"
	ALL       attendence_type = "ALL"
)

func (at *attendence_type) Scan(value any) error {
	*at = attendence_type(value.(string))
	return nil
}

func (at attendence_type) Value() (driver.Value, error) {
	return string(at), nil
}

// ====================================================

type participant_data string

const (
	NAME         participant_data = "NAME"
	ORGANIZATION participant_data = "ORGANIZATION"
	REFID        participant_data = "REFID"
	PHOTO        participant_data = "PHOTO"
)

type participant_field []participant_data

func (pf *participant_field) Scan(value any) error {
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

	str = strings.Trim(str, "{}")

	if value == "" {
		*pf = nil
		return nil
	}

	items := strings.Split(str, ",")
	out := make([]participant_data, len(items))
	for i, item := range items {
		out[i] = participant_data(strings.TrimSpace(item))
	}

	*pf = out
	return nil
}

func (pf participant_field) Value() (driver.Value, error) {
	if len(pf) == 0 {
		return "{}", nil
	}

	items := make([]string, len(pf))

	for i, p := range pf {
		items[i] = string(p)
	}

	return fmt.Sprintf("{%s}", strings.Join(items, ",")), nil
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

type Event struct {
	ID                    datatypes.UUID          `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name                  string                  `gorm:"type:text;not null" json:"name"`
	Organizer             string                  `gorm:"type:text;not null" json:"organizer"`
	Description           string                  `gorm:"type:text" json:"description"`
	Date                  datatypes.Date          `gorm:"type:timestamp;not null" json:"date"`
	StartTime             datatypes.Time          `gorm:"type:time;not null" json:"start_time"`
	EndTime               datatypes.Time          `gorm:"type:time;not null" json:"end_time"`
	Location              string                  `gorm:"type:text;not null" json:"location"`
	AttendenceType        attendence_type         `gorm:"type:attendence_type;not null" json:"attendance_type"`
	AllowAllToScan        bool                    `gorm:"type:bool;not null" json:"allow_all_to_scan"`
	EvaluationForm        string                  `gorm:"type:text" json:"evaluation_form"`
	RevealedFields        participant_field       `gorm:"type:[]participant_data;not null" json:"revealed_fields"`
	EventWhitelist        []EventWhitelist        `gorm:"foreignKey:EventID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"event_whitelist"`
	EventAllowedFaculties []EventAllowedFaculties `gorm:"foreignKey:EventID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"event_allowed_faculties"`
	EventAgenda           []EventAgenda           `gorm:"foreignKey:EventID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"event_agenda"`
}

type EventWhitelist struct {
	ID            datatypes.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	EventID       datatypes.UUID `gorm:"type:uuid;not null" json:"event_id"`
	AttendeeRefID uint64         `gorm:"type:bigint;not null" json:"attendee_ref_id"`

	Event Event `gorm:"foreignKey:EventID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	User  User  `gorm:"foreignKey:AttendeeRefID;references:RefID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type EventAllowedFaculties struct {
	ID        datatypes.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	EventID   datatypes.UUID `gorm:"type:uuid;not null" json:"event_id"`
	FacultyNO uint8          `gorm:"type:int8;not null" json:"faculty_no"`

	Event Event `gorm:"foreignKey:EventID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type EventAgenda struct {
	ID           datatypes.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	EventID      datatypes.UUID `gorm:"type:uuid;not null" json:"event_id"`
	ActivityName string         `gorm:"type:text;not null" json:"activity_name"`
	StartTime    datatypes.Time `gorm:"type:time;not null" json:"start_time"`
	EndTime      datatypes.Time `gorm:"type:time;not null" json:"end_time"`

	Event Event `gorm:"foreignKey:EventID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type EventParticipants struct {
	ID               datatypes.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	EventID          datatypes.UUID `gorm:"type:uuid;not null" json:"event_id"`
	CheckinTimestamp time.Time      `gorm:"type:timestamptz;not null" json:"checkin_timestamp"`
	ParticipantRefID uint64         `gorm:"type:bigint;not null" json:"participant_ref_id"`
	FirstName        string         `gorm:"type:text;not null" json:"first_name"`
	SurName          string         `gorm:"type:text;not null" json:"sur_name"`
	Organization     string         `gorm:"type:text;not null" json:"organization"`
	ScannedLocation  Point          `gorm:"type:point;not null" json:"scanned_location"`
	ScannerID        datatypes.UUID `gorm:"type:uuid" json:"scanner_id"`

	Event                      Event `gorm:"foreignKey:EventID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ParticipantRefIDForeignKey User  `gorm:"foreignKey:UserRefID;references:RefID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ScannerIDForeignKey        User  `gorm:"foreignKey:ScannerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}
