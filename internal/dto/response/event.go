package response

import "time"

type status string

const (
	SUCCESS   status = "success"
	DUPLICATE status = "duplicate"
	FAIL      status = "fail"
	LATE      status = "late"
)

type GetParticipantRes struct {
	FirstnameTH  string    `json:"firstname_th"`
	SurnameTH    string    `json:"surname_th"`
	TitleTH      string    `json:"title_th"`
	FirstnameEN  string    `json:"firstname_en"`
	SurnameEN    string    `json:"surname_en"`
	TitleEN      string    `json:"title_en"`
	RefID        uint64    `json:"ref_id"`
	Organization string    `json:"organization"`
	CheckInTime  time.Time `json:"check_in_time"`
	Status       string    `json:"status"`
	Code         string    `json:"code"`
}
