package entity

type CUNEXProfileResponse struct {
	Email         *string `json:"email"`
	FacultyCode   string  `json:"facultyCode"`
	FacultyNameEN string  `json:"facultyNameEN"`
	FacultyNameTH string  `json:"facultyNameTH"`
	FirstNameEN   string  `json:"firstNameEN"`
	FirstNameTH   string  `json:"firstNameTH"`
	LastNameEN    string  `json:"lastNameEN"`
	LastNameTH    string  `json:"lastNameTH"`
	RefId         string  `json:"refId"`
	StudentYear   string  `json:"studentYear"`
	TitleNameEN   string  `json:"titleNameEN"`
	TitleNameTH   string  `json:"titleNameTH"`
	UserId        string  `json:"userId"`
	UserType      string  `json:"userType"`
}

type UserTypes string

const (
	STUDENTS UserTypes = "student"
	STAFFS   UserTypes = "staff"
)

type CUNEXGetQRSuccessResponse struct {
	DepartmentNameEN string    `json:"departmentNameEN"`
	DepartmentNameTH string    `json:"departmentNameTH"`
	FacultyCode      string    `json:"facultyCode"`
	FacultyNameEN    string    `json:"facultyNameEN"`
	FacultyNameTH    string    `json:"facultyNameTH"`
	FirstNameEN      string    `json:"firstNameEN"`
	FirstNameTH      string    `json:"firstNameTH"`
	Gender           string    `json:"gender"`
	LastNameEN       string    `json:"lastNameEN"`
	LastNameTH       string    `json:"lastNameTH"`
	ProfileImageUrl  string    `json:"profileImageUrl"`
	RefId            string    `json:"refId"`
	UserType         UserTypes `json:"userType"`
}
