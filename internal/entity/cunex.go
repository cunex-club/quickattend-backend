package entity

type CUNEXUserResponse struct {
	UserId      string `json:"userId"`
	UserType    string `json:"userType"`
	RefId       string `json:"refId"`
	FirstNameTH string `json:"firstNameTH"`
	LastNameTH  string `json:"lastNameTH"`
	FirstnameEN string `json:"firstNameEN"`
	SurnameEN   string `json:"lastNameEN"`
}
