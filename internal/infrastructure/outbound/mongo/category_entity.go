package mongo

import (
	"time"
)

// CategoryAttributeEntity represents embedded category attribute in MongoDB
type CategoryAttributeEntity struct {
	AttributeID string `bson:"attributeId"`
	Slug        string `bson:"slug"`
	Role        string `bson:"role"`
	SortOrder   int32  `bson:"sortOrder"`
	Filterable  bool   `bson:"filterable"`
	Searchable  bool   `bson:"searchable"`
}

// CategoryEntity represents the MongoDB document structure
type CategoryEntity struct {
	ID         string                    `bson:"_id"`
	Version    int64                     `bson:"version"`
	Name       string                    `bson:"name"`
	Enabled    bool                      `bson:"enabled"`
	Attributes []CategoryAttributeEntity `bson:"attributes,omitempty"`
	CreatedAt  time.Time                 `bson:"createdAt"`
	ModifiedAt time.Time                 `bson:"modifiedAt"`
}
