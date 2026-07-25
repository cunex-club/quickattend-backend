package response

type RegistrationSummary struct {
	TotalEligible *int `json:"totalEligible"`
	TotalStudent  int  `json:"totalStudent"`
	TotalStaff    int  `json:"totalStaff"`
	TotalAll      int  `json:"totalAll"`
}

type OrganizationStat struct {
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

type EventDashboard struct {
	Summary           RegistrationSummary `json:"summary"`
	OrganizationStats []OrganizationStat  `json:"organizationStats"`
	TimeSeriesStats   []TimeStat          `json:"timeSeriesStats"`
}
