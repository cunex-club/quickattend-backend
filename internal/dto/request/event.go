package request

type CreateAgendaReq struct {
	ActivityName string `json:"activity_name" validate:"required"`
	StartTime    string `json:"start_time" validate:"required"` // RFC3339 UTC
	EndTime      string `json:"end_time" validate:"required"`   // RFC3339 UTC
}

type ManagerStaffReq struct {
	RefID uint64 `json:"ref_id" validate:"required"`
	Role  string `json:"role" validate:"required,oneof=OWNER STAFF MANAGER"`
}

type CreateEventReq struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Organizer   string `json:"organizer" validate:"required"`

	StartTime string `json:"start_time" validate:"required"` // RFC3339 UTC
	EndTime   string `json:"end_time" validate:"required"`   // RFC3339 UTC
	Timezone  string `json:"timezone" validate:"required"`   // e.g. Asia/Bangkok

	Location     string            `json:"location" validate:"required"`
	LocationLat  float64           `json:"location_lat"`
	LocationLong float64           `json:"location_long"`
	Agenda       []CreateAgendaReq `json:"agenda" validate:"dive"`

	AttendanceType string   `json:"attendance_type" validate:"required,oneof=ALL WHITELIST FACULTIES"`
	Attendee       []any    `json:"attendee" validate:"required"`
	RevealedFields []string `json:"revealed_fields" validate:"required,min=1,dive,oneof=NAME ORGANIZATION REFID PHOTO"`

	ManagersAndStaff []ManagerStaffReq `json:"managers_and_staff" validate:"dive"`

	AllowAllToScan *bool  `json:"allow_all_to_scan"`
	EvaluationForm string `json:"evaluation_form"`
}

type UpdateEventReq = CreateEventReq

type DuplicateEventReq struct {
	Location       string            `json:"location" validate:"required"`
	LocationLat    float64           `json:"location_lat"`
	LocationLong   float64           `json:"location_long"`
	StartTime      string            `json:"start_time" validate:"required"` // RFC3339 UTC
	EndTime        string            `json:"end_time" validate:"required"`   // RFC3339 UTC
	Timezone       string            `json:"timezone" validate:"required"`   // e.g. Asia/Bangkok
	Agenda         []CreateAgendaReq `json:"agenda" validate:"dive"`
	EvaluationForm *string           `json:"evaluation_form"`
}

type CommentReq struct {
	Comment            string `json:"comment"`
	EncodedOneTimeCode string `json:"one_time_code"`
}

type PostParticipantReqBody struct {
	EventId          string  `json:"event_id"`
	ScannedLocationX float64 `json:"scanned_location_long"`
	ScannedLocationY float64 `json:"scanned_location_lat"`
}

