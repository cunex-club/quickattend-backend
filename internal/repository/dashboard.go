package repository

import (
	"context"
)

type DashboardRepository interface {
	GetEventDashboardData(eventID string, ctx context.Context) error
}

func (r *repository) GetEventDashboardData(eventID string, ctx context.Context) error {
	return nil
}
