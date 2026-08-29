package entity

import (
	"time"

	"gorm.io/datatypes"
)

type EventParticipantStats struct {
	EventID           datatypes.UUID `gorm:"column:event_id"`
	TotalParticipants int            `gorm:"column:total_participants"`
	TotalStudent      int            `gorm:"column:total_student"`
	TotalStaff        int            `gorm:"column:total_staff"`
	TotalEligible      *int           `gorm:"column:total_eligible"`
	OrganizationStats datatypes.JSON `gorm:"column:organization_stats"`
	TimeSeriesStats   datatypes.JSON `gorm:"column:time_series_stats"`
	CreatedAt         time.Time      `gorm:"column:created_at"`
}

func (EventParticipantStats) TableName() string {
	return "event_participant_stats"
}

type RegistrationSummary struct {
	TotalStudent int `gorm:"column:total_student"`
	TotalStaff   int `gorm:"column:total_staff"`
	TotalAll     int `gorm:"column:total_all"`
}

type OrganizationStat struct {
	Organization string `gorm:"column:organization"`
	StudentCount int    `gorm:"column:student_count"`
	StaffCount   int    `gorm:"column:staff_count"`
	TotalCount   int    `gorm:"column:total_count"`
}

type TimeStat struct {
	TimeBucket   string `gorm:"column:time_bucket"`
	StudentCount int    `gorm:"column:student_count"`
	StaffCount   int    `gorm:"column:staff_count"`
	TotalCount   int    `gorm:"column:total_count"`
}
