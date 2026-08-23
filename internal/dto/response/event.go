package response

import (
	"time"

	"github.com/cunex-club/quickattend-backend/internal/entity"
)

type DuplicateEventRes struct {
	DuplicatedEventId string `json:"event_id"`
}

type status string

const (
	SUCCESS   status = "success"
	DUPLICATE status = "duplicate"
	FAIL      status = "fail"
)

type PostParticipantRes struct {
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

type RecentParticipantRes struct {
	RefID       uint64    `json:"ref_id"`
	TitleTH     *string   `json:"title_th"`
	FirstnameTH *string   `json:"firstname_th"`
	SurnameTH   *string   `json:"surname_th"`
	TitleEN     *string   `json:"title_en"`
	FirstnameEN *string   `json:"firstname_en"`
	SurnameEN   *string   `json:"surname_en"`
	CheckInTime time.Time `json:"check_in_time"`
}

type GetOneEventAgenda struct {
	ActivityName string    `json:"activity_name"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
}

type GetOneEventUser struct {
	RefID         string  `json:"ref_id"`
	UserType      string  `json:"user_type"`
	FirstnameTH   *string `json:"firstname_th"`
	SurnameTH     *string `json:"surname_th"`
	TitleTH       *string `json:"title_th"`
	FacultyNameTH *string `json:"faculty_name_th"`
	FirstnameEN   *string `json:"firstname_en"`
	SurnameEN     *string `json:"surname_en"`
	TitleEN       *string `json:"title_en"`
	FacultyNameEN *string `json:"faculty_name_en"`
	Role          string  `json:"role"`
}

type GetOneEventUserPending struct {
	RefID string `json:"ref_id"`
	Role  string `json:"role"`
}

type GetOneEventAllowedFaculties struct {
	FacultyNO uint8 `json:"faculty_no"`
}

type GetOneEventWhitelist struct {
	RefID         string  `json:"ref_id"`
	UserType      string  `json:"user_type"`
	FirstnameTH   *string `json:"firstname_th"`
	SurnameTH     *string `json:"surname_th"`
	TitleTH       *string `json:"title_th"`
	FacultyNameTH *string `json:"faculty_name_th"`
	FirstnameEN   *string `json:"firstname_en"`
	SurnameEN     *string `json:"surname_en"`
	TitleEN       *string `json:"title_en"`
	FacultyNameEN *string `json:"faculty_name_en"`
}

type GetOneEventWhitelistPending struct {
	RefID string `json:"ref_id"`
}

type GetOneEventRes struct {
	Name             string                        `json:"name"`
	Organizer        string                        `json:"organizer"`
	Description      *string                       `json:"description"`
	StartTime        time.Time                     `json:"start_time"`
	EndTime          time.Time                     `json:"end_time"`
	Location         string                        `json:"location"`
	LocationLat      float64                       `json:"location_lat"`
	LocationLong     float64                       `json:"location_long"`
	TotalRegistered  uint16                        `json:"total_registered"`
	EvaluationForm   *string                       `json:"evaluation_form"`
	AllowAllToScan   bool                          `json:"allow_all_to_scan"`
	AttendanceType   entity.AttendanceType         `json:"attendance_type"`
	RevealedFields   []string                      `json:"revealed_fields"`
	Role             *string                       `json:"role"`
	Agenda           []GetOneEventAgenda           `json:"agenda"`
	User             []GetOneEventUser             `json:"users"`
	UserPending      []GetOneEventUserPending      `json:"users_pending"`
	AllowedFaculties []GetOneEventAllowedFaculties `json:"allowed_faculties"`
	WhiteList        []GetOneEventWhitelist        `json:"whitelist"`
	WhiteListPending []GetOneEventWhitelistPending `json:"whitelist_pending"`
}

type GetEventsRes struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Organizer      string    `json:"organizer"`
	Description    *string   `json:"description"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Location       string    `json:"location"`
	Role           *string   `json:"role"`
	EvaluationForm *string   `json:"evaluation_form"`
}

type CreateEventRes struct {
	ID string `json:"id"`
}

type UpdateEventRes struct {
	ID string `json:"id"`
}

type GetDiscoveryEventsRes struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Organizer      string    `json:"organizer"`
	Description    *string   `json:"description"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Location       string    `json:"location"`
	EvaluationForm *string   `json:"evaluation_form"`
	LocationLat    float64   `json:"location_lat"`
	LocationLong   float64   `json:"location_long"`
	AllowAllToScan bool      `json:"allow_all_to_scan"`
}
