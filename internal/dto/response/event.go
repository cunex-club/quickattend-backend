package response

import "time"

type DuplicateEventRes struct {
	DuplicatedEventId string `json:"event_id"`
}

type status string

const (
	SUCCESS   status = "success"
	DUPLICATE status = "duplicate"
	FAIL      status = "fail"
)

type GetParticipantRes struct {
	FirstnameTH     *string   `json:"firstname_th"`
	SurnameTH       *string   `json:"surname_th"`
	TitleTH         *string   `json:"title_th"`
	FirstnameEN     *string   `json:"firstname_en"`
	SurnameEN       *string   `json:"surname_en"`
	TitleEN         *string   `json:"title_en"`
	RefID           *string   `json:"ref_id"`
	OrganizationTH  *string   `json:"organization_th"`
	OrganizationEN  *string   `json:"organization_en"`
	CheckInTime     time.Time `json:"check_in_time"`
	Status          string    `json:"status"`
	Code            string    `json:"code"`
	ProfileImageUrl *string   `json:"profile_image_url"`
}

type GetOneEventAgenda struct {
	ActivityName string    `json:"activity_name"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
}

type GetOneEventUser struct {
	FirstnameTH string `json:"firstname_th"`
	SurnameTH   string `json:"surname_th"`
	TitleTH     string `json:"title_th"`
	FirstnameEN string `json:"firstname_en"`
	SurnameEN   string `json:"surname_en"`
	TitleEN     string `json:"title_en"`
	Role        string `json:"role"`
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
	AllowAllToScan  bool                `json:"allow_all_to_scan"`
	RevealedFields  []string            `json:"revealed_fields"`
	Role            *string             `json:"role"`
	Agenda          []GetOneEventAgenda `json:"agenda"`
	User            []GetOneEventUser   `json:"users"`
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
