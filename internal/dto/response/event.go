package response

import (
	"gorm.io/datatypes"
)

type RawGetEventsIndividualEvent struct {
	ID             datatypes.UUID `gorm:"column:id"`
	Name           string         `gorm:"column:name"`
	Organizer      string         `gorm:"column:organizer"`
	Description    *string        `gorm:"column:description"`
	Date           datatypes.Date `gorm:"column:date"`
	StartTime      datatypes.Time `gorm:"column:start_time"`
	EndTime        datatypes.Time `gorm:"column:end_time"`
	Location       string         `gorm:"column:location"`
	Role           *string        `gorm:"column:role"`
	EvaluationForm *string        `gorm:"column:evaluation_form"`
}

type GetEventsIndividualEvent struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Organizer      string         `json:"organizer"`
	Description    *string        `json:"description"`
	Date           datatypes.Date `json:"date"`
	StartTime      datatypes.Time `json:"start_time"`
	EndTime        datatypes.Time `json:"end_time"`
	Location       string         `json:"location"`
	Role           *string        `json:"role,omitempty"`
	EvaluationForm *string        `json:"evaluation_form"`
}
