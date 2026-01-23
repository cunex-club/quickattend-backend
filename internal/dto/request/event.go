package response

type CheckInReq struct {
	Comment               string `json:"comment"`
	EncodedOneTimeCode string `json:"one_time_code"`
}
