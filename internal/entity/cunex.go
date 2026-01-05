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
	UserType     UserTypes `json:"userType"`
	RefId        string    `json:"refId"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	Organization string    `json:"organization"`
	PhotoBase64  string    `json:"photoBase64"`
}
