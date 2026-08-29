package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
	"github.com/cunex-club/quickattend-backend/internal/entity"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DashboardRepository interface {
	GetRegistrationSummary(ctx context.Context, eventID uuid.UUID) (*dtoRes.RegistrationSummary, error)
	GetEventDashboardData(ctx context.Context, eventID uuid.UUID) (*dtoRes.EventDashboard, error)
	// SnapshotEventStats persists anonymized totals for eventID so they survive PurgeStaleParticipants.
	SnapshotEventStats(ctx context.Context, eventID uuid.UUID) error
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

func (r *repository) GetEventDashboardData(ctx context.Context, eventID uuid.UUID) (*dtoRes.EventDashboard, error) {
	var startTime time.Time
	if err := r.db.WithContext(ctx).
		Model(&entity.Event{}).
		Select("start_time").
		Where("id = ?", eventID).
		Scan(&startTime).Error; err != nil {
		return nil, err
	}
	if time.Now().UTC().Before(startTime.UTC()) {
		return nil, entity.ErrEventNotStarted
	}

	// Raw rows may already be purged for old events — use the snapshot if one exists.
	if snapshot, ok, err := r.getParticipantStatsSnapshot(ctx, eventID); err != nil {
		return nil, err
	} else if ok {
		return snapshot, nil
	}

	summary, err := r.GetRegistrationSummary(ctx, eventID)
	if err != nil {
		return nil, err
	}

	orgRows, err := r.getOrganizationStats(ctx, eventID)
	if err != nil {
		return nil, err
	}

	timeRows, err := r.getTimeStats(ctx, eventID)
	if err != nil {
		return nil, err
	}

	orgStats := make([]dtoRes.OrganizationStat, 0, len(orgRows))
	for _, row := range orgRows {
		orgStats = append(orgStats, dtoRes.OrganizationStat{
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

	return &dtoRes.EventDashboard{
		Summary:           *summary,
		OrganizationStats: orgStats,
		TimeSeriesStats:   timeStats,
	}, nil
}

// getEligibleCount returns the total number of eligible attendees for a
// WHITELIST-type event: the distinct union of confirmed whitelist rows,
// pending whitelist rows, and users who have already been scanned in.
// Including already-scanned users guarantees totalAll <= totalEligible even
// if an organizer removes a whitelisted user after they have attended.
// Returns nil for any non-whitelist attendance type.
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
				SUM(CASE WHEN u.user_type = 'student' THEN 1 ELSE 0 END) AS total_student,
				SUM(CASE WHEN u.user_type = 'staff' THEN 1 ELSE 0 END) AS total_staff,
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

func (r *repository) getOrganizationStats(ctx context.Context, eventID uuid.UUID) ([]entity.OrganizationStat, error) {
	var rows []entity.OrganizationStat

	err := r.db.WithContext(ctx).Raw(`
		SELECT
				COALESCE(ep.organization, 'ไม่ระบุ') AS organization,
				SUM(CASE WHEN u.user_type = 'student' THEN 1 ELSE 0 END) AS student_count,
				SUM(CASE WHEN u.user_type = 'staff' THEN 1 ELSE 0 END) AS staff_count,
			COUNT(*) AS total_count
		FROM event_participants ep
		JOIN users u ON u.id = ep.participant_id
		WHERE ep.event_id = ?
			GROUP BY COALESCE(ep.organization, 'ไม่ระบุ')
			ORDER BY total_count DESC, organization ASC
	`, eventID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// getTimeStats returns one row per hour covering the event window
// [DATE_TRUNC('hour', start_time), LEAST(end_time, NOW())], zero-filling
// hours with no scans so the frontend can render a chart without gap logic.
// `time_bucket` is emitted as an RFC 3339 UTC string (e.g. "2026-04-14T09:00:00Z").
func (r *repository) getTimeStats(ctx context.Context, eventID uuid.UUID) ([]entity.TimeStat, error) {
	var rows []entity.TimeStat

	err := r.db.WithContext(ctx).Raw(`
		WITH ev AS (
			SELECT
				DATE_TRUNC('hour', start_time) AS win_start,
				DATE_TRUNC('hour', LEAST(end_time, NOW())) AS win_end
			FROM events
			WHERE id = ?
		),
		buckets AS (
			SELECT generate_series(ev.win_start, ev.win_end, INTERVAL '1 hour') AS bucket
			FROM ev
		),
		scans AS (
			SELECT
				DATE_TRUNC('hour', ep.scanned_timestamp) AS bucket,
					u.user_type
			FROM event_participants ep
			JOIN users u ON u.id = ep.participant_id
			WHERE ep.event_id = ?
		)
		SELECT
			TO_CHAR(b.bucket AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS time_bucket,
				COALESCE(SUM(CASE WHEN s.user_type = 'student' THEN 1 ELSE 0 END), 0)::int AS student_count,
				COALESCE(SUM(CASE WHEN s.user_type = 'staff' THEN 1 ELSE 0 END), 0)::int AS staff_count,
			COUNT(s.bucket)::int AS total_count
		FROM buckets b
		LEFT JOIN scans s ON s.bucket = b.bucket
		GROUP BY b.bucket
		ORDER BY b.bucket ASC
	`, eventID, eventID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// getParticipantStatsSnapshot returns the persisted event_participant_stats
// row for eventID as an EventDashboard, if one exists. ok is false (with a
// nil error) when no snapshot has been written yet.
func (r *repository) getParticipantStatsSnapshot(ctx context.Context, eventID uuid.UUID) (*dtoRes.EventDashboard, bool, error) {
	var row entity.EventParticipantStats
	err := r.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var orgStats []dtoRes.OrganizationStat
	if err := json.Unmarshal(row.OrganizationStats, &orgStats); err != nil {
		return nil, false, err
	}
	var timeStats []dtoRes.TimeStat
	if err := json.Unmarshal(row.TimeSeriesStats, &timeStats); err != nil {
		return nil, false, err
	}

	return &dtoRes.EventDashboard{
		Summary: dtoRes.RegistrationSummary{
			TotalEligible: row.TotalEligible,
			TotalStudent:  row.TotalStudent,
			TotalStaff:    row.TotalStaff,
			TotalAll:      row.TotalParticipants,
		},
		OrganizationStats: orgStats,
		TimeSeriesStats:   timeStats,
	}, true, nil
}

// SnapshotEventStats computes this event's anonymized totals via the same 4 queries GetEventDashboardData uses, then upserts them into event_participant_stats.
func (r *repository) SnapshotEventStats(ctx context.Context, eventID uuid.UUID) error {
	summary, err := r.getParticipantSummary(ctx, eventID)
	if err != nil {
		return err
	}

	totalEligible, err := r.getEligibleCount(ctx, eventID)
	if err != nil {
		return err
	}

	orgRows, err := r.getOrganizationStats(ctx, eventID)
	if err != nil {
		return err
	}

	timeRows, err := r.getTimeStats(ctx, eventID)
	if err != nil {
		return err
	}

	orgJSON, err := json.Marshal(orgRows)
	if err != nil {
		return err
	}
	timeJSON, err := json.Marshal(timeRows)
	if err != nil {
		return err
	}

	stats := entity.EventParticipantStats{
		EventID:           datatypes.UUID(eventID),
		TotalParticipants: summary.TotalAll,
		TotalStudent:      summary.TotalStudent,
		TotalStaff:        summary.TotalStaff,
		TotalEligible:     totalEligible,
		OrganizationStats: datatypes.JSON(orgJSON),
		TimeSeriesStats:   datatypes.JSON(timeJSON),
	}

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&stats).Error
}
