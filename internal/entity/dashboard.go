package entity

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
