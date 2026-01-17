package response

import (
	"time"
)

type GetEventsRes struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Organizer      string    `json:"organizer"`
	Description    *string   `json:"description"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Location       string    `json:"location"`
	Role           *string   `json:"role,omitempty"`
	EvaluationForm *string   `json:"evaluation_form"`
}
