package entity

type RegistrationSummary struct {
	TotalStudent int `gorm:"column:total_student"`
	TotalStaff   int `gorm:"column:total_staff"`
	TotalAll     int `gorm:"column:total_all"`
}