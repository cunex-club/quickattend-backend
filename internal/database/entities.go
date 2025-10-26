package database

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"gorm.io/datatypes"
)

// ====================================================

type role string

const (
	OWNER   role = "OWNER"
	STAFF   role = "STAFF"
	MANAGER role = "MANAGER"
)

func (r *role) Scan(value any) error {
	*r = role(value.(string))
	return nil
}

func (r role) Value() (driver.Value, error) {
	return string(r), nil
}

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

type User struct {
	ID          datatypes.UUID `gorm:"default:uuid_generate_v4();primaryKey"`
	RefID       uint8
	FirstnameTH string
	SurnameTH   string
	TitleTH     string
	FirstnameEN string
	SurnameEN   string
	TitleEN     string
}

type EventUsers struct {
	ID      datatypes.UUID `gorm:"default:uuid_generate_v4();primaryKey"`
	role    role           `gorm:"type:role"`
	UserID  datatypes.UUID
	EventID datatypes.UUID
}

type Event struct {
	ID             datatypes.UUID `gorm:"default:uuid_generate_v4();primaryKey"`
	Name           string
	Organizer      string
	Description    string
	Date           datatypes.Date
	StartTime      datatypes.Time
	EndTime        datatypes.Time
	Location       string
	AttendenceType attendence_type `gorm:"type:attendence_type"`
	AllowAllToScan bool
	EvaluationForm string
	RevealedField  participant_field `gorm:"type:[]participant_data"`
}

type EventWhitelist struct {
	ID            datatypes.UUID `gorm:"default:uuid_generate_v4();primaryKey"`
	EventID       datatypes.UUID
	AttendeeRefID datatypes.UUID
}

type EventAllowedFaculties struct {
	ID        datatypes.UUID `gorm:"default:uuid_generate_v4();primaryKey"`
	EventID   datatypes.UUID
	FacultyNO int8
}

type EventAgenda struct {
	ID           datatypes.UUID `gorm:"default:uuid_generate_v4();primaryKey"`
	EventID      datatypes.UUID // foreign key
	ActivityName string
	StartTime    datatypes.Time
	EndTime      datatypes.Time

	Event Event `gorm:"foreignKey:EventID;references:ID"`
}

type EventParticipants struct {
	ID               datatypes.UUID `gorm:"default:uuid_generate_v4();primaryKey"`
	EventID          datatypes.UUID // foreign key
	CheckinTimestamp time.Time
	ParticipantRefID uint8
	UserRefID        uint8 // foreign key
	FirstName        string
	SurName          string
	Organization     string
	ScannedLocation  Point          `gorm:"type:point"`
	ScannerID        datatypes.UUID // foreign key

	Event                      Event `gorm:"foreignKey:EventID;references:ID"` // For column `EventID`, refer to `ID` of `Event` table
	ParticipantRefIDForeignKey User  `gorm:"foreignKey:UserRefID;references:RefID"`
	ScannerIDForeignKey        User  `gorm:"foreignKey:ScannerID;references:ID"`
}
