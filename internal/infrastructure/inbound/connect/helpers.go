package connect

import (
	"github.com/google/uuid"
)

// parseUUIDPtr parses a string into a *uuid.UUID, returning nil on failure.
func parseUUIDPtr(s string) *uuid.UUID {
	u, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &u
}
