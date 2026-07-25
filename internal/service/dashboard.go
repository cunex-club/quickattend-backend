package service

import (
	"context"
	"errors"
	"fmt"

	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DashboardService interface {
	GetEventDashboardData(ctx context.Context, eventID uuid.UUID, userID uuid.UUID) (*dtoRes.EventDashboard, error)
}

// ErrDashboardForbidden is returned when the caller holds no role in the event
// they are asking the dashboard for.
var ErrDashboardForbidden = errors.New("you do not have access to this event's dashboard")

func (s *service) GetEventDashboardData(ctx context.Context, eventID uuid.UUID, userID uuid.UUID) (*dtoRes.EventDashboard, error) {
	// Authentication alone is not enough: the dashboard exposes per-event
	// attendance data, so the caller must hold a role in *this* event.
	role, err := s.repo.Event.GetUserRoleInEvent(eventID, userID, ctx)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check dashboard permission: %w", err)
	}
	if role == nil {
		return nil, ErrDashboardForbidden
	}

	data, err := s.repo.Dashboard.GetEventDashboardData(ctx, eventID)
	if err != nil {
		return nil, err
	}

	return data, nil
}
