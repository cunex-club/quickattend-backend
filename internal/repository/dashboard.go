package repository

import (
	"context"
	"time"

	"github.com/cunex-club/quickattend-backend/internal/entity"
	"github.com/google/uuid"
)

type DashboardRepository interface {
	GetDashboardEvent(ctx context.Context, eventID uuid.UUID) (*DashboardEventRow, error)
	CountRegisteredParticipants(ctx context.Context, eventID uuid.UUID) (int, error)
	CountExpectedWhitelistParticipants(ctx context.Context, eventID uuid.UUID) (int, error)
	GetExpectedFacultyCodes(ctx context.Context, eventID uuid.UUID, mode entity.AttendanceType) ([]int, error)
	GetRegisteredParticipantRows(ctx context.Context, eventID uuid.UUID) ([]DashboardRegisteredParticipantRow, error)
	GetWhitelistParticipantRows(ctx context.Context, eventID uuid.UUID) ([]DashboardWhitelistParticipantRow, error)
}

type DashboardEventRow struct {
	AttendenceType entity.AttendanceType `gorm:"column:attendence_type"`
	StartTime      time.Time             `gorm:"column:start_time"`
	EndTime        time.Time             `gorm:"column:end_time"`
}

type DashboardRegisteredParticipantRow struct {
	RefID            uint64    `gorm:"column:ref_id"`
	ScannedTimestamp time.Time `gorm:"column:scanned_timestamp"`
}

type DashboardWhitelistParticipantRow struct {
	ID           string     `gorm:"column:id"`
	RefID        uint64     `gorm:"column:ref_id"`
	FullName     string     `gorm:"column:full_name"`
	RegisteredAt *time.Time `gorm:"column:registered_at"`
}

func (r *repository) GetDashboardEvent(ctx context.Context, eventID uuid.UUID) (*DashboardEventRow, error) {
	var row DashboardEventRow

	err := r.db.WithContext(ctx).
		Table("events").
		Select("attendence_type, start_time, end_time").
		Where("id = ?", eventID).
		First(&row).
		Error

	if err != nil {
		return nil, err
	}

	return &row, nil
}

func (r *repository) CountRegisteredParticipants(ctx context.Context, eventID uuid.UUID) (int, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Table("event_participants").
		Where("event_id = ?", eventID).
		Count(&count).
		Error

	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (r *repository) CountExpectedWhitelistParticipants(ctx context.Context, eventID uuid.UUID) (int, error) {
	var count int

	err := r.db.WithContext(ctx).
		Raw(`
			SELECT COUNT(*) FROM (
				SELECT attendee_ref_id
				FROM event_whitelists
				WHERE event_id = ?

				UNION

				SELECT attendee_ref_id
				FROM event_whitelist_pendings
				WHERE event_id = ?
			) AS expected_whitelist
		`, eventID, eventID).
		Scan(&count).
		Error

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *repository) GetExpectedFacultyCodes(ctx context.Context, eventID uuid.UUID, mode entity.AttendanceType) ([]int, error) {
	var codes []int

	switch mode {
	case entity.AttendanceFaculties:
		err := r.db.WithContext(ctx).
			Table("event_allowed_faculties").
			Select("faculty_no").
			Where("event_id = ?", eventID).
			Order("faculty_no").
			Scan(&codes).
			Error

		if err != nil {
			return nil, err
		}

	case entity.AttendanceWhitelist:
		err := r.db.WithContext(ctx).
			Raw(`
				SELECT DISTINCT (attendee_ref_id % 100)::int AS faculty_code
				FROM (
					SELECT attendee_ref_id
					FROM event_whitelists
					WHERE event_id = ?

					UNION

					SELECT attendee_ref_id
					FROM event_whitelist_pendings
					WHERE event_id = ?
				) AS whitelist_union
				ORDER BY faculty_code
			`, eventID, eventID).
			Scan(&codes).
			Error

		if err != nil {
			return nil, err
		}

	default:
		return []int{}, nil
	}

	return codes, nil
}

func (r *repository) GetRegisteredParticipantRows(ctx context.Context, eventID uuid.UUID) ([]DashboardRegisteredParticipantRow, error) {
	var rows []DashboardRegisteredParticipantRow

	err := r.db.WithContext(ctx).
		Raw(`
			SELECT
				u.ref_id AS ref_id,
				ep.scanned_timestamp AS scanned_timestamp
			FROM event_participants ep
			JOIN users u ON u.id = ep.participant_id
			WHERE ep.event_id = ?
			ORDER BY ep.scanned_timestamp ASC
		`, eventID).
		Scan(&rows).
		Error

	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *repository) GetWhitelistParticipantRows(ctx context.Context, eventID uuid.UUID) ([]DashboardWhitelistParticipantRow, error) {
	var rows []DashboardWhitelistParticipantRow

	err := r.db.WithContext(ctx).
		Raw(`
			WITH whitelist_union AS (
				SELECT id::text AS id, attendee_ref_id
				FROM event_whitelists
				WHERE event_id = ?

				UNION ALL

				SELECT id::text AS id, attendee_ref_id
				FROM event_whitelist_pendings
				WHERE event_id = ?
			),
			whitelist_dedup AS (
				SELECT
					MIN(id) AS id,
					attendee_ref_id
				FROM whitelist_union
				GROUP BY attendee_ref_id
			)
			SELECT
				w.id AS id,
				w.attendee_ref_id AS ref_id,
				COALESCE(
					NULLIF(TRIM(CONCAT_WS(' ', NULLIF(u.title_th, ''), u.firstname_th, u.surname_th)), ''),
					w.attendee_ref_id::text
				) AS full_name,
				ep.scanned_timestamp AS registered_at
			FROM whitelist_dedup w
			LEFT JOIN users u ON u.ref_id = w.attendee_ref_id
			LEFT JOIN event_participants ep
				ON ep.event_id = ?
				AND ep.participant_id = u.id
			ORDER BY w.attendee_ref_id ASC
		`, eventID, eventID, eventID).
		Scan(&rows).
		Error

	if err != nil {
		return nil, err
	}

	return rows, nil
}