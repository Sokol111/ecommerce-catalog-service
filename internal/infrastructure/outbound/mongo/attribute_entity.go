package mongo

import (
	"time"
)

// OptionEntity represents an embedded attribute option in MongoDB
type OptionEntity struct {
	Name      string  `bson:"name"`
	Slug      string  `bson:"slug"`
	ColorCode *string `bson:"colorCode,omitempty"`
	SortOrder int32   `bson:"sortOrder"`
}

// AttributeEntity represents the MongoDB document structure
type AttributeEntity struct {
	ID         string         `bson:"_id"`
	Version    int64          `bson:"version"`
	Name       string         `bson:"name"`
	Slug       string         `bson:"slug"`
	Type       string         `bson:"type"`
	Unit       *string        `bson:"unit,omitempty"`
	Enabled    bool           `bson:"enabled"`
	Options    []OptionEntity `bson:"options,omitempty"`
	CreatedAt  time.Time      `bson:"createdAt"`
	ModifiedAt time.Time      `bson:"modifiedAt"`
}
