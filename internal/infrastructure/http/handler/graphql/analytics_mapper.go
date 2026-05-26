package graphql

import (
	"time"

	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/handler/graphql/model"
	"github.com/cunex-club/quickattend-backend/internal/service"
)

func toServiceRegistrationAnalyticsFilter(filter *model.RegistrationAnalyticsFilterInput) service.RegistrationAnalyticsFilter {
	if filter == nil {
		return service.RegistrationAnalyticsFilter{}
	}

	return service.RegistrationAnalyticsFilter{
		FacultyIDs:  append([]string{}, filter.FacultyIds...),
		TimeSlotIDs: append([]string{}, filter.TimeSlotIds...),
	}
}

func toServiceRegistrationAnalyticsFocus(focus *model.RegistrationAnalyticsFocusInput) *service.RegistrationAnalyticsFocus {
	if focus == nil {
		return nil
	}

	return &service.RegistrationAnalyticsFocus{
		Dimension: service.FocusDimension(focus.Dimension),
		ID:        focus.ID,
	}
}

func toServiceParticipantState(state *model.ParticipantState) *service.ParticipantState {
	if state == nil {
		return nil
	}

	v := service.ParticipantState(*state)
	return &v
}

func toModelRegistrationAnalytics(data *service.RegistrationAnalyticsData) *model.RegistrationAnalytics {
	faculties := make([]*model.FacultyStat, 0, len(data.Faculties))
	for _, item := range data.Faculties {
		faculties = append(faculties, toModelFacultyStat(item))
	}

	timeSlots := make([]*model.TimeSlotStat, 0, len(data.TimeSlots))
	for _, item := range data.TimeSlots {
		timeSlots = append(timeSlots, toModelTimeSlotStat(item))
	}

	matrix := make([]*model.FacultyTimeSlotStat, 0, len(data.Matrix))
	for _, item := range data.Matrix {
		matrix = append(matrix, toModelFacultyTimeSlotStat(item))
	}

	return &model.RegistrationAnalytics{
		Mode:      model.RegistrationMode(data.Mode),
		Summary:   toModelSummary(data.Summary),
		Focus:     toModelFocus(data.Focus),
		Faculties: faculties,
		TimeSlots: timeSlots,
		Matrix:    matrix,
	}
}

func toModelSummary(data service.SummaryData) *model.Summary {
	return &model.Summary{
		RegisteredCount:   data.RegisteredCount,
		ExpectedCount:     data.ExpectedCount,
		UnregisteredCount: data.UnregisteredCount,
	}
}

func toModelFocus(data *service.FocusData) *model.Focus {
	if data == nil {
		return nil
	}

	return &model.Focus{
		Dimension: model.FocusDimension(data.Dimension),
		ID:        data.ID,
		Label:     data.Label,
	}
}

func toModelFaculty(data service.FacultyData) *model.Faculty {
	return &model.Faculty{
		ID:   data.ID,
		Name: data.Name,
	}
}

func toModelTimeSlot(data service.TimeSlotData) *model.TimeSlot {
	return &model.TimeSlot{
		ID:      data.ID,
		Label:   data.Label,
		StartAt: formatDateTime(data.StartAt),
		EndAt:   formatDateTime(data.EndAt),
	}
}

func toModelFacultyStat(data service.FacultyStatData) *model.FacultyStat {
	return &model.FacultyStat{
		Faculty: toModelFaculty(data.Faculty),
		Summary: toModelSummary(data.Summary),
	}
}

func toModelTimeSlotStat(data service.TimeSlotStatData) *model.TimeSlotStat {
	return &model.TimeSlotStat{
		TimeSlot: toModelTimeSlot(data.TimeSlot),
		Summary:  toModelSummary(data.Summary),
	}
}

func toModelFacultyTimeSlotStat(data service.FacultyTimeSlotStatData) *model.FacultyTimeSlotStat {
	return &model.FacultyTimeSlotStat{
		Faculty:  toModelFaculty(data.Faculty),
		TimeSlot: toModelTimeSlot(data.TimeSlot),
		Summary:  toModelSummary(data.Summary),
	}
}

func toModelWhitelistParticipant(data service.WhitelistParticipantData) *model.WhitelistParticipant {
	var registeredTimeSlot *model.TimeSlot
	if data.RegisteredTimeSlot != nil {
		registeredTimeSlot = toModelTimeSlot(*data.RegisteredTimeSlot)
	}

	return &model.WhitelistParticipant{
		ID:                data.ID,
		FullName:          data.FullName,
		Faculty:           toModelFaculty(data.Faculty),
		State:             model.ParticipantState(data.State),
		RegisteredAt:       formatDateTimePtr(data.RegisteredAt),
		RegisteredTimeSlot: registeredTimeSlot,
	}
}

func formatDateTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func formatDateTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}

	formatted := t.Format(time.RFC3339)
	return &formatted
}