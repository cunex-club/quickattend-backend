package request

type CreateAgendaReq struct {
	ActivityName string `json:"activity_name" validate:"required"`
	StartTime    string `json:"start_time" validate:"required"` // RFC3339 UTC
	EndTime      string `json:"end_time" validate:"required"`   // RFC3339 UTC
}

type ManagerStaffReq struct {
	RefID uint64 `json:"ref_id" validate:"required"`
	Role  string `json:"role" validate:"required,oneof=owner staff manager"`
}

type CreateEventReq struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Organizer   string `json:"organizer" validate:"required"`

	StartTime string `json:"start_time" validate:"required"` // RFC3339 UTC
	EndTime   string `json:"end_time" validate:"required"`   // RFC3339 UTC
	Timezone  string `json:"timezone" validate:"required"`   // e.g. Asia/Bangkok

	Location string            `json:"location" validate:"required"`
	Agenda   []CreateAgendaReq `json:"agenda" validate:"dive"`

	AttendanceType string   `json:"attendance_type" validate:"required,oneof=all whitelist faculties"`
	Attendee       []any    `json:"attendee" validate:"required"`
	RevealedFields []string `json:"revealed_fields" validate:"required,min=1,dive,oneof=name organization refid photo"`

	ManagersAndStaff []ManagerStaffReq `json:"managers_and_staff" validate:"dive"`

	AllowAllToScan *bool  `json:"allow_all_to_scan"`
	EvaluationForm string `json:"evaluation_form"`
}

type UpdateEventReq = CreateEventReq
