package attribute

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

func TestNewAttribute(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		attrName    string
		slug        string
		attrType    AttributeType
		unit        *string
		enabled     bool
		options     []Option
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid single type attribute without options",
			id:       "",
			attrName: "Color",
			slug:     "color",
			attrType: AttributeTypeSingle,
			unit:     nil,
			enabled:  true,
			options:  nil,
			wantErr:  false,
		},
		{
			name:     "valid attribute with custom ID",
			id:       "custom-attr-id",
			attrName: "Size",
			slug:     "size",
			attrType: AttributeTypeMultiple,
			unit:     ptr("cm"),
			enabled:  false,
			options:  nil,
			wantErr:  false,
		},
		{
			name:     "valid attribute with options",
			id:       "",
			attrName: "Color",
			slug:     "color",
			attrType: AttributeTypeSingle,
			unit:     nil,
			enabled:  true,
			options: []Option{
				{Name: "Red", Slug: "red", ColorCode: ptr("#FF0000"), SortOrder: 1},
				{Name: "Blue", Slug: "blue", ColorCode: ptr("#0000FF"), SortOrder: 2},
			},
			wantErr: false,
		},
		{
			name:     "valid range type with unit",
			id:       "",
			attrName: "Weight",
			slug:     "weight",
			attrType: AttributeTypeRange,
			unit:     ptr("kg"),
			enabled:  true,
			options:  nil,
			wantErr:  false,
		},
		{
			name:     "valid boolean type",
			id:       "",
			attrName: "Is Organic",
			slug:     "is-organic",
			attrType: AttributeTypeBoolean,
			enabled:  true,
			wantErr:  false,
		},
		{
			name:     "valid text type",
			id:       "",
			attrName: "Description",
			slug:     "description",
			attrType: AttributeTypeText,
			enabled:  true,
			wantErr:  false,
		},
		{
			name:        "error when name is empty",
			id:          "",
			attrName:    "",
			slug:        "test",
			attrType:    AttributeTypeSingle,
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name:        "error when name is too long",
			id:          "",
			attrName:    strings.Repeat("a", 101),
			slug:        "test",
			attrType:    AttributeTypeSingle,
			wantErr:     true,
			errContains: "name is too long",
		},
		{
			name:        "error when slug is empty",
			id:          "",
			attrName:    "Test",
			slug:        "",
			attrType:    AttributeTypeSingle,
			wantErr:     true,
			errContains: "slug is required",
		},
		{
			name:        "error when slug is too long",
			id:          "",
			attrName:    "Test",
			slug:        strings.Repeat("a", 51),
			attrType:    AttributeTypeSingle,
			wantErr:     true,
			errContains: "slug is too long",
		},
		{
			name:        "error when slug has invalid characters - uppercase",
			id:          "",
			attrName:    "Test",
			slug:        "Invalid-Slug",
			attrType:    AttributeTypeSingle,
			wantErr:     true,
			errContains: "slug must contain only lowercase",
		},
		{
			name:        "error when slug has invalid characters - spaces",
			id:          "",
			attrName:    "Test",
			slug:        "invalid slug",
			attrType:    AttributeTypeSingle,
			wantErr:     true,
			errContains: "slug must contain only lowercase",
		},
		{
			name:        "error when slug has invalid characters - special chars",
			id:          "",
			attrName:    "Test",
			slug:        "invalid_slug",
			attrType:    AttributeTypeSingle,
			wantErr:     true,
			errContains: "slug must contain only lowercase",
		},
		{
			name:        "error when slug starts with hyphen",
			id:          "",
			attrName:    "Test",
			slug:        "-invalid",
			attrType:    AttributeTypeSingle,
			wantErr:     true,
			errContains: "slug must contain only lowercase",
		},
		{
			name:        "error when slug ends with hyphen",
			id:          "",
			attrName:    "Test",
			slug:        "invalid-",
			attrType:    AttributeTypeSingle,
			wantErr:     true,
			errContains: "slug must contain only lowercase",
		},
		{
			name:        "error when invalid attribute type",
			id:          "",
			attrName:    "Test",
			slug:        "test",
			attrType:    AttributeType("invalid"),
			wantErr:     true,
			errContains: "invalid attribute type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr, err := NewAttribute(
				tt.id,
				tt.attrName,
				tt.slug,
				tt.attrType,
				tt.unit,
				tt.enabled,
				tt.options,
			)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidAttributeData)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, attr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, attr)
				if tt.id == "" {
					assert.NotEmpty(t, attr.ID)
				} else {
					assert.Equal(t, tt.id, attr.ID)
				}
				assert.Equal(t, 1, attr.Version)
				assert.Equal(t, tt.attrName, attr.Name)
				assert.Equal(t, tt.slug, attr.Slug)
				assert.Equal(t, tt.attrType, attr.Type)
				assert.Equal(t, tt.unit, attr.Unit)
				assert.Equal(t, tt.enabled, attr.Enabled)
				assert.Equal(t, tt.options, attr.Options)
				assert.False(t, attr.CreatedAt.IsZero())
				assert.False(t, attr.ModifiedAt.IsZero())
			}
		})
	}
}

func TestNewAttribute_ValidSlugs(t *testing.T) {
	validSlugs := []string{
		"color",
		"size",
		"is-organic",
		"weight-kg",
		"size123",
		"a",
		"a1",
		"abc-123-def",
	}

	for _, slug := range validSlugs {
		t.Run("valid slug: "+slug, func(t *testing.T) {
			attr, err := NewAttribute("", "Test", slug, AttributeTypeSingle, nil, true, nil)
			require.NoError(t, err)
			assert.Equal(t, slug, attr.Slug)
		})
	}
}

func TestAttribute_Update(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *Attribute
		newName     string
		unit        *string
		enabled     bool
		options     []Option
		wantErr     bool
		errContains string
	}{
		{
			name: "successful update",
			setup: func() *Attribute {
				a, _ := NewAttribute("", "Original", "original", AttributeTypeSingle, nil, false, nil)
				return a
			},
			newName: "Updated Name",
			unit:    ptr("new-unit"),
			enabled: true,
			options: []Option{
				{Name: "Option 1", Slug: "option-1", SortOrder: 1},
			},
			wantErr: false,
		},
		{
			name: "error when updating with empty name",
			setup: func() *Attribute {
				a, _ := NewAttribute("", "Original", "original", AttributeTypeSingle, nil, false, nil)
				return a
			},
			newName:     "",
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name: "error when updating with too long name",
			setup: func() *Attribute {
				a, _ := NewAttribute("", "Original", "original", AttributeTypeSingle, nil, false, nil)
				return a
			},
			newName:     strings.Repeat("a", 101),
			wantErr:     true,
			errContains: "name is too long",
		},
		{
			name: "slug and type remain unchanged after update",
			setup: func() *Attribute {
				a, _ := NewAttribute("", "Original", "original-slug", AttributeTypeRange, nil, false, nil)
				return a
			},
			newName: "New Name",
			unit:    nil,
			enabled: true,
			options: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := tt.setup()
			originalSlug := attr.Slug
			originalType := attr.Type
			originalModifiedAt := attr.ModifiedAt

			time.Sleep(time.Millisecond)

			err := attr.Update(tt.newName, tt.unit, tt.enabled, tt.options)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidAttributeData)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.newName, attr.Name)
				assert.Equal(t, tt.unit, attr.Unit)
				assert.Equal(t, tt.enabled, attr.Enabled)
				assert.Equal(t, tt.options, attr.Options)
				// Slug and Type should not change
				assert.Equal(t, originalSlug, attr.Slug)
				assert.Equal(t, originalType, attr.Type)
				assert.True(t, attr.ModifiedAt.After(originalModifiedAt))
			}
		})
	}
}

func TestValidateOptions(t *testing.T) {
	tests := []struct {
		name        string
		options     []Option
		wantErr     bool
		errContains string
	}{
		{
			name:    "nil options",
			options: nil,
			wantErr: false,
		},
		{
			name:    "empty options slice",
			options: []Option{},
			wantErr: false,
		},
		{
			name: "valid options",
			options: []Option{
				{Name: "Red", Slug: "red", SortOrder: 1},
				{Name: "Blue", Slug: "blue", SortOrder: 2},
			},
			wantErr: false,
		},
		{
			name: "valid option with color code",
			options: []Option{
				{Name: "Red", Slug: "red", ColorCode: ptr("#FF0000"), SortOrder: 1},
			},
			wantErr: false,
		},
		{
			name: "error when option name is empty",
			options: []Option{
				{Name: "", Slug: "test", SortOrder: 1},
			},
			wantErr:     true,
			errContains: "option name is required",
		},
		{
			name: "error when option name is too long",
			options: []Option{
				{Name: strings.Repeat("a", 101), Slug: "test", SortOrder: 1},
			},
			wantErr:     true,
			errContains: "option name is too long",
		},
		{
			name: "error when option slug is empty",
			options: []Option{
				{Name: "Test", Slug: "", SortOrder: 1},
			},
			wantErr:     true,
			errContains: "option slug is required",
		},
		{
			name: "error when option slug is too long",
			options: []Option{
				{Name: "Test", Slug: strings.Repeat("a", 51), SortOrder: 1},
			},
			wantErr:     true,
			errContains: "option slug is too long",
		},
		{
			name: "error when option slug has invalid format",
			options: []Option{
				{Name: "Test", Slug: "Invalid-Slug", SortOrder: 1},
			},
			wantErr:     true,
			errContains: "option slug must contain only lowercase",
		},
		{
			name: "error when duplicate option slugs",
			options: []Option{
				{Name: "Red", Slug: "color", SortOrder: 1},
				{Name: "Blue", Slug: "color", SortOrder: 2},
			},
			wantErr:     true,
			errContains: "duplicate option slug",
		},
		{
			name: "error when option sortOrder is negative",
			options: []Option{
				{Name: "Test", Slug: "test", SortOrder: -1},
			},
			wantErr:     true,
			errContains: "option sortOrder cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOptions(tt.options)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestReconstruct(t *testing.T) {
	t.Run("reconstructs attribute without validation", func(t *testing.T) {
		createdAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		modifiedAt := time.Date(2025, 6, 15, 12, 30, 0, 0, time.UTC)
		options := []Option{
			{Name: "Red", Slug: "red", SortOrder: 1},
		}

		// Reconstruct should not validate - it's for rebuilding from persistence
		attr := Reconstruct(
			"attr-123",
			5,
			"",                       // Empty name would fail validation
			"INVALID",                // Invalid slug would fail validation
			AttributeType("unknown"), // Invalid type would fail validation
			ptr("unit"),
			true,
			options,
			createdAt,
			modifiedAt,
		)

		require.NotNil(t, attr)
		assert.Equal(t, "attr-123", attr.ID)
		assert.Equal(t, 5, attr.Version)
		assert.Equal(t, "", attr.Name)
		assert.Equal(t, "INVALID", attr.Slug)
		assert.Equal(t, AttributeType("unknown"), attr.Type)
		assert.Equal(t, ptr("unit"), attr.Unit)
		assert.True(t, attr.Enabled)
		assert.Equal(t, options, attr.Options)
		assert.Equal(t, createdAt, attr.CreatedAt)
		assert.Equal(t, modifiedAt, attr.ModifiedAt)
	})
}

func TestAttributeTypeConstants(t *testing.T) {
	assert.Equal(t, AttributeType("single"), AttributeTypeSingle)
	assert.Equal(t, AttributeType("multiple"), AttributeTypeMultiple)
	assert.Equal(t, AttributeType("range"), AttributeTypeRange)
	assert.Equal(t, AttributeType("boolean"), AttributeTypeBoolean)
	assert.Equal(t, AttributeType("text"), AttributeTypeText)
}

func TestIsValidAttributeType(t *testing.T) {
	validTypes := []AttributeType{
		AttributeTypeSingle,
		AttributeTypeMultiple,
		AttributeTypeRange,
		AttributeTypeBoolean,
		AttributeTypeText,
	}

	for _, attrType := range validTypes {
		t.Run("valid type: "+string(attrType), func(t *testing.T) {
			assert.True(t, isValidAttributeType(attrType))
		})
	}

	invalidTypes := []AttributeType{
		"",
		"invalid",
		"SINGLE",
		"Single",
	}

	for _, attrType := range invalidTypes {
		t.Run("invalid type: "+string(attrType), func(t *testing.T) {
			assert.False(t, isValidAttributeType(attrType))
		})
	}
}

func TestOption(t *testing.T) {
	opt := Option{
		Name:      "Red",
		Slug:      "red",
		ColorCode: ptr("#FF0000"),
		SortOrder: 1,
	}

	assert.Equal(t, "Red", opt.Name)
	assert.Equal(t, "red", opt.Slug)
	assert.Equal(t, "#FF0000", *opt.ColorCode)
	assert.Equal(t, 1, opt.SortOrder)
}
