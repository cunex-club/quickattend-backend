package response

type GetUserByRefIdRes struct {
	RefID         string  `json:"ref_id"`
	UserType      string  `json:"user_type"`
	FirstnameTH   *string `json:"firstname_th"`
	SurnameTH     *string `json:"surname_th"`
	TitleTH       *string `json:"title_th"`
	FacultyNameTH *string `json:"faculty_name_th"`
	FirstnameEN   *string `json:"firstname_en"`
	SurnameEN     *string `json:"surname_en"`
	TitleEN       *string `json:"title_en"`
	FacultyNameEN *string `json:"faculty_name_en"`
}
