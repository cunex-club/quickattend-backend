package response

import (
	"time"
)

type GetOneEventAgenda struct {
	ActivityName string    `json:"activity_name"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
}

type GetOneEventRes struct {
	Name            string              `json:"name"`
	Organizer       string              `json:"organizer"`
	Description     *string             `json:"description"`
	StartTime       time.Time           `json:"start_time"`
	EndTime         time.Time           `json:"end_time"`
	Location        string              `json:"location"`
	TotalRegistered uint16              `json:"total_registered"`
	EvaluationForm  *string             `json:"evaluation_form"`
	Role            *string             `json:"role"`
	Agenda          []GetOneEventAgenda `json:"agenda"`
}

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

type CreateEventRes struct {
	ID string `json:"id"`
}

type UpdateEventRes struct {
	ID string `json:"id"`
}
