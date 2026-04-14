package response

type RegistrationSummary struct {
	TotalEligible *int `json:"totalEligible"`
	TotalStudent  int  `json:"totalStudent"`
	TotalStaff    int  `json:"totalStaff"`
	TotalAll      int  `json:"totalAll"`
}

type FacultyStat struct {
	Organization string `json:"organization"`
	StudentCount int    `json:"studentCount"`
	StaffCount   int    `json:"staffCount"`
	TotalCount   int    `json:"totalCount"`
}

type TimeStat struct {
	TimeBucket   string `json:"timeBucket"`
	StudentCount int    `json:"studentCount"`
	StaffCount   int    `json:"staffCount"`
	TotalCount   int    `json:"totalCount"`
}

type DashboardReadyData struct {
	Summary         RegistrationSummary `json:"summary"`
	FacultyStats    []FacultyStat       `json:"facultyStats"`
	TimeSeriesStats []TimeStat          `json:"timeSeriesStats"`
}
