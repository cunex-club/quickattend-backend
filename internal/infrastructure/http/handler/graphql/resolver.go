package graphql

import "context"

type DashboardReader interface {
	GetEventDashboardData(ctx context.Context, eventID string) (*DashboardReadyDataDTO, error)
}

type Resolver struct {
	DashboardReader DashboardReader
}

type DashboardReadyDataDTO struct {
	Summary         RegistrationSummaryDTO
	FacultyStats    []FacultyStatDTO
	TimeSeriesStats []TimeStatDTO
}

type RegistrationSummaryDTO struct {
	TotalEligible int
	TotalStudent  int
	TotalStaff    int
	TotalAll      int
}

type FacultyStatDTO struct {
	Organization string
	StudentCount int
	StaffCount   int
	TotalCount   int
}

type TimeStatDTO struct {
	TimeBucket   string
	StudentCount int
	StaffCount   int
	TotalCount   int
}