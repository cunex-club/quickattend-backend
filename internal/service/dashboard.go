package service

import (
	"context"

	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
	"github.com/google/uuid"
)

type DashboardService interface {
	GetEventDashboardData(ctx context.Context, eventID uuid.UUID) (*dtoRes.DashboardReadyData, error)
}

func (s *service) GetEventDashboardData(ctx context.Context, eventID uuid.UUID) (*dtoRes.DashboardReadyData, error) {
	data, err := s.repo.Dashboard.GetEventDashboardData(ctx, eventID)
	if err != nil {
		return nil, err
	}

	return data, nil
}
