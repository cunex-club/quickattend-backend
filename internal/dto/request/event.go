package response

type CheckInReq struct {
	EncodedOneTimeCode string `json:"one_time_code"`
}
