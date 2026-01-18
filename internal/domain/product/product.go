package product

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ProductAttribute represents an attribute value assigned to a product
type ProductAttribute struct {
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
	Attributes  []ProductAttribute
	CreatedAt   time.Time
	ModifiedAt  time.Time
}

// NewProduct creates a new product with validation
func NewProduct(name string, description *string, price float32, quantity int, imageID *string, categoryID *string, enabled bool, attributes []ProductAttribute) (*Product, error) {
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
func NewProductWithID(id, name string, description *string, price float32, quantity int, imageID *string, categoryID *string, enabled bool, attributes []ProductAttribute) (*Product, error) {
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
func Reconstruct(id string, version int, name string, description *string, price float32, quantity int, imageID *string, categoryID *string, enabled bool, attributes []ProductAttribute, createdAt, modifiedAt time.Time) *Product {
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
func (p *Product) Update(name string, description *string, price float32, quantity int, imageID *string, categoryID *string, enabled bool, attributes []ProductAttribute) error {
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
		return errors.New("name is required")
	}

	if len(name) > 255 {
		return errors.New("name is too long (max 255 characters)")
	}

	if price < 0 {
		return errors.New("price must be positive")
	}

	if quantity < 0 {
		return errors.New("quantity cannot be negative")
	}

	return nil
}

// validateEnabledState validates that a product can be enabled
func validateEnabledState(enabled bool, price float32, quantity int, imageID *string, categoryID *string) error {
	if !enabled {
		return nil // No validation needed when disabling
	}

	if price <= 0 {
		return errors.New("cannot enable product: price must be greater than 0")
	}

	if quantity <= 0 {
		return errors.New("cannot enable product: quantity must be greater than 0")
	}

	if imageID == nil {
		return errors.New("cannot enable product: imageID is required")
	}

	if categoryID == nil {
		return errors.New("cannot enable product: categoryID is required")
	}

	return nil
}
