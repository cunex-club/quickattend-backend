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
	Agenda          []GetOneEventAgenda `json:"agenda"`
}
