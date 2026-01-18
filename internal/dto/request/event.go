package response

type PostParticipantReqBody struct {
	EventId          string  `json:"event_id"`
	ScannedLocationX float64 `json:"scanned_location_long"`
	ScannedLocationY float64 `json:"scanned_location_lat"`
}
