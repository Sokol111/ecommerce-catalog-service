package product

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T {
	return &v
}

func TestNewProduct(t *testing.T) {
	tests := []struct {
		name        string
		productName string
		description *string
		price       float32
		quantity    int
		imageID     *string
		categoryID  *string
		enabled     bool
		attributes  []AttributeValue
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid disabled product without optional fields",
			productName: "Test Product",
			description: nil,
			price:       0,
			quantity:    0,
			imageID:     nil,
			categoryID:  nil,
			enabled:     false,
			attributes:  nil,
			wantErr:     false,
		},
		{
			name:        "valid enabled product with all fields",
			productName: "Test Product",
			description: ptr("Test description"),
			price:       99.99,
			quantity:    10,
			imageID:     ptr("image-123"),
			categoryID:  ptr("category-456"),
			enabled:     true,
			attributes: []AttributeValue{
				{AttributeID: "attr-1", OptionSlugValue: ptr("red")},
			},
			wantErr: false,
		},
		{
			name:        "error when name is empty",
			productName: "",
			price:       10,
			quantity:    5,
			enabled:     false,
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name:        "error when name is too long",
			productName: strings.Repeat("a", 256),
			price:       10,
			quantity:    5,
			enabled:     false,
			wantErr:     true,
			errContains: "name is too long",
		},
		{
			name:        "error when price is negative",
			productName: "Test Product",
			price:       -1,
			quantity:    5,
			enabled:     false,
			wantErr:     true,
			errContains: "price must be positive",
		},
		{
			name:        "error when quantity is negative",
			productName: "Test Product",
			price:       10,
			quantity:    -1,
			enabled:     false,
			wantErr:     true,
			errContains: "quantity cannot be negative",
		},
		{
			name:        "error when enabling product with zero price",
			productName: "Test Product",
			price:       0,
			quantity:    10,
			imageID:     ptr("image-123"),
			categoryID:  ptr("category-456"),
			enabled:     true,
			wantErr:     true,
			errContains: "cannot enable product with price <= 0",
		},
		{
			name:        "error when enabling product with zero quantity",
			productName: "Test Product",
			price:       99.99,
			quantity:    0,
			imageID:     ptr("image-123"),
			categoryID:  ptr("category-456"),
			enabled:     true,
			wantErr:     true,
			errContains: "cannot enable product with quantity <= 0",
		},
		{
			name:        "error when enabling product without imageID",
			productName: "Test Product",
			price:       99.99,
			quantity:    10,
			imageID:     nil,
			categoryID:  ptr("category-456"),
			enabled:     true,
			wantErr:     true,
			errContains: "cannot enable product without imageID",
		},
		{
			name:        "error when enabling product without categoryID",
			productName: "Test Product",
			price:       99.99,
			quantity:    10,
			imageID:     ptr("image-123"),
			categoryID:  nil,
			enabled:     true,
			wantErr:     true,
			errContains: "cannot enable product without categoryID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, err := NewProduct(
				tt.productName,
				tt.description,
				tt.price,
				tt.quantity,
				tt.imageID,
				tt.categoryID,
				tt.enabled,
				tt.attributes,
			)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidProductData)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, product)
			} else {
				require.NoError(t, err)
				require.NotNil(t, product)
				assert.NotEmpty(t, product.ID)
				assert.Equal(t, 1, product.Version)
				assert.Equal(t, tt.productName, product.Name)
				assert.Equal(t, tt.description, product.Description)
				assert.Equal(t, tt.price, product.Price)
				assert.Equal(t, tt.quantity, product.Quantity)
				assert.Equal(t, tt.imageID, product.ImageID)
				assert.Equal(t, tt.categoryID, product.CategoryID)
				assert.Equal(t, tt.enabled, product.Enabled)
				assert.Equal(t, tt.attributes, product.Attributes)
				assert.False(t, product.CreatedAt.IsZero())
				assert.False(t, product.ModifiedAt.IsZero())
			}
		})
	}
}

func TestNewProductWithID(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		productName string
		price       float32
		quantity    int
		enabled     bool
		wantErr     bool
	}{
		{
			name:        "valid product with custom ID",
			id:          "custom-id-123",
			productName: "Test Product",
			price:       0,
			quantity:    0,
			enabled:     false,
			wantErr:     false,
		},
		{
			name:        "validation still applies with custom ID",
			id:          "custom-id-123",
			productName: "",
			price:       10,
			quantity:    5,
			enabled:     false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, err := NewProductWithID(
				tt.id,
				tt.productName,
				nil,
				tt.price,
				tt.quantity,
				nil,
				nil,
				tt.enabled,
				nil,
			)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, product)
			} else {
				require.NoError(t, err)
				require.NotNil(t, product)
				assert.Equal(t, tt.id, product.ID)
			}
		})
	}
}

func TestProduct_Update(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *Product
		newName     string
		description *string
		price       float32
		quantity    int
		imageID     *string
		categoryID  *string
		enabled     bool
		attributes  []AttributeValue
		wantErr     bool
		errContains string
	}{
		{
			name: "successful update",
			setup: func() *Product {
				p, _ := NewProduct("Original", nil, 0, 0, nil, nil, false, nil)
				return p
			},
			newName:     "Updated Name",
			description: ptr("New description"),
			price:       199.99,
			quantity:    20,
			imageID:     ptr("new-image"),
			categoryID:  ptr("new-category"),
			enabled:     true,
			attributes:  []AttributeValue{{AttributeID: "attr-1"}},
			wantErr:     false,
		},
		{
			name: "error when updating with empty name",
			setup: func() *Product {
				p, _ := NewProduct("Original", nil, 0, 0, nil, nil, false, nil)
				return p
			},
			newName:  "",
			price:    10,
			quantity: 5,
			enabled:  false,
			wantErr:  true,
		},
		{
			name: "error when enabling without required fields",
			setup: func() *Product {
				p, _ := NewProduct("Original", nil, 0, 0, nil, nil, false, nil)
				return p
			},
			newName:  "Updated",
			price:    99.99,
			quantity: 10,
			imageID:  nil,
			enabled:  true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product := tt.setup()
			originalModifiedAt := product.ModifiedAt

			err := product.Update(
				tt.newName,
				tt.description,
				tt.price,
				tt.quantity,
				tt.imageID,
				tt.categoryID,
				tt.enabled,
				tt.attributes,
			)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidProductData)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.newName, product.Name)
				assert.Equal(t, tt.description, product.Description)
				assert.Equal(t, tt.price, product.Price)
				assert.Equal(t, tt.quantity, product.Quantity)
				assert.Equal(t, tt.imageID, product.ImageID)
				assert.Equal(t, tt.categoryID, product.CategoryID)
				assert.Equal(t, tt.enabled, product.Enabled)
				assert.Equal(t, tt.attributes, product.Attributes)
				assert.True(t, product.ModifiedAt.After(originalModifiedAt) || product.ModifiedAt.Equal(originalModifiedAt))
			}
		})
	}
}

func TestReconstruct(t *testing.T) {
	t.Run("reconstructs product without validation", func(t *testing.T) {
		// Reconstruct should not validate - it's for rebuilding from persistence
		product := Reconstruct(
			"id-123",
			5,
			"", // Empty name would fail validation in NewProduct
			nil,
			-100, // Negative price would fail validation
			-50,  // Negative quantity would fail validation
			nil,
			nil,
			true, // Enabled without required fields
			nil,
			fixedTime(),
			fixedTime(),
		)

		require.NotNil(t, product)
		assert.Equal(t, "id-123", product.ID)
		assert.Equal(t, 5, product.Version)
		assert.Equal(t, "", product.Name)
		assert.Equal(t, float32(-100), product.Price)
		assert.Equal(t, -50, product.Quantity)
		assert.True(t, product.Enabled)
	})
}

func TestValidateProductData(t *testing.T) {
	tests := []struct {
		name        string
		productName string
		price       float32
		quantity    int
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid data",
			productName: "Product",
			price:       10,
			quantity:    5,
			wantErr:     false,
		},
		{
			name:        "valid with zero price",
			productName: "Product",
			price:       0,
			quantity:    5,
			wantErr:     false,
		},
		{
			name:        "valid with zero quantity",
			productName: "Product",
			price:       10,
			quantity:    0,
			wantErr:     false,
		},
		{
			name:        "name at max length",
			productName: strings.Repeat("a", 255),
			price:       10,
			quantity:    5,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProductData(tt.productName, tt.price, tt.quantity)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateEnabledState(t *testing.T) {
	tests := []struct {
		name        string
		enabled     bool
		price       float32
		quantity    int
		imageID     *string
		categoryID  *string
		wantErr     bool
		errContains string
	}{
		{
			name:       "disabled product - no validation",
			enabled:    false,
			price:      0,
			quantity:   0,
			imageID:    nil,
			categoryID: nil,
			wantErr:    false,
		},
		{
			name:       "enabled product with all requirements",
			enabled:    true,
			price:      99.99,
			quantity:   10,
			imageID:    ptr("image-123"),
			categoryID: ptr("category-456"),
			wantErr:    false,
		},
		{
			name:       "enabled but price is zero",
			enabled:    true,
			price:      0,
			quantity:   10,
			imageID:    ptr("image-123"),
			categoryID: ptr("category-456"),
			wantErr:    true,
		},
		{
			name:       "enabled but price is negative",
			enabled:    true,
			price:      -10,
			quantity:   10,
			imageID:    ptr("image-123"),
			categoryID: ptr("category-456"),
			wantErr:    true,
		},
		{
			name:       "enabled but quantity is zero",
			enabled:    true,
			price:      99.99,
			quantity:   0,
			imageID:    ptr("image-123"),
			categoryID: ptr("category-456"),
			wantErr:    true,
		},
		{
			name:       "enabled but no imageID",
			enabled:    true,
			price:      99.99,
			quantity:   10,
			imageID:    nil,
			categoryID: ptr("category-456"),
			wantErr:    true,
		},
		{
			name:       "enabled but no categoryID",
			enabled:    true,
			price:      99.99,
			quantity:   10,
			imageID:    ptr("image-123"),
			categoryID: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEnabledState(tt.enabled, tt.price, tt.quantity, tt.imageID, tt.categoryID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func fixedTime() (t time.Time) {
	return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
}
