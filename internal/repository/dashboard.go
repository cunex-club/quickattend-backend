package repository

import (
	"context"

	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
	"github.com/cunex-club/quickattend-backend/internal/entity"
	"github.com/google/uuid"
)

type DashboardRepository interface {
	GetRegistrationSummary(ctx context.Context, eventID uuid.UUID) (*dtoRes.RegistrationSummary, error)
}

func (r *repository) GetRegistrationSummary(ctx context.Context, eventID uuid.UUID) (*dtoRes.RegistrationSummary, error) {
	row, err := r.getParticipantSummary(ctx, eventID)
	if err != nil {
		return nil, err
	}

	return &dtoRes.RegistrationSummary{
		TotalEligible: 0,
		TotalStudent:  row.TotalStudent,
		TotalStaff:    row.TotalStaff,
		TotalAll:      row.TotalAll,
	}, nil
}

// helpers function for dashboard
func (r *repository) getParticipantSummary(ctx context.Context, eventID uuid.UUID) (*entity.RegistrationSummary, error) {
	var row entity.RegistrationSummary

	err := r.db.WithContext(ctx).Raw(`
		SELECT
			SUM(
				CASE
					WHEN CHAR_LENGTH(CAST(u.ref_id AS TEXT)) = 10 THEN 1
					ELSE 0
				END
			) AS total_student,
			SUM(
				CASE
					WHEN CHAR_LENGTH(CAST(u.ref_id AS TEXT)) < 10 THEN 1
					ELSE 0
				END
			) AS total_staff,
			COUNT(*) AS total_all
		FROM event_participants ep
		JOIN users u ON u.id = ep.participant_id
		WHERE ep.event_id = ?
	`, eventID).Scan(&row).Error

	if err != nil {
		return nil, err
	}

	return &row, nil
}
