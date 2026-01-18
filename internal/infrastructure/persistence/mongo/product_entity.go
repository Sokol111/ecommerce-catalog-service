package mongo

import (
	"time"
)

// productAttributeEntity represents an attribute value in MongoDB
type productAttributeEntity struct {
	AttributeID      string   `bson:"attributeId"`
	OptionSlugValue  *string  `bson:"optionSlugValue,omitempty"`
	OptionSlugValues []string `bson:"optionSlugValues,omitempty"`
	NumericValue     *float32 `bson:"numericValue,omitempty"`
	TextValue        *string  `bson:"textValue,omitempty"`
	BooleanValue     *bool    `bson:"booleanValue,omitempty"`
}

// productEntity represents the MongoDB document structure
type productEntity struct {
	ID          string                   `bson:"_id"`
	Version     int                      `bson:"version"`
	Name        string                   `bson:"name"`
	Description *string                  `bson:"description,omitempty"`
	Price       float32                  `bson:"price"`
	Quantity    int                      `bson:"quantity"`
	ImageID     *string                  `bson:"imageId,omitempty"`
	CategoryID  *string                  `bson:"categoryId,omitempty"`
	Enabled     bool                     `bson:"enabled"`
	Attributes  []productAttributeEntity `bson:"attributes,omitempty"`
	CreatedAt   time.Time                `bson:"createdAt"`
	ModifiedAt  time.Time                `bson:"modifiedAt"`
}
