package entity

import (
	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	RefID       int8      `gorm:"not null" json:"ref_id"`
	FirstnameTH string    `gorm:"type:text;not null" json:"firstname_th"`
	SurnameTH   string    `gorm:"type:text;not null" json:"surname_th"`
	TitleTH     string    `gorm:"type:text;not null" json:"title_th"`
	FirstnameEN string    `gorm:"type:text;not null" json:"firstname_en"`
	SurnameEN   string    `gorm:"type:text;not null" json:"surname_en"`
	TitleEN     string    `gorm:"type:text;not null" json:"title_en"`
}
