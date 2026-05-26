package graphql

import (
	"fmt"

	"github.com/google/uuid"
)

func parseEventID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid event id: %w", err)
	}

	return id, nil
}
