package entity

type CUNEXUserResponse struct {
	UserId      string `json:"userId"`
	UserType    string `json:"userType"`
	RefId       string `json:"refId"`
	FirstNameTH string `json:"firstNameTH"`
	LastNameTH  string `json:"lastNameTH"`
	FirstnameEN string `json:"firstNameEN"`
	LastNameEN  string `json:"lastNameEN"`
}

type CUNEXGetQRErrorResponse struct {
	ErrorCode string `json:"errorCode"`
	Message   string `json:"message"`
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
