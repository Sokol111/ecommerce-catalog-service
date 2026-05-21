package attribute

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// AttributeType represents the type of attribute
type AttributeType string

const (
	AttributeTypeSingle   AttributeType = "single"
	AttributeTypeMultiple AttributeType = "multiple"
	AttributeTypeRange    AttributeType = "range"
	AttributeTypeBoolean  AttributeType = "boolean"
	AttributeTypeText     AttributeType = "text"
)

// Option represents an attribute option (embedded in Attribute)
type Option struct {
	Name      string
	Slug      string
	ColorCode *string
	SortOrder int
}

// Attribute - domain aggregate root
type Attribute struct {
	ID         string
	Version    int
	Name       string
	Slug       string
	Type       AttributeType
	Unit       *string
	Enabled    bool
	Options    []Option
	CreatedAt  time.Time
	ModifiedAt time.Time
}

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// NewAttribute creates a new attribute with validation.
// If id is empty, a new UUID will be generated.
func NewAttribute(
	id string,
	name string,
	slug string,
	attrType AttributeType,
	unit *string,
	enabled bool,
	options []Option,
) (*Attribute, error) {
	if err := validateAttributeData(name, slug, attrType); err != nil {
		return nil, err
	}

	if err := validateOptions(options); err != nil {
		return nil, err
	}

	if id == "" {
		id = uuid.New().String()
	}

	now := time.Now().UTC()
	return &Attribute{
		ID:         id,
		Version:    1,
		Name:       name,
		Slug:       slug,
		Type:       attrType,
		Unit:       unit,
		Enabled:    enabled,
		Options:    options,
		CreatedAt:  now,
		ModifiedAt: now,
	}, nil
}

// Reconstruct rebuilds an attribute from persistence (no validation)
func Reconstruct(
	id string,
	version int,
	name string,
	slug string,
	attrType AttributeType,
	unit *string,
	enabled bool,
	options []Option,
	createdAt time.Time,
	modifiedAt time.Time,
) *Attribute {
	return &Attribute{
		ID:         id,
		Version:    version,
		Name:       name,
		Slug:       slug,
		Type:       attrType,
		Unit:       unit,
		Enabled:    enabled,
		Options:    options,
		CreatedAt:  createdAt,
		ModifiedAt: modifiedAt,
	}
}

// Update modifies attribute data with validation
// Note: slug and type are immutable and cannot be changed after creation
func (a *Attribute) Update(
	name string,
	unit *string,
	enabled bool,
	options []Option,
) error {
	if name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidAttributeData)
	}

	if len(name) > 100 {
		return fmt.Errorf("%w: name is too long (max 100 characters)", ErrInvalidAttributeData)
	}

	if err := validateOptions(options); err != nil {
		return err
	}

	a.Name = name
	a.Unit = unit
	a.Enabled = enabled
	a.Options = options
	a.ModifiedAt = time.Now().UTC()

	return nil
}

// validateAttributeData validates business rules
func validateAttributeData(name string, slug string, attrType AttributeType) error {
	if name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidAttributeData)
	}

	if len(name) > 100 {
		return fmt.Errorf("%w: name is too long (max 100 characters)", ErrInvalidAttributeData)
	}

	if slug == "" {
		return fmt.Errorf("%w: slug is required", ErrInvalidAttributeData)
	}

	if len(slug) > 50 {
		return fmt.Errorf("%w: slug is too long (max 50 characters)", ErrInvalidAttributeData)
	}

	if !slugRegex.MatchString(slug) {
		return fmt.Errorf("%w: slug must contain only lowercase letters, numbers, and hyphens", ErrInvalidAttributeData)
	}

	if !isValidAttributeType(attrType) {
		return fmt.Errorf("%w: invalid attribute type", ErrInvalidAttributeData)
	}

	return nil
}

func isValidAttributeType(t AttributeType) bool {
	switch t {
	case AttributeTypeSingle, AttributeTypeMultiple, AttributeTypeRange, AttributeTypeBoolean, AttributeTypeText:
		return true
	}
	return false
}

// validateOptions validates option data
func validateOptions(options []Option) error {
	if len(options) == 0 {
		return nil
	}

	slugs := make(map[string]bool)
	for _, opt := range options {
		if opt.Name == "" {
			return fmt.Errorf("%w: option name is required", ErrInvalidAttributeData)
		}
		if len(opt.Name) > 100 {
			return fmt.Errorf("%w: option name is too long (max 100 characters)", ErrInvalidAttributeData)
		}
		if opt.Slug == "" {
			return fmt.Errorf("%w: option slug is required", ErrInvalidAttributeData)
		}
		if len(opt.Slug) > 50 {
			return fmt.Errorf("%w: option slug is too long (max 50 characters)", ErrInvalidAttributeData)
		}
		if !slugRegex.MatchString(opt.Slug) {
			return fmt.Errorf("%w: option slug must contain only lowercase letters, numbers, and hyphens", ErrInvalidAttributeData)
		}
		if slugs[opt.Slug] {
			return fmt.Errorf("%w: duplicate option slug: %s", ErrInvalidAttributeData, opt.Slug)
		}
		slugs[opt.Slug] = true
		if opt.SortOrder < 0 {
			return fmt.Errorf("%w: option sortOrder cannot be negative", ErrInvalidAttributeData)
		}
	}
	return nil
}
