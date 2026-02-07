package response

type CheckInReq struct {
	Comment            string `json:"comment"`
	EncodedOneTimeCode string `json:"one_time_code"`
}

type PostParticipantReqBody struct {
	EventId          string  `json:"event_id"`
	ScannedLocationX float64 `json:"scanned_location_long"`
	ScannedLocationY float64 `json:"scanned_location_lat"`
}
