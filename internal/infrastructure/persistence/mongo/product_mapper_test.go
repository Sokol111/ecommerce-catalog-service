package mongo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/product"
)

func ptrFloat32(v float32) *float32 {
	return &v
}

func ptrBool(v bool) *bool {
	return &v
}

func TestProductMapper_ToEntity(t *testing.T) {
	mapper := newProductMapper()

	t.Run("maps all fields correctly", func(t *testing.T) {
		now := time.Now().UTC()
		domainProduct := product.Reconstruct(
			"prod-123",
			2,
			"iPhone 15 Pro",
			ptr("Latest iPhone model"),
			999.99,
			50,
			ptr("image-123"),
			ptr("category-phones"),
			true,
			[]product.AttributeValue{
				{
					AttributeID:     "attr-color",
					OptionSlugValue: ptr("black"),
				},
				{
					AttributeID:      "attr-storage",
					OptionSlugValues: []string{"128gb", "256gb", "512gb"},
				},
				{
					AttributeID:  "attr-weight",
					NumericValue: ptrFloat32(187.5),
				},
			},
			now,
			now,
		)

		entity := mapper.ToEntity(domainProduct)

		require.NotNil(t, entity)
		assert.Equal(t, "prod-123", entity.ID)
		assert.Equal(t, 2, entity.Version)
		assert.Equal(t, "iPhone 15 Pro", entity.Name)
		assert.Equal(t, ptr("Latest iPhone model"), entity.Description)
		assert.Equal(t, float32(999.99), entity.Price)
		assert.Equal(t, 50, entity.Quantity)
		assert.Equal(t, ptr("image-123"), entity.ImageID)
		assert.Equal(t, ptr("category-phones"), entity.CategoryID)
		assert.True(t, entity.Enabled)
		assert.Equal(t, now, entity.CreatedAt)
		assert.Equal(t, now, entity.ModifiedAt)

		require.Len(t, entity.Attributes, 3)
		assert.Equal(t, "attr-color", entity.Attributes[0].AttributeID)
		assert.Equal(t, ptr("black"), entity.Attributes[0].OptionSlugValue)

		assert.Equal(t, "attr-storage", entity.Attributes[1].AttributeID)
		assert.Equal(t, []string{"128gb", "256gb", "512gb"}, entity.Attributes[1].OptionSlugValues)

		assert.Equal(t, "attr-weight", entity.Attributes[2].AttributeID)
		assert.Equal(t, ptrFloat32(187.5), entity.Attributes[2].NumericValue)
	})

	t.Run("maps product without optional fields", func(t *testing.T) {
		now := time.Now().UTC()
		domainProduct := product.Reconstruct(
			"prod-456",
			1,
			"Simple Product",
			nil,
			10.0,
			100,
			nil,
			nil,
			false,
			nil,
			now,
			now,
		)

		entity := mapper.ToEntity(domainProduct)

		require.NotNil(t, entity)
		assert.Equal(t, "prod-456", entity.ID)
		assert.Equal(t, "Simple Product", entity.Name)
		assert.Nil(t, entity.Description)
		assert.Nil(t, entity.ImageID)
		assert.Nil(t, entity.CategoryID)
		assert.False(t, entity.Enabled)
		assert.Nil(t, entity.Attributes)
	})

	t.Run("maps all attribute types", func(t *testing.T) {
		now := time.Now().UTC()
		domainProduct := product.Reconstruct(
			"prod-789",
			1,
			"Test Product",
			nil,
			50.0,
			10,
			nil,
			nil,
			true,
			[]product.AttributeValue{
				{AttributeID: "single", OptionSlugValue: ptr("option-1")},
				{AttributeID: "multiple", OptionSlugValues: []string{"opt-a", "opt-b"}},
				{AttributeID: "numeric", NumericValue: ptrFloat32(42.5)},
				{AttributeID: "text", TextValue: ptr("Some text value")},
				{AttributeID: "boolean", BooleanValue: ptrBool(true)},
			},
			now,
			now,
		)

		entity := mapper.ToEntity(domainProduct)

		require.Len(t, entity.Attributes, 5)
		assert.Equal(t, ptr("option-1"), entity.Attributes[0].OptionSlugValue)
		assert.Equal(t, []string{"opt-a", "opt-b"}, entity.Attributes[1].OptionSlugValues)
		assert.Equal(t, ptrFloat32(42.5), entity.Attributes[2].NumericValue)
		assert.Equal(t, ptr("Some text value"), entity.Attributes[3].TextValue)
		assert.Equal(t, ptrBool(true), entity.Attributes[4].BooleanValue)
	})
}

func TestProductMapper_ToDomain(t *testing.T) {
	mapper := newProductMapper()

	t.Run("maps all fields correctly", func(t *testing.T) {
		now := time.Now().UTC()
		entity := &productEntity{
			ID:          "prod-123",
			Version:     5,
			Name:        "MacBook Pro",
			Description: ptr("Professional laptop"),
			Price:       2499.99,
			Quantity:    25,
			ImageID:     ptr("img-macbook"),
			CategoryID:  ptr("cat-laptops"),
			Enabled:     true,
			Attributes: []productAttributeEntity{
				{AttributeID: "attr-cpu", OptionSlugValue: ptr("m3-pro")},
				{AttributeID: "attr-ram", OptionSlugValues: []string{"16gb", "32gb"}},
				{AttributeID: "attr-screen", NumericValue: ptrFloat32(14.2)},
			},
			CreatedAt:  now,
			ModifiedAt: now,
		}

		domain := mapper.ToDomain(entity)

		require.NotNil(t, domain)
		assert.Equal(t, "prod-123", domain.ID)
		assert.Equal(t, 5, domain.Version)
		assert.Equal(t, "MacBook Pro", domain.Name)
		assert.Equal(t, ptr("Professional laptop"), domain.Description)
		assert.Equal(t, float32(2499.99), domain.Price)
		assert.Equal(t, 25, domain.Quantity)
		assert.Equal(t, ptr("img-macbook"), domain.ImageID)
		assert.Equal(t, ptr("cat-laptops"), domain.CategoryID)
		assert.True(t, domain.Enabled)
		assert.Equal(t, now.UTC(), domain.CreatedAt)
		assert.Equal(t, now.UTC(), domain.ModifiedAt)

		require.Len(t, domain.Attributes, 3)
		assert.Equal(t, "attr-cpu", domain.Attributes[0].AttributeID)
		assert.Equal(t, ptr("m3-pro"), domain.Attributes[0].OptionSlugValue)
	})

	t.Run("maps entity without optional fields", func(t *testing.T) {
		now := time.Now().UTC()
		entity := &productEntity{
			ID:         "prod-456",
			Version:    1,
			Name:       "Basic Product",
			Price:      5.0,
			Quantity:   1000,
			Enabled:    false,
			CreatedAt:  now,
			ModifiedAt: now,
		}

		domain := mapper.ToDomain(entity)

		require.NotNil(t, domain)
		assert.Equal(t, "prod-456", domain.ID)
		assert.Nil(t, domain.Description)
		assert.Nil(t, domain.ImageID)
		assert.Nil(t, domain.CategoryID)
		assert.Nil(t, domain.Attributes)
	})

	t.Run("converts time to UTC", func(t *testing.T) {
		loc, _ := time.LoadLocation("Asia/Tokyo")
		localTime := time.Date(2024, 3, 20, 9, 0, 0, 0, loc)

		entity := &productEntity{
			ID:         "prod-789",
			Version:    1,
			Name:       "Test",
			Price:      10.0,
			Quantity:   1,
			Enabled:    true,
			CreatedAt:  localTime,
			ModifiedAt: localTime,
		}

		domain := mapper.ToDomain(entity)

		assert.Equal(t, time.UTC, domain.CreatedAt.Location())
		assert.Equal(t, time.UTC, domain.ModifiedAt.Location())
	})
}

func TestProductMapper_GetID(t *testing.T) {
	mapper := newProductMapper()

	entity := &productEntity{ID: "product-id-xyz"}

	assert.Equal(t, "product-id-xyz", mapper.GetID(entity))
}

func TestProductMapper_GetVersion(t *testing.T) {
	mapper := newProductMapper()

	entity := &productEntity{Version: 12}

	assert.Equal(t, 12, mapper.GetVersion(entity))
}

func TestProductMapper_SetVersion(t *testing.T) {
	mapper := newProductMapper()

	entity := &productEntity{Version: 1}

	mapper.SetVersion(entity, 20)

	assert.Equal(t, 20, entity.Version)
}

func TestProductMapper_RoundTrip(t *testing.T) {
	mapper := newProductMapper()

	t.Run("domain -> entity -> domain preserves all data", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Millisecond)
		original := product.Reconstruct(
			"prod-roundtrip",
			3,
			"Samsung Galaxy S24",
			ptr("Flagship smartphone"),
			899.99,
			100,
			ptr("img-galaxy"),
			ptr("cat-smartphones"),
			true,
			[]product.AttributeValue{
				{AttributeID: "color", OptionSlugValue: ptr("phantom-black")},
				{AttributeID: "storage", OptionSlugValues: []string{"256gb", "512gb"}},
				{AttributeID: "weight", NumericValue: ptrFloat32(168.0)},
				{AttributeID: "notes", TextValue: ptr("Includes charger")},
				{AttributeID: "5g", BooleanValue: ptrBool(true)},
			},
			now,
			now,
		)

		entity := mapper.ToEntity(original)
		restored := mapper.ToDomain(entity)

		assert.Equal(t, original.ID, restored.ID)
		assert.Equal(t, original.Version, restored.Version)
		assert.Equal(t, original.Name, restored.Name)
		assert.Equal(t, original.Description, restored.Description)
		assert.Equal(t, original.Price, restored.Price)
		assert.Equal(t, original.Quantity, restored.Quantity)
		assert.Equal(t, original.ImageID, restored.ImageID)
		assert.Equal(t, original.CategoryID, restored.CategoryID)
		assert.Equal(t, original.Enabled, restored.Enabled)
		assert.Equal(t, original.CreatedAt, restored.CreatedAt)
		assert.Equal(t, original.ModifiedAt, restored.ModifiedAt)

		require.Len(t, restored.Attributes, len(original.Attributes))
		for i, attr := range original.Attributes {
			assert.Equal(t, attr.AttributeID, restored.Attributes[i].AttributeID)
			assert.Equal(t, attr.OptionSlugValue, restored.Attributes[i].OptionSlugValue)
			assert.Equal(t, attr.OptionSlugValues, restored.Attributes[i].OptionSlugValues)
			assert.Equal(t, attr.NumericValue, restored.Attributes[i].NumericValue)
			assert.Equal(t, attr.TextValue, restored.Attributes[i].TextValue)
			assert.Equal(t, attr.BooleanValue, restored.Attributes[i].BooleanValue)
		}
	})
}

func TestMapProductAttributeToEntity(t *testing.T) {
	t.Run("maps single option value", func(t *testing.T) {
		attr := product.AttributeValue{
			AttributeID:     "attr-1",
			OptionSlugValue: ptr("red"),
		}

		entity := mapProductAttributeToEntity(attr, 0)

		assert.Equal(t, "attr-1", entity.AttributeID)
		assert.Equal(t, ptr("red"), entity.OptionSlugValue)
		assert.Nil(t, entity.OptionSlugValues)
	})

	t.Run("maps multiple option values", func(t *testing.T) {
		attr := product.AttributeValue{
			AttributeID:      "attr-2",
			OptionSlugValues: []string{"small", "medium", "large"},
		}

		entity := mapProductAttributeToEntity(attr, 0)

		assert.Equal(t, "attr-2", entity.AttributeID)
		assert.Nil(t, entity.OptionSlugValue)
		assert.Equal(t, []string{"small", "medium", "large"}, entity.OptionSlugValues)
	})

	t.Run("maps all value types", func(t *testing.T) {
		attr := product.AttributeValue{
			AttributeID:      "attr-3",
			OptionSlugValue:  ptr("opt"),
			OptionSlugValues: []string{"a", "b"},
			NumericValue:     ptrFloat32(99.9),
			TextValue:        ptr("text"),
			BooleanValue:     ptrBool(false),
		}

		entity := mapProductAttributeToEntity(attr, 0)

		assert.Equal(t, ptr("opt"), entity.OptionSlugValue)
		assert.Equal(t, []string{"a", "b"}, entity.OptionSlugValues)
		assert.Equal(t, ptrFloat32(99.9), entity.NumericValue)
		assert.Equal(t, ptr("text"), entity.TextValue)
		assert.Equal(t, ptrBool(false), entity.BooleanValue)
	})
}

func TestMapProductAttributeToDomain(t *testing.T) {
	t.Run("maps single option value", func(t *testing.T) {
		entity := productAttributeEntity{
			AttributeID:     "attr-1",
			OptionSlugValue: ptr("blue"),
		}

		attr := mapProductAttributeToDomain(entity, 0)

		assert.Equal(t, "attr-1", attr.AttributeID)
		assert.Equal(t, ptr("blue"), attr.OptionSlugValue)
		assert.Nil(t, attr.OptionSlugValues)
	})

	t.Run("maps multiple option values", func(t *testing.T) {
		entity := productAttributeEntity{
			AttributeID:      "attr-2",
			OptionSlugValues: []string{"xs", "s", "m", "l", "xl"},
		}

		attr := mapProductAttributeToDomain(entity, 0)

		assert.Equal(t, "attr-2", attr.AttributeID)
		assert.Equal(t, []string{"xs", "s", "m", "l", "xl"}, attr.OptionSlugValues)
	})

	t.Run("maps all value types", func(t *testing.T) {
		entity := productAttributeEntity{
			AttributeID:      "attr-3",
			OptionSlugValue:  ptr("value"),
			OptionSlugValues: []string{"x", "y"},
			NumericValue:     ptrFloat32(123.45),
			TextValue:        ptr("description"),
			BooleanValue:     ptrBool(true),
		}

		attr := mapProductAttributeToDomain(entity, 0)

		assert.Equal(t, ptr("value"), attr.OptionSlugValue)
		assert.Equal(t, []string{"x", "y"}, attr.OptionSlugValues)
		assert.Equal(t, ptrFloat32(123.45), attr.NumericValue)
		assert.Equal(t, ptr("description"), attr.TextValue)
		assert.Equal(t, ptrBool(true), attr.BooleanValue)
	})
}
