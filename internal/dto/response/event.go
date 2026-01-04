package response

import (
	"gorm.io/datatypes"
)

type GetEventsRes struct {
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
