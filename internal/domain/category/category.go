package category

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AttributeRole defines how an attribute is used in a category
type AttributeRole string

const (
	// AttributeRoleVariant - creates product variants (color, size) - buyer can choose
	AttributeRoleVariant AttributeRole = "variant"
	// AttributeRoleSpecification - describes the product (processor, screen) - shown in specs
	AttributeRoleSpecification AttributeRole = "specification"
)

// CategoryAttribute represents an attribute assigned to a category
type CategoryAttribute struct {
	AttributeID string
	Slug        string // Attribute slug (immutable, stored for events)
	Role        AttributeRole
	Required    bool
	SortOrder   int
	Filterable  bool
	Searchable  bool
}

// Category - domain aggregate root
type Category struct {
	ID         string
	Version    int
	Name       string
	Enabled    bool
	Attributes []CategoryAttribute
	CreatedAt  time.Time
	ModifiedAt time.Time
}

// NewCategory creates a new category with validation
func NewCategory(name string, enabled bool, attributes []CategoryAttribute) (*Category, error) {
	if err := validateCategoryData(name); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &Category{
		ID:         uuid.New().String(),
		Version:    1,
		Name:       name,
		Enabled:    enabled,
		Attributes: attributes,
		CreatedAt:  now,
		ModifiedAt: now,
	}, nil
}

// NewCategoryWithID creates a category with a specific ID (for idempotency)
func NewCategoryWithID(id, name string, enabled bool, attributes []CategoryAttribute) (*Category, error) {
	if err := validateCategoryData(name); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &Category{
		ID:         id,
		Version:    1,
		Name:       name,
		Enabled:    enabled,
		Attributes: attributes,
		CreatedAt:  now,
		ModifiedAt: now,
	}, nil
}

// Reconstruct rebuilds a category from persistence (no validation)
func Reconstruct(id string, version int, name string, enabled bool, attributes []CategoryAttribute, createdAt, modifiedAt time.Time) *Category {
	return &Category{
		ID:         id,
		Version:    version,
		Name:       name,
		Enabled:    enabled,
		Attributes: attributes,
		CreatedAt:  createdAt,
		ModifiedAt: modifiedAt,
	}
}

// Update modifies category data with validation
func (c *Category) Update(name string, enabled bool, attributes []CategoryAttribute) error {
	if err := validateCategoryData(name); err != nil {
		return err
	}

	c.Name = name
	c.Enabled = enabled
	c.Attributes = attributes
	c.ModifiedAt = time.Now().UTC()

	return nil
}

// ChangeName updates the name with validation
func (c *Category) ChangeName(newName string) error {
	if err := validateCategoryData(newName); err != nil {
		return err
	}

	c.Name = newName
	c.ModifiedAt = time.Now().UTC()
	return nil
}

// Enable activates the category
func (c *Category) Enable() {
	c.Enabled = true
	c.ModifiedAt = time.Now().UTC()
}

// Disable deactivates the category
func (c *Category) Disable() {
	c.Enabled = false
	c.ModifiedAt = time.Now().UTC()
}

// IncrementVersion increments version for optimistic locking
func (c *Category) IncrementVersion() {
	c.Version++
}

// validateCategoryData validates business rules
func validateCategoryData(name string) error {
	if name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidCategoryData)
	}

	if len(name) > 255 {
		return fmt.Errorf("%w: name is too long (max 255 characters)", ErrInvalidCategoryData)
	}

	return nil
}
