package service

import (
	"context"

	gql "github.com/cunex-club/quickattend-backend/internal/infrastructure/http/handler/graphql"
)

type DashboardService interface {
	GetEventDashboardData(eventID string, ctx context.Context) (*gql.DashboardReadyDataDTO, error)
}

func (s *service) GetEventDashboardData(eventID string, ctx context.Context) (*gql.DashboardReadyDataDTO, error) {
	// Fetch event details
	return &gql.DashboardReadyDataDTO{
		Summary: gql.RegistrationSummaryDTO{
			TotalEligible: 100,
			TotalStudent:  70,
			TotalStaff:    30,
			TotalAll:      100,
		},
		FacultyStats: []gql.FacultyStatDTO{
			{
				Organization: "Engineering",
				StudentCount: 50,
				StaffCount:   10,
				TotalCount:   60,
			},
		},
		TimeSeriesStats: []gql.TimeStatDTO{
			{
				TimeBucket:   "16:00",
				StudentCount: 20,
				StaffCount:   5,
				TotalCount:   25,
			},
		},
	}, nil
}
