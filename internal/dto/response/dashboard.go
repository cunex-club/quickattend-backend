package response

type RegistrationSummary struct {
	TotalEligible int `json:"totalEligible"`
	TotalStudent  int `json:"totalStudent"`
	TotalStaff    int `json:"totalStaff"`
	TotalAll      int `json:"totalAll"`
}
