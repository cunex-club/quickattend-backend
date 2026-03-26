package service

import (
	"context"

	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
	"github.com/google/uuid"
)

type DashboardService interface {
	GetRegistrationSummary(ctx context.Context, eventID uuid.UUID) (*dtoRes.RegistrationSummary, error)
}

func (s *service) GetRegistrationSummary(ctx context.Context, eventID uuid.UUID) (*dtoRes.RegistrationSummary, error) {
	data, err := s.repo.Dashboard.GetRegistrationSummary(ctx, eventID)
	if err != nil {
		return nil, err
	}
	return data, nil
}

