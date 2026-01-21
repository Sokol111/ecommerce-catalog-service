package mongo

import (
	"time"
)

// optionEntity represents an embedded attribute option in MongoDB
type optionEntity struct {
	Name      string  `bson:"name"`
	Slug      string  `bson:"slug"`
	ColorCode *string `bson:"colorCode,omitempty"`
	SortOrder int     `bson:"sortOrder"`
}

// attributeEntity represents the MongoDB document structure
type attributeEntity struct {
	ID         string         `bson:"_id"`
	Version    int            `bson:"version"`
	Name       string         `bson:"name"`
	Slug       string         `bson:"slug"`
	Type       string         `bson:"type"`
	Unit       *string        `bson:"unit,omitempty"`
	Enabled    bool           `bson:"enabled"`
	Options    []optionEntity `bson:"options,omitempty"`
	CreatedAt  time.Time      `bson:"createdAt"`
	ModifiedAt time.Time      `bson:"modifiedAt"`
}
