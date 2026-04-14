package repository

import (
	"context"

	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
	"github.com/cunex-club/quickattend-backend/internal/entity"
	"github.com/google/uuid"
)



type DashboardRepository interface {
	GetRegistrationSummary(ctx context.Context, eventID uuid.UUID) (*dtoRes.RegistrationSummary, error)
	GetEventDashboardData(ctx context.Context, eventID uuid.UUID) (*dtoRes.DashboardReadyData, error)
}

func (r *repository) GetRegistrationSummary(ctx context.Context, eventID uuid.UUID) (*dtoRes.RegistrationSummary, error) {
	row, err := r.getParticipantSummary(ctx, eventID)
	if err != nil {
		return nil, err
	}

	totalEligible, err := r.getEligibleCount(ctx, eventID)
	if err != nil {
		return nil, err
	}

	return &dtoRes.RegistrationSummary{
		TotalEligible: totalEligible,
		TotalStudent:  row.TotalStudent,
		TotalStaff:    row.TotalStaff,
		TotalAll:      row.TotalAll,
	}, nil
}

func (r *repository) GetEventDashboardData(ctx context.Context, eventID uuid.UUID) (*dtoRes.DashboardReadyData, error) {
	summary, err := r.GetRegistrationSummary(ctx, eventID)
	if err != nil {
		return nil, err
	}

	facultyRows, err := r.getFacultyStats(ctx, eventID)
	if err != nil {
		return nil, err
	}

	timeRows, err := r.getTimeStats(ctx, eventID)
	if err != nil {
		return nil, err
	}

	facultyStats := make([]dtoRes.FacultyStat, 0, len(facultyRows))
	for _, row := range facultyRows {
		facultyStats = append(facultyStats, dtoRes.FacultyStat{
			Organization: row.Organization,
			StudentCount: row.StudentCount,
			StaffCount:   row.StaffCount,
			TotalCount:   row.TotalCount,
		})
	}

	timeStats := make([]dtoRes.TimeStat, 0, len(timeRows))
	for _, row := range timeRows {
		timeStats = append(timeStats, dtoRes.TimeStat{
			TimeBucket:   row.TimeBucket,
			StudentCount: row.StudentCount,
			StaffCount:   row.StaffCount,
			TotalCount:   row.TotalCount,
		})
	}

	return &dtoRes.DashboardReadyData{
		Summary:         *summary,
		FacultyStats:    facultyStats,
		TimeSeriesStats: timeStats,
	}, nil
}


// Currently for WHITELIST typed event:
// EligibleCount = 
// the union of current whitelist rows, pending whitelist rows, and already scanned participant
// this covers the case where an organizer removes a user from the whitelist after they've attended.
// For other attendance type, returns nil.
func (r *repository) getEligibleCount(ctx context.Context, eventID uuid.UUID) (*int, error) {
	var attendanceType entity.AttendanceType
	err := r.db.WithContext(ctx).
		Model(&entity.Event{}).
		Select("attendence_type").
		Where("id = ?", eventID).
		Scan(&attendanceType).Error
	if err != nil {
		return nil, err
	}

	if attendanceType != entity.AttendanceWhitelist {
		return nil, nil
	}

	var count int
	err = r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM (
			SELECT attendee_ref_id FROM event_whitelists WHERE event_id = ?
			UNION
			SELECT attendee_ref_id FROM event_whitelist_pendings WHERE event_id = ?
			UNION
			SELECT u.ref_id FROM event_participants ep
			JOIN users u ON u.id = ep.participant_id
			WHERE ep.event_id = ?
		) w
	`, eventID, eventID, eventID).Scan(&count).Error
	if err != nil {
		return nil, err
	}

	return &count, nil
}

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

func (r *repository) getFacultyStats(ctx context.Context, eventID uuid.UUID) ([]entity.FacultyStat, error) {
	var rows []entity.FacultyStat

	err := r.db.WithContext(ctx).Raw(`
		SELECT
			ep.organization,
			SUM(
				CASE
					WHEN CHAR_LENGTH(CAST(u.ref_id AS TEXT)) = 10 THEN 1
					ELSE 0
				END
			) AS student_count,
			SUM(
				CASE
					WHEN CHAR_LENGTH(CAST(u.ref_id AS TEXT)) < 10 THEN 1
					ELSE 0
				END
			) AS staff_count,
			COUNT(*) AS total_count
		FROM event_participants ep
		JOIN users u ON u.id = ep.participant_id
		WHERE ep.event_id = ?
		GROUP BY ep.organization
		ORDER BY total_count DESC, ep.organization ASC
	`, eventID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *repository) getTimeStats(ctx context.Context, eventID uuid.UUID) ([]entity.TimeStat, error) {
	var rows []entity.TimeStat

	err := r.db.WithContext(ctx).Raw(`
		SELECT
			TO_CHAR(DATE_TRUNC('hour', ep.scanned_timestamp AT TIME ZONE 'Asia/Bangkok'), 'HH24:00') AS time_bucket,
			SUM(
				CASE
					WHEN CHAR_LENGTH(CAST(u.ref_id AS TEXT)) = 10 THEN 1
					ELSE 0
				END
			) AS student_count,
			SUM(
				CASE
					WHEN CHAR_LENGTH(CAST(u.ref_id AS TEXT)) < 10 THEN 1
					ELSE 0
				END
			) AS staff_count,
			COUNT(*) AS total_count
		FROM event_participants ep
		JOIN users u ON u.id = ep.participant_id
		WHERE ep.event_id = ?
		GROUP BY DATE_TRUNC('hour', ep.scanned_timestamp AT TIME ZONE 'Asia/Bangkok')
		ORDER BY DATE_TRUNC('hour', ep.scanned_timestamp AT TIME ZONE 'Asia/Bangkok') ASC
	`, eventID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}
