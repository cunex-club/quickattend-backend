package response

import (
	"gorm.io/datatypes"
)

type DuplicateEventRes struct {
	DuplicatedEventId datatypes.UUID `json:"event_id"`
}
