package entity

import (
	"database/sql/driver"

	"gorm.io/datatypes"
)

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

type User struct {
	ID          datatypes.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	RefID       uint64         `gorm:"type:bigint;not null" json:"ref_id"`
	FirstnameTH string         `gorm:"type:text;not null" json:"firstname_th"`
	SurnameTH   string         `gorm:"type:text;not null" json:"surname_th"`
	TitleTH     string         `gorm:"type:text;not null" json:"title_th"`
	FirstnameEN string         `gorm:"type:text;not null" json:"firstname_en"`
	SurnameEN   string         `gorm:"type:text;not null" json:"surname_en"`
	TitleEN     string         `gorm:"type:text;not null" json:"title_en"`
}

type EventUser struct {
	ID      datatypes.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Role    role           `gorm:"type:role;not null" json:"role"`
	UserID  datatypes.UUID `gorm:"type:uuid;not null" json:"user_id"`
	EventID datatypes.UUID `gorm:"type:uuid;not null" json:"event_id"`

	Event Event `gorm:"foreignKey:EventID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	User  User  `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
