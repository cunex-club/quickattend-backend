package response

type VerifyTokenRes struct {
	AccessToken string `json:"access_token"`
}

type GetAuthUserRes struct {
	ID          string `json:"id"`
	RefID       uint64 `json:"ref_id"`
	FirstnameTH string `json:"firstname_th"`
	SurnameTH   string `json:"surname_th"`
	TitleTH     string `json:"title_th"`
	FirstnameEN string `json:"firstname_en"`
	SurnameEN   string `json:"surname_en"`
	TitleEN     string `json:"title_en"`
}
