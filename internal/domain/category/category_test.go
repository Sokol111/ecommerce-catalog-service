package category

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCategory(t *testing.T) {
	tests := []struct {
		name        string
		catName     string
		enabled     bool
		attributes  []CategoryAttribute
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid category without attributes",
			catName:    "Electronics",
			enabled:    true,
			attributes: nil,
			wantErr:    false,
		},
		{
			name:    "valid category with attributes",
			catName: "Clothing",
			enabled: false,
			attributes: []CategoryAttribute{
				{
					AttributeID: "attr-1",
					Slug:        "color",
					Role:        AttributeRoleVariant,
					Required:    true,
					SortOrder:   1,
					Filterable:  true,
					Searchable:  true,
				},
				{
					AttributeID: "attr-2",
					Slug:        "size",
					Role:        AttributeRoleSpecification,
					Required:    false,
					SortOrder:   2,
					Filterable:  true,
					Searchable:  false,
				},
			},
			wantErr: false,
		},
		{
			name:        "error when name is empty",
			catName:     "",
			enabled:     true,
			attributes:  nil,
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name:        "error when name is too long",
			catName:     strings.Repeat("a", 256),
			enabled:     true,
			attributes:  nil,
			wantErr:     true,
			errContains: "name is too long",
		},
		{
			name:       "valid name at max length",
			catName:    strings.Repeat("a", 255),
			enabled:    true,
			attributes: nil,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category, err := NewCategory(tt.catName, tt.enabled, tt.attributes)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidCategoryData)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, category)
			} else {
				require.NoError(t, err)
				require.NotNil(t, category)
				assert.NotEmpty(t, category.ID)
				assert.Equal(t, 1, category.Version)
				assert.Equal(t, tt.catName, category.Name)
				assert.Equal(t, tt.enabled, category.Enabled)
				assert.Equal(t, tt.attributes, category.Attributes)
				assert.False(t, category.CreatedAt.IsZero())
				assert.False(t, category.ModifiedAt.IsZero())
			}
		})
	}
}

func TestNewCategoryWithID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		catName string
		enabled bool
		wantErr bool
	}{
		{
			name:    "valid category with custom ID",
			id:      "custom-cat-id",
			catName: "Electronics",
			enabled: true,
			wantErr: false,
		},
		{
			name:    "validation still applies with custom ID",
			id:      "custom-cat-id",
			catName: "",
			enabled: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category, err := NewCategoryWithID(tt.id, tt.catName, tt.enabled, nil)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, category)
			} else {
				require.NoError(t, err)
				require.NotNil(t, category)
				assert.Equal(t, tt.id, category.ID)
				assert.Equal(t, tt.catName, category.Name)
			}
		})
	}
}

func TestCategory_Update(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() *Category
		newName    string
		enabled    bool
		attributes []CategoryAttribute
		wantErr    bool
	}{
		{
			name: "successful update",
			setup: func() *Category {
				c, _ := NewCategory("Original", false, nil)
				return c
			},
			newName: "Updated Name",
			enabled: true,
			attributes: []CategoryAttribute{
				{AttributeID: "attr-1", Slug: "color", Role: AttributeRoleVariant},
			},
			wantErr: false,
		},
		{
			name: "error when updating with empty name",
			setup: func() *Category {
				c, _ := NewCategory("Original", false, nil)
				return c
			},
			newName: "",
			enabled: false,
			wantErr: true,
		},
		{
			name: "error when updating with too long name",
			setup: func() *Category {
				c, _ := NewCategory("Original", false, nil)
				return c
			},
			newName: strings.Repeat("a", 256),
			enabled: false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category := tt.setup()
			originalModifiedAt := category.ModifiedAt

			// Small delay to ensure ModifiedAt changes
			time.Sleep(time.Millisecond)

			err := category.Update(tt.newName, tt.enabled, tt.attributes)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidCategoryData)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.newName, category.Name)
				assert.Equal(t, tt.enabled, category.Enabled)
				assert.Equal(t, tt.attributes, category.Attributes)
				assert.True(t, category.ModifiedAt.After(originalModifiedAt))
			}
		})
	}
}

func TestCategory_ChangeName(t *testing.T) {
	tests := []struct {
		name        string
		newName     string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid name change",
			newName: "New Category Name",
			wantErr: false,
		},
		{
			name:        "error when new name is empty",
			newName:     "",
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name:        "error when new name is too long",
			newName:     strings.Repeat("a", 256),
			wantErr:     true,
			errContains: "name is too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category, _ := NewCategory("Original", false, nil)
			originalModifiedAt := category.ModifiedAt

			time.Sleep(time.Millisecond)

			err := category.ChangeName(tt.newName)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidCategoryData)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.newName, category.Name)
				assert.True(t, category.ModifiedAt.After(originalModifiedAt))
			}
		})
	}
}

func TestCategory_Enable(t *testing.T) {
	category, _ := NewCategory("Test", false, nil)
	assert.False(t, category.Enabled)

	originalModifiedAt := category.ModifiedAt
	time.Sleep(time.Millisecond)

	category.Enable()

	assert.True(t, category.Enabled)
	assert.True(t, category.ModifiedAt.After(originalModifiedAt))
}

func TestCategory_Disable(t *testing.T) {
	category, _ := NewCategory("Test", true, nil)
	assert.True(t, category.Enabled)

	originalModifiedAt := category.ModifiedAt
	time.Sleep(time.Millisecond)

	category.Disable()

	assert.False(t, category.Enabled)
	assert.True(t, category.ModifiedAt.After(originalModifiedAt))
}

func TestCategory_IncrementVersion(t *testing.T) {
	category, _ := NewCategory("Test", false, nil)
	assert.Equal(t, 1, category.Version)

	category.IncrementVersion()
	assert.Equal(t, 2, category.Version)

	category.IncrementVersion()
	assert.Equal(t, 3, category.Version)
}

func TestReconstruct(t *testing.T) {
	t.Run("reconstructs category without validation", func(t *testing.T) {
		createdAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		modifiedAt := time.Date(2025, 6, 15, 12, 30, 0, 0, time.UTC)
		attributes := []CategoryAttribute{
			{AttributeID: "attr-1", Slug: "color"},
		}

		// Reconstruct should not validate - it's for rebuilding from persistence
		category := Reconstruct(
			"cat-123",
			5,
			"", // Empty name would fail validation in NewCategory
			true,
			attributes,
			createdAt,
			modifiedAt,
		)

		require.NotNil(t, category)
		assert.Equal(t, "cat-123", category.ID)
		assert.Equal(t, 5, category.Version)
		assert.Equal(t, "", category.Name)
		assert.True(t, category.Enabled)
		assert.Equal(t, attributes, category.Attributes)
		assert.Equal(t, createdAt, category.CreatedAt)
		assert.Equal(t, modifiedAt, category.ModifiedAt)
	})
}

func TestAttributeRoleConstants(t *testing.T) {
	assert.Equal(t, AttributeRole("variant"), AttributeRoleVariant)
	assert.Equal(t, AttributeRole("specification"), AttributeRoleSpecification)
}

func TestCategoryAttribute(t *testing.T) {
	attr := CategoryAttribute{
		AttributeID: "attr-123",
		Slug:        "color",
		Role:        AttributeRoleVariant,
		Required:    true,
		SortOrder:   1,
		Filterable:  true,
		Searchable:  false,
	}

	assert.Equal(t, "attr-123", attr.AttributeID)
	assert.Equal(t, "color", attr.Slug)
	assert.Equal(t, AttributeRoleVariant, attr.Role)
	assert.True(t, attr.Required)
	assert.Equal(t, 1, attr.SortOrder)
	assert.True(t, attr.Filterable)
	assert.False(t, attr.Searchable)
}
