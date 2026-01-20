package entity

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"gorm.io/datatypes"
)

type role string

const (
	OWNER   role = "OWNER"
	STAFF   role = "STAFF"
	MANAGER role = "MANAGER"
)

func (r *role) Scan(value any) error {
	if value == nil {
		*r = ""
		return nil
	}
	switch v := value.(type) {
	case string:
		*r = role(v)
		return nil
	case []byte:
		*r = role(string(v))
		return nil
	default:
		return fmt.Errorf("cannot scan %T into role", value)
	}
}

func (r role) Value() (driver.Value, error) {
	return string(r), nil
}

func ParseRole(s string) (role, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "owner":
		return OWNER, nil
	case "staff":
		return STAFF, nil
	case "manager":
		return MANAGER, nil
	default:
		return "", fmt.Errorf("invalid role")
	}
}

// ====================================================

type User struct {
	ID          datatypes.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	RefID       uint64         `gorm:"type:bigint;not null" json:"ref_id"`
	FirstnameTH string         `gorm:"type:text;not null" json:"firstname_th"`
	SurnameTH   string         `gorm:"type:text;not null" json:"surname_th"`
	TitleTH     string         `gorm:"type:text;not null" json:"title_th"`
	FirstnameEN string         `gorm:"type:text;not null" json:"firstname_en"`
	SurnameEN   string         `gorm:"type:text;not null" json:"surname_en"`
	TitleEN     string         `gorm:"type:text;not null" json:"title_en"`
}

type EventUser struct {
	ID      datatypes.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Role    role           `gorm:"type:role;not null" json:"role"`
	UserID  datatypes.UUID `gorm:"type:uuid;not null" json:"user_id"`
	EventID datatypes.UUID `gorm:"type:uuid;not null" json:"event_id"`

	Event Event `gorm:"foreignKey:EventID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	User  User  `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type EventUserInput struct {
	RefID uint64
	Role  role
}