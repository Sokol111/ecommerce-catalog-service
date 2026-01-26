package product

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AttributeValue represents an attribute value assigned to a product
type AttributeValue struct {
	AttributeID      string
	OptionSlugValue  *string  // Slug of selected option (for single type)
	OptionSlugValues []string // Slugs of selected options (for multiple type)
	NumericValue     *float32 // Numeric value (for range type)
	TextValue        *string  // Free text value (for text type)
	BooleanValue     *bool    // Boolean value (for boolean type)
}

// Product - domain aggregate root
type Product struct {
	ID          string
	Version     int
	Name        string
	Description *string
	Price       float32
	Quantity    int
	ImageID     *string
	CategoryID  *string
	Enabled     bool
	Attributes  []AttributeValue
	CreatedAt   time.Time
	ModifiedAt  time.Time
}

// NewProduct creates a new product with validation
func NewProduct(name string, description *string, price float32, quantity int, imageID *string, categoryID *string, enabled bool, attributes []AttributeValue) (*Product, error) {
	if err := validateProductData(name, price, quantity); err != nil {
		return nil, err
	}

	if err := validateEnabledState(enabled, price, quantity, imageID, categoryID); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &Product{
		ID:          uuid.New().String(),
		Version:     1,
		Name:        name,
		Description: description,
		Price:       price,
		Quantity:    quantity,
		ImageID:     imageID,
		CategoryID:  categoryID,
		Enabled:     enabled,
		Attributes:  attributes,
		CreatedAt:   now,
		ModifiedAt:  now,
	}, nil
}

// NewProductWithID creates a product with a specific ID (for idempotency)
func NewProductWithID(id, name string, description *string, price float32, quantity int, imageID *string, categoryID *string, enabled bool, attributes []AttributeValue) (*Product, error) {
	if err := validateProductData(name, price, quantity); err != nil {
		return nil, err
	}

	if err := validateEnabledState(enabled, price, quantity, imageID, categoryID); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &Product{
		ID:          id,
		Version:     1,
		Name:        name,
		Description: description,
		Price:       price,
		Quantity:    quantity,
		ImageID:     imageID,
		CategoryID:  categoryID,
		Enabled:     enabled,
		Attributes:  attributes,
		CreatedAt:   now,
		ModifiedAt:  now,
	}, nil
}

// Reconstruct rebuilds a product from persistence (no validation)
func Reconstruct(id string, version int, name string, description *string, price float32, quantity int, imageID *string, categoryID *string, enabled bool, attributes []AttributeValue, createdAt, modifiedAt time.Time) *Product {
	return &Product{
		ID:          id,
		Version:     version,
		Name:        name,
		Description: description,
		Price:       price,
		Quantity:    quantity,
		ImageID:     imageID,
		CategoryID:  categoryID,
		Enabled:     enabled,
		Attributes:  attributes,
		CreatedAt:   createdAt,
		ModifiedAt:  modifiedAt,
	}
}

// Update modifies product data with validation
func (p *Product) Update(name string, description *string, price float32, quantity int, imageID *string, categoryID *string, enabled bool, attributes []AttributeValue) error {
	if err := validateProductData(name, price, quantity); err != nil {
		return err
	}

	if err := validateEnabledState(enabled, price, quantity, imageID, categoryID); err != nil {
		return err
	}

	p.Name = name
	p.Description = description
	p.Price = price
	p.Quantity = quantity
	p.ImageID = imageID
	p.CategoryID = categoryID
	p.Enabled = enabled
	p.Attributes = attributes
	p.ModifiedAt = time.Now().UTC()

	return nil
}

// validateProductData validates business rules
func validateProductData(name string, price float32, quantity int) error {
	if name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidProductData)
	}

	if len(name) > 255 {
		return fmt.Errorf("%w: name is too long (max 255 characters)", ErrInvalidProductData)
	}

	if price < 0 {
		return fmt.Errorf("%w: price must be positive", ErrInvalidProductData)
	}

	if quantity < 0 {
		return fmt.Errorf("%w: quantity cannot be negative", ErrInvalidProductData)
	}

	return nil
}

// validateEnabledState validates that a product can be enabled
func validateEnabledState(enabled bool, price float32, quantity int, imageID *string, categoryID *string) error {
	if !enabled {
		return nil // No validation needed when disabling
	}

	if price <= 0 {
		return fmt.Errorf("%w: cannot enable product with price <= 0", ErrInvalidProductData)
	}

	if quantity <= 0 {
		return fmt.Errorf("%w: cannot enable product with quantity <= 0", ErrInvalidProductData)
	}

	if imageID == nil {
		return fmt.Errorf("%w: cannot enable product without imageID", ErrInvalidProductData)
	}

	if categoryID == nil {
		return fmt.Errorf("%w: cannot enable product without categoryID", ErrInvalidProductData)
	}

	return nil
}
