package mongo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
)

func ptr[T any](v T) *T {
	return &v
}

func TestAttributeMapper_ToEntity(t *testing.T) {
	mapper := newAttributeMapper()

	t.Run("maps all fields correctly", func(t *testing.T) {
		now := time.Now().UTC()
		domainAttr := attribute.Reconstruct(
			"attr-123",
			1,
			"Color",
			"color",
			attribute.AttributeTypeSingle,
			ptr("cm"),
			true,
			[]attribute.Option{
				{Name: "Red", Slug: "red", ColorCode: ptr("#FF0000"), SortOrder: 1},
				{Name: "Blue", Slug: "blue", ColorCode: ptr("#0000FF"), SortOrder: 2},
			},
			now,
			now,
		)

		entity := mapper.ToEntity(domainAttr)

		require.NotNil(t, entity)
		assert.Equal(t, "attr-123", entity.ID)
		assert.Equal(t, 1, entity.Version)
		assert.Equal(t, "Color", entity.Name)
		assert.Equal(t, "color", entity.Slug)
		assert.Equal(t, "single", entity.Type)
		assert.Equal(t, ptr("cm"), entity.Unit)
		assert.True(t, entity.Enabled)
		assert.Equal(t, now, entity.CreatedAt)
		assert.Equal(t, now, entity.ModifiedAt)

		require.Len(t, entity.Options, 2)
		assert.Equal(t, "Red", entity.Options[0].Name)
		assert.Equal(t, "red", entity.Options[0].Slug)
		assert.Equal(t, ptr("#FF0000"), entity.Options[0].ColorCode)
		assert.Equal(t, 1, entity.Options[0].SortOrder)
		assert.Equal(t, "Blue", entity.Options[1].Name)
		assert.Equal(t, "blue", entity.Options[1].Slug)
	})

	t.Run("maps attribute without options", func(t *testing.T) {
		now := time.Now().UTC()
		domainAttr := attribute.Reconstruct(
			"attr-456",
			2,
			"Weight",
			"weight",
			attribute.AttributeTypeRange,
			ptr("kg"),
			false,
			nil,
			now,
			now,
		)

		entity := mapper.ToEntity(domainAttr)

		require.NotNil(t, entity)
		assert.Equal(t, "attr-456", entity.ID)
		assert.Equal(t, "range", entity.Type)
		assert.Empty(t, entity.Options)
	})

	t.Run("maps attribute without unit", func(t *testing.T) {
		now := time.Now().UTC()
		domainAttr := attribute.Reconstruct(
			"attr-789",
			1,
			"Is Organic",
			"is-organic",
			attribute.AttributeTypeBoolean,
			nil,
			true,
			nil,
			now,
			now,
		)

		entity := mapper.ToEntity(domainAttr)

		require.NotNil(t, entity)
		assert.Nil(t, entity.Unit)
		assert.Equal(t, "boolean", entity.Type)
	})
}

func TestAttributeMapper_ToDomain(t *testing.T) {
	mapper := newAttributeMapper()

	t.Run("maps all fields correctly", func(t *testing.T) {
		now := time.Now().UTC()
		entity := &attributeEntity{
			ID:      "attr-123",
			Version: 3,
			Name:    "Size",
			Slug:    "size",
			Type:    "multiple",
			Unit:    ptr("cm"),
			Enabled: true,
			Options: []optionEntity{
				{Name: "Small", Slug: "small", ColorCode: nil, SortOrder: 1},
				{Name: "Medium", Slug: "medium", ColorCode: nil, SortOrder: 2},
				{Name: "Large", Slug: "large", ColorCode: nil, SortOrder: 3},
			},
			CreatedAt:  now,
			ModifiedAt: now,
		}

		domain := mapper.ToDomain(entity)

		require.NotNil(t, domain)
		assert.Equal(t, "attr-123", domain.ID)
		assert.Equal(t, 3, domain.Version)
		assert.Equal(t, "Size", domain.Name)
		assert.Equal(t, "size", domain.Slug)
		assert.Equal(t, attribute.AttributeTypeMultiple, domain.Type)
		assert.Equal(t, ptr("cm"), domain.Unit)
		assert.True(t, domain.Enabled)
		assert.Equal(t, now.UTC(), domain.CreatedAt)
		assert.Equal(t, now.UTC(), domain.ModifiedAt)

		require.Len(t, domain.Options, 3)
		assert.Equal(t, "Small", domain.Options[0].Name)
		assert.Equal(t, "small", domain.Options[0].Slug)
		assert.Equal(t, 1, domain.Options[0].SortOrder)
	})

	t.Run("maps entity without options", func(t *testing.T) {
		now := time.Now().UTC()
		entity := &attributeEntity{
			ID:         "attr-456",
			Version:    1,
			Name:       "Description",
			Slug:       "description",
			Type:       "text",
			Unit:       nil,
			Enabled:    false,
			Options:    nil,
			CreatedAt:  now,
			ModifiedAt: now,
		}

		domain := mapper.ToDomain(entity)

		require.NotNil(t, domain)
		assert.Equal(t, attribute.AttributeTypeText, domain.Type)
		assert.Nil(t, domain.Unit)
		assert.Empty(t, domain.Options)
	})

	t.Run("converts time to UTC", func(t *testing.T) {
		// Create time in a different timezone
		loc, _ := time.LoadLocation("America/New_York")
		localTime := time.Date(2024, 1, 15, 10, 30, 0, 0, loc)

		entity := &attributeEntity{
			ID:         "attr-789",
			Version:    1,
			Name:       "Test",
			Slug:       "test",
			Type:       "single",
			Enabled:    true,
			Options:    nil,
			CreatedAt:  localTime,
			ModifiedAt: localTime,
		}

		domain := mapper.ToDomain(entity)

		assert.Equal(t, time.UTC, domain.CreatedAt.Location())
		assert.Equal(t, time.UTC, domain.ModifiedAt.Location())
	})
}

func TestAttributeMapper_GetID(t *testing.T) {
	mapper := newAttributeMapper()

	entity := &attributeEntity{ID: "test-id-123"}

	assert.Equal(t, "test-id-123", mapper.GetID(entity))
}

func TestAttributeMapper_GetVersion(t *testing.T) {
	mapper := newAttributeMapper()

	entity := &attributeEntity{Version: 5}

	assert.Equal(t, 5, mapper.GetVersion(entity))
}

func TestAttributeMapper_SetVersion(t *testing.T) {
	mapper := newAttributeMapper()

	entity := &attributeEntity{Version: 1}

	mapper.SetVersion(entity, 10)

	assert.Equal(t, 10, entity.Version)
}

func TestAttributeMapper_RoundTrip(t *testing.T) {
	mapper := newAttributeMapper()

	t.Run("domain -> entity -> domain preserves all data", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Millisecond)
		original := attribute.Reconstruct(
			"attr-roundtrip",
			7,
			"Material",
			"material",
			attribute.AttributeTypeSingle,
			nil,
			true,
			[]attribute.Option{
				{Name: "Cotton", Slug: "cotton", ColorCode: nil, SortOrder: 1},
				{Name: "Polyester", Slug: "polyester", ColorCode: ptr("#123456"), SortOrder: 2},
			},
			now,
			now,
		)

		entity := mapper.ToEntity(original)
		restored := mapper.ToDomain(entity)

		assert.Equal(t, original.ID, restored.ID)
		assert.Equal(t, original.Version, restored.Version)
		assert.Equal(t, original.Name, restored.Name)
		assert.Equal(t, original.Slug, restored.Slug)
		assert.Equal(t, original.Type, restored.Type)
		assert.Equal(t, original.Unit, restored.Unit)
		assert.Equal(t, original.Enabled, restored.Enabled)
		assert.Equal(t, original.CreatedAt, restored.CreatedAt)
		assert.Equal(t, original.ModifiedAt, restored.ModifiedAt)

		require.Len(t, restored.Options, len(original.Options))
		for i, opt := range original.Options {
			assert.Equal(t, opt.Name, restored.Options[i].Name)
			assert.Equal(t, opt.Slug, restored.Options[i].Slug)
			assert.Equal(t, opt.ColorCode, restored.Options[i].ColorCode)
			assert.Equal(t, opt.SortOrder, restored.Options[i].SortOrder)
		}
	})
}
