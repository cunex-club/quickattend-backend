package service

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cunex-club/quickattend-backend/internal/entity"
	"github.com/google/uuid"
)

type DashboardService interface {
	GetRegistrationSummary(ctx context.Context, eventID uuid.UUID) (*RegistrationSummaryData, error)
	GetRegistrationAnalytics(ctx context.Context, eventID uuid.UUID, filter RegistrationAnalyticsFilter, focus *RegistrationAnalyticsFocus) (*RegistrationAnalyticsData, error)
	GetWhitelistParticipants(ctx context.Context, eventID uuid.UUID, filter RegistrationAnalyticsFilter, state *ParticipantState) ([]WhitelistParticipantData, error)
}

type RegistrationSummaryData struct {
	TotalEligible int
	TotalStudent  int
	TotalStaff    int
	TotalAll      int
}

type RegistrationAnalyticsFilter struct {
	FacultyIDs  []string
	TimeSlotIDs []string
}

type RegistrationAnalyticsFocus struct {
	Dimension FocusDimension
	ID        string
}

type RegistrationAnalyticsData struct {
	Mode      RegistrationMode
	Summary   SummaryData
	Focus     *FocusData
	Faculties []FacultyStatData
	TimeSlots []TimeSlotStatData
	Matrix    []FacultyTimeSlotStatData
}

type RegistrationMode string

const (
	RegistrationModeOpen      RegistrationMode = "OPEN"
	RegistrationModeWhitelist RegistrationMode = "WHITELIST"
	RegistrationModeFaculties RegistrationMode = "FACULTIES"
)

type FocusDimension string

const (
	FocusDimensionFaculty  FocusDimension = "FACULTY"
	FocusDimensionTimeSlot FocusDimension = "TIME_SLOT"
)

type ParticipantState string

const (
	ParticipantStateRegistered    ParticipantState = "REGISTERED"
	ParticipantStateNotRegistered ParticipantState = "NOT_REGISTERED"
)

type SummaryData struct {
	RegisteredCount   int
	ExpectedCount     *int
	UnregisteredCount *int
}

type FocusData struct {
	Dimension FocusDimension
	ID        string
	Label     string
}

type FacultyData struct {
	ID   string
	Name string
}

type TimeSlotData struct {
	ID      string
	Label   string
	StartAt time.Time
	EndAt   time.Time
}

type FacultyStatData struct {
	Faculty FacultyData
	Summary SummaryData
}

type TimeSlotStatData struct {
	TimeSlot TimeSlotData
	Summary  SummaryData
}

type FacultyTimeSlotStatData struct {
	Faculty  FacultyData
	TimeSlot TimeSlotData
	Summary  SummaryData
}

type WhitelistParticipantData struct {
	ID                 string
	FullName           string
	Faculty            FacultyData
	State              ParticipantState
	RegisteredAt       *time.Time
	RegisteredTimeSlot *TimeSlotData
}

type registeredParticipant struct {
	RefID            uint64
	FacultyCode      int
	TimeSlotID       string
	ScannedTimestamp time.Time
}

const (
	facultyEngineering  = 1
	facultyArts         = 2
	facultyScience      = 3
	facultyPolitical    = 4
	facultyArchitecture = 5
)

var facultyNameByCode = map[int]string{
	facultyEngineering:  "คณะวิศวกรรมศาสตร์",
	facultyArts:         "คณะอักษรศาสตร์",
	facultyScience:      "คณะวิทยาศาสตร์",
	facultyPolitical:    "คณะรัฐศาสตร์",
	facultyArchitecture: "คณะสถาปัตยกรรมศาสตร์",
}

func (s *service) GetRegistrationSummary(ctx context.Context, eventID uuid.UUID) (*RegistrationSummaryData, error) {
	event, err := s.repo.Dashboard.GetDashboardEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	registeredCount, err := s.repo.Dashboard.CountRegisteredParticipants(ctx, eventID)
	if err != nil {
		return nil, err
	}

	totalEligible := 0
	if event.AttendenceType == entity.AttendanceWhitelist {
		totalEligible, err = s.repo.Dashboard.CountExpectedWhitelistParticipants(ctx, eventID)
		if err != nil {
			return nil, err
		}
	}

	return &RegistrationSummaryData{
		TotalEligible: totalEligible,
		TotalStudent:  registeredCount,
		TotalStaff:    0,
		TotalAll:      registeredCount,
	}, nil
}

func (s *service) GetRegistrationAnalytics(ctx context.Context, eventID uuid.UUID, filter RegistrationAnalyticsFilter, focus *RegistrationAnalyticsFocus) (*RegistrationAnalyticsData, error) {
	event, err := s.repo.Dashboard.GetDashboardEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.Dashboard.GetRegisteredParticipantRows(ctx, eventID)
	if err != nil {
		return nil, err
	}

	expectedFacultyCodes, err := s.repo.Dashboard.GetExpectedFacultyCodes(ctx, eventID, event.AttendenceType)
	if err != nil {
		return nil, err
	}

	slots := buildHourlySlots(event.StartTime, event.EndTime)

	registered := make([]registeredParticipant, 0, len(rows))
	for _, row := range rows {
		slot := timeSlotForTime(row.ScannedTimestamp)
		slots[slot.ID] = slot

		registered = append(registered, registeredParticipant{
			RefID:            row.RefID,
			FacultyCode:      facultyCodeFromRefID(row.RefID),
			TimeSlotID:       slot.ID,
			ScannedTimestamp: row.ScannedTimestamp,
		})
	}

	facultySet, hasFacultyFilter := buildFacultyFilterSet(filter.FacultyIDs)
	timeSlotSet, hasTimeSlotFilter := buildTimeSlotFilterSet(filter.TimeSlotIDs, slots)

	filtered := filterRegisteredParticipants(
		registered,
		facultySet,
		hasFacultyFilter,
		timeSlotSet,
		hasTimeSlotFilter,
	)

	displayFacultyCodes := resolveDisplayFacultyCodes(
		expectedFacultyCodes,
		registered,
		facultySet,
		hasFacultyFilter,
	)

	displayTimeSlots := resolveDisplayTimeSlots(
		slots,
		filtered,
		timeSlotSet,
		hasTimeSlotFilter,
	)

	expectedCount, unregisteredCount, err := s.buildAnalyticsExpectedSummary(
		ctx,
		eventID,
		event.AttendenceType,
		facultySet,
		hasFacultyFilter,
		hasTimeSlotFilter,
		len(filtered),
	)
	if err != nil {
		return nil, err
	}

	var focusData *FocusData
	if focus != nil {
		focusData = resolveFocus(*focus, displayTimeSlots)
	}

	return &RegistrationAnalyticsData{
		Mode: mapRegistrationMode(event.AttendenceType),
		Summary: SummaryData{
			RegisteredCount:   len(filtered),
			ExpectedCount:     expectedCount,
			UnregisteredCount: unregisteredCount,
		},
		Focus:     focusData,
		Faculties: buildFacultyStats(displayFacultyCodes, filtered),
		TimeSlots: buildTimeSlotStats(displayTimeSlots, filtered),
		Matrix:    buildMatrix(displayFacultyCodes, displayTimeSlots, filtered),
	}, nil
}

func (s *service) GetWhitelistParticipants(ctx context.Context, eventID uuid.UUID, filter RegistrationAnalyticsFilter, state *ParticipantState) ([]WhitelistParticipantData, error) {
	event, err := s.repo.Dashboard.GetDashboardEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.Dashboard.GetWhitelistParticipantRows(ctx, eventID)
	if err != nil {
		return nil, err
	}

	slots := buildHourlySlots(event.StartTime, event.EndTime)
	facultySet, hasFacultyFilter := buildFacultyFilterSet(filter.FacultyIDs)
	timeSlotSet, hasTimeSlotFilter := buildTimeSlotFilterSet(filter.TimeSlotIDs, slots)

	out := make([]WhitelistParticipantData, 0, len(rows))

	for _, row := range rows {
		facultyCode := facultyCodeFromRefID(row.RefID)
		if hasFacultyFilter && !containsInt(facultySet, facultyCode) {
			continue
		}

		currentState := ParticipantStateNotRegistered
		var registeredSlot *TimeSlotData

		if row.RegisteredAt != nil {
			slot := timeSlotForTime(*row.RegisteredAt)
			slots[slot.ID] = slot

			registeredSlot = &slot
			currentState = ParticipantStateRegistered
		}

		if hasTimeSlotFilter {
			if registeredSlot == nil || !containsString(timeSlotSet, registeredSlot.ID) {
				continue
			}
		}

		if state != nil && currentState != *state {
			continue
		}

		out = append(out, WhitelistParticipantData{
			ID:                 row.ID,
			FullName:           row.FullName,
			Faculty:            facultyFromCode(facultyCode),
			State:              currentState,
			RegisteredAt:       row.RegisteredAt,
			RegisteredTimeSlot: registeredSlot,
		})
	}

	return out, nil
}

func (s *service) buildAnalyticsExpectedSummary(
	ctx context.Context,
	eventID uuid.UUID,
	mode entity.AttendanceType,
	facultySet map[int]struct{},
	hasFacultyFilter bool,
	hasTimeSlotFilter bool,
	registeredCount int,
) (*int, *int, error) {
	if mode != entity.AttendanceWhitelist {
		return nil, nil, nil
	}

	if hasTimeSlotFilter {
		return nil, nil, nil
	}

	expectedCount, err := s.countExpectedWhitelistInScope(ctx, eventID, facultySet, hasFacultyFilter)
	if err != nil {
		return nil, nil, err
	}

	unregisteredCount := expectedCount - registeredCount
	if unregisteredCount < 0 {
		unregisteredCount = 0
	}

	return intPtr(expectedCount), intPtr(unregisteredCount), nil
}

func (s *service) countExpectedWhitelistInScope(ctx context.Context, eventID uuid.UUID, facultySet map[int]struct{}, hasFacultyFilter bool) (int, error) {
	if !hasFacultyFilter {
		return s.repo.Dashboard.CountExpectedWhitelistParticipants(ctx, eventID)
	}

	rows, err := s.repo.Dashboard.GetWhitelistParticipantRows(ctx, eventID)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, row := range rows {
		facultyCode := facultyCodeFromRefID(row.RefID)
		if containsInt(facultySet, facultyCode) {
			count++
		}
	}

	return count, nil
}

func filterRegisteredParticipants(
	rows []registeredParticipant,
	facultySet map[int]struct{},
	hasFacultyFilter bool,
	timeSlotSet map[string]struct{},
	hasTimeSlotFilter bool,
) []registeredParticipant {
	filtered := make([]registeredParticipant, 0, len(rows))

	for _, row := range rows {
		if hasFacultyFilter && !containsInt(facultySet, row.FacultyCode) {
			continue
		}

		if hasTimeSlotFilter && !containsString(timeSlotSet, row.TimeSlotID) {
			continue
		}

		filtered = append(filtered, row)
	}

	return filtered
}

func mapRegistrationMode(mode entity.AttendanceType) RegistrationMode {
	switch mode {
	case entity.AttendanceWhitelist:
		return RegistrationModeWhitelist
	case entity.AttendanceFaculties:
		return RegistrationModeFaculties
	default:
		return RegistrationModeOpen
	}
}

func facultyCodeFromRefID(refID uint64) int {
	return int(refID % 100)
}

func facultyFromCode(code int) FacultyData {
	id := strconv.Itoa(code)

	name, ok := facultyNameByCode[code]
	if !ok {
		name = fmt.Sprintf("คณะ/หน่วยงาน %02d", code)
	}

	return FacultyData{
		ID:   id,
		Name: name,
	}
}

func buildFacultyFilterSet(ids []string) (map[int]struct{}, bool) {
	out := map[int]struct{}{}

	for _, id := range ids {
		cleaned := strings.TrimSpace(id)
		if cleaned == "" || isFilterAll(cleaned) {
			continue
		}

		value, err := strconv.Atoi(cleaned)
		if err != nil {
			continue
		}

		out[value] = struct{}{}
	}

	return out, len(out) > 0
}

func buildTimeSlotFilterSet(ids []string, slots map[string]TimeSlotData) (map[string]struct{}, bool) {
	out := map[string]struct{}{}

	for _, id := range ids {
		cleaned := strings.TrimSpace(id)
		if cleaned == "" || isFilterAll(cleaned) {
			continue
		}

		if slot, ok := slots[cleaned]; ok {
			out[slot.ID] = struct{}{}
			continue
		}

		if parsed, err := time.Parse(time.RFC3339, cleaned); err == nil {
			slot := timeSlotForTime(parsed)
			out[slot.ID] = struct{}{}
			continue
		}

		for _, slot := range slots {
			if legacyTimeSlotID(slot.StartAt) == cleaned || slot.Label == cleaned {
				out[slot.ID] = struct{}{}
				break
			}
		}
	}

	return out, len(out) > 0
}

func isFilterAll(s string) bool {
	upper := strings.ToUpper(strings.TrimSpace(s))
	return upper == "ALL" || upper == "ทั้งหมด"
}

func resolveDisplayFacultyCodes(expected []int, rows []registeredParticipant, filter map[int]struct{}, hasFilter bool) []int {
	set := map[int]struct{}{}

	if hasFilter {
		for code := range filter {
			set[code] = struct{}{}
		}
	} else {
		for _, code := range expected {
			set[code] = struct{}{}
		}

		for _, row := range rows {
			set[row.FacultyCode] = struct{}{}
		}
	}

	out := make([]int, 0, len(set))
	for code := range set {
		out = append(out, code)
	}

	sort.Ints(out)
	return out
}

func resolveDisplayTimeSlots(
	slots map[string]TimeSlotData,
	rows []registeredParticipant,
	filter map[string]struct{},
	hasFilter bool,
) []TimeSlotData {
	out := make([]TimeSlotData, 0)

	if hasFilter {
		for id := range filter {
			if slot, ok := slots[id]; ok {
				out = append(out, slot)
			}
		}
	} else {
		usedSlotIDs := map[string]struct{}{}

		for _, row := range rows {
			usedSlotIDs[row.TimeSlotID] = struct{}{}
		}

		for id := range usedSlotIDs {
			if slot, ok := slots[id]; ok {
				out = append(out, slot)
			}
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].StartAt.Before(out[j].StartAt)
	})

	return out
}

func buildFacultyStats(codes []int, rows []registeredParticipant) []FacultyStatData {
	countByFaculty := map[int]int{}

	for _, row := range rows {
		countByFaculty[row.FacultyCode]++
	}

	out := make([]FacultyStatData, 0, len(codes))
	for _, code := range codes {
		out = append(out, FacultyStatData{
			Faculty: facultyFromCode(code),
			Summary: SummaryData{
				RegisteredCount: countByFaculty[code],
			},
		})
	}

	return out
}

func buildTimeSlotStats(slots []TimeSlotData, rows []registeredParticipant) []TimeSlotStatData {
	countBySlot := map[string]int{}

	for _, row := range rows {
		countBySlot[row.TimeSlotID]++
	}

	out := make([]TimeSlotStatData, 0, len(slots))
	for _, slot := range slots {
		out = append(out, TimeSlotStatData{
			TimeSlot: slot,
			Summary: SummaryData{
				RegisteredCount: countBySlot[slot.ID],
			},
		})
	}

	return out
}

func buildMatrix(codes []int, slots []TimeSlotData, rows []registeredParticipant) []FacultyTimeSlotStatData {
	count := map[int]map[string]int{}

	for _, row := range rows {
		if _, ok := count[row.FacultyCode]; !ok {
			count[row.FacultyCode] = map[string]int{}
		}

		count[row.FacultyCode][row.TimeSlotID]++
	}

	slotByID := make(map[string]TimeSlotData, len(slots))
	for _, slot := range slots {
		slotByID[slot.ID] = slot
	}

	out := make([]FacultyTimeSlotStatData, 0)

	for _, code := range codes {
		bySlot, ok := count[code]
		if !ok {
			continue
		}

		for _, slot := range slots {
			value := bySlot[slot.ID]
			if value == 0 {
				continue
			}

			out = append(out, FacultyTimeSlotStatData{
				Faculty:  facultyFromCode(code),
				TimeSlot: slotByID[slot.ID],
				Summary: SummaryData{
					RegisteredCount: value,
				},
			})
		}
	}

	return out
}

func resolveFocus(focus RegistrationAnalyticsFocus, slots []TimeSlotData) *FocusData {
	switch focus.Dimension {
	case FocusDimensionFaculty:
		code, err := strconv.Atoi(strings.TrimSpace(focus.ID))
		if err != nil {
			return nil
		}

		return &FocusData{
			Dimension: FocusDimensionFaculty,
			ID:        strconv.Itoa(code),
			Label:     facultyFromCode(code).Name,
		}

	case FocusDimensionTimeSlot:
		for _, slot := range slots {
			if slot.ID == focus.ID || legacyTimeSlotID(slot.StartAt) == focus.ID || slot.Label == focus.ID {
				return &FocusData{
					Dimension: FocusDimensionTimeSlot,
					ID:        slot.ID,
					Label:     slot.Label,
				}
			}
		}
	}

	return nil
}

func buildHourlySlots(startUTC time.Time, endUTC time.Time) map[string]TimeSlotData {
	out := map[string]TimeSlotData{}

	start := startUTC.In(thaiLoc).Truncate(time.Hour)
	end := ceilToHour(endUTC.In(thaiLoc))

	if !end.After(start) {
		end = start.Add(time.Hour)
	}

	for cursor := start; cursor.Before(end); cursor = cursor.Add(time.Hour) {
		slot := buildTimeSlot(cursor)
		out[slot.ID] = slot
	}

	return out
}

func timeSlotForTime(t time.Time) TimeSlotData {
	return buildTimeSlot(t.In(thaiLoc).Truncate(time.Hour))
}

func buildTimeSlot(start time.Time) TimeSlotData {
	end := start.Add(time.Hour)

	return TimeSlotData{
		ID:      start.Format(time.RFC3339),
		Label:   fmt.Sprintf("%02d:00 - %02d:00 น.", start.Hour(), end.Hour()),
		StartAt: start,
		EndAt:   end,
	}
}

func legacyTimeSlotID(start time.Time) string {
	end := start.Add(time.Hour)
	return fmt.Sprintf("%02d:00-%02d:00", start.Hour(), end.Hour())
}

func ceilToHour(t time.Time) time.Time {
	truncated := t.Truncate(time.Hour)
	if truncated.Equal(t) {
		return truncated
	}

	return truncated.Add(time.Hour)
}

func containsInt(set map[int]struct{}, value int) bool {
	_, ok := set[value]
	return ok
}

func containsString(set map[string]struct{}, value string) bool {
	_, ok := set[value]
	return ok
}

func intPtr(v int) *int {
	return &v
}
