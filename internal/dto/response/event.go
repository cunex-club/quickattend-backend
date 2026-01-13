package response

import "gorm.io/datatypes"

type GetOneEventAgenda struct {
	ActivityName string         `json:"activity_name"`
	StartTime    datatypes.Time `json:"start_time"`
	EndTime      datatypes.Time `json:"end_time"`
}

type GetOneEventRes struct {
	Name            string              `gorm:"column:name" json:"name"`
	Organizer       string              `gorm:"column:organizer" json:"organizer"`
	Description     *string             `gorm:"column:description" json:"description"`
	Date            datatypes.Date      `gorm:"column:date" json:"date"`
	StartTime       datatypes.Time      `gorm:"column:start_time" json:"start_time"`
	EndTime         datatypes.Time      `gorm:"column:end_time" json:"end_time"`
	Location        string              `gorm:"column:location" json:"location"`
	TotalRegistered uint16              `gorm:"column:total_registered" json:"total_registered"`
	EvaluationForm  *string             `gorm:"column:evaluation_form" json:"evaluation_form"`
	Agenda          []GetOneEventAgenda `gorm:"column:agenda" json:"agenda"`
}
