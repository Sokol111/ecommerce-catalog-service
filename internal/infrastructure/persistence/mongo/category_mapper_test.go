package mongo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/category"
)

func TestCategoryMapper_ToEntity(t *testing.T) {
	mapper := newCategoryMapper()

	t.Run("maps all fields correctly", func(t *testing.T) {
		now := time.Now().UTC()
		domainCategory := category.Reconstruct(
			"cat-123",
			2,
			"Electronics",
			true,
			[]category.CategoryAttribute{
				{
					AttributeID: "attr-1",
					Slug:        "color",
					Role:        category.AttributeRoleVariant,
					Required:    true,
					SortOrder:   1,
					Filterable:  true,
					Searchable:  true,
				},
				{
					AttributeID: "attr-2",
					Slug:        "size",
					Role:        category.AttributeRoleSpecification,
					Required:    false,
					SortOrder:   2,
					Filterable:  true,
					Searchable:  false,
				},
			},
			now,
			now,
		)

		entity := mapper.ToEntity(domainCategory)

		require.NotNil(t, entity)
		assert.Equal(t, "cat-123", entity.ID)
		assert.Equal(t, 2, entity.Version)
		assert.Equal(t, "Electronics", entity.Name)
		assert.True(t, entity.Enabled)
		assert.Equal(t, now, entity.CreatedAt)
		assert.Equal(t, now, entity.ModifiedAt)

		require.Len(t, entity.Attributes, 2)
		assert.Equal(t, "attr-1", entity.Attributes[0].AttributeID)
		assert.Equal(t, "variant", entity.Attributes[0].Role)
		assert.True(t, entity.Attributes[0].Required)
		assert.Equal(t, 1, entity.Attributes[0].SortOrder)
		assert.True(t, entity.Attributes[0].Filterable)
		assert.True(t, entity.Attributes[0].Searchable)

		assert.Equal(t, "attr-2", entity.Attributes[1].AttributeID)
		assert.Equal(t, "specification", entity.Attributes[1].Role)
		assert.False(t, entity.Attributes[1].Required)
	})

	t.Run("maps category without attributes", func(t *testing.T) {
		now := time.Now().UTC()
		domainCategory := category.Reconstruct(
			"cat-456",
			1,
			"Books",
			false,
			nil,
			now,
			now,
		)

		entity := mapper.ToEntity(domainCategory)

		require.NotNil(t, entity)
		assert.Equal(t, "cat-456", entity.ID)
		assert.Equal(t, "Books", entity.Name)
		assert.False(t, entity.Enabled)
		assert.Nil(t, entity.Attributes)
	})

	t.Run("maps category with empty attributes slice", func(t *testing.T) {
		now := time.Now().UTC()
		domainCategory := category.Reconstruct(
			"cat-789",
			1,
			"Clothing",
			true,
			[]category.CategoryAttribute{},
			now,
			now,
		)

		entity := mapper.ToEntity(domainCategory)

		require.NotNil(t, entity)
		assert.Empty(t, entity.Attributes)
	})
}

func TestCategoryMapper_ToDomain(t *testing.T) {
	mapper := newCategoryMapper()

	t.Run("maps all fields correctly", func(t *testing.T) {
		now := time.Now().UTC()
		entity := &categoryEntity{
			ID:      "cat-123",
			Version: 3,
			Name:    "Home & Garden",
			Enabled: true,
			Attributes: []categoryAttributeEntity{
				{
					AttributeID: "attr-10",
					Role:        "variant",
					Required:    true,
					SortOrder:   1,
					Filterable:  true,
					Searchable:  true,
				},
				{
					AttributeID: "attr-20",
					Role:        "specification",
					Required:    false,
					SortOrder:   2,
					Filterable:  false,
					Searchable:  true,
				},
			},
			CreatedAt:  now,
			ModifiedAt: now,
		}

		domain := mapper.ToDomain(entity)

		require.NotNil(t, domain)
		assert.Equal(t, "cat-123", domain.ID)
		assert.Equal(t, 3, domain.Version)
		assert.Equal(t, "Home & Garden", domain.Name)
		assert.True(t, domain.Enabled)
		assert.Equal(t, now.UTC(), domain.CreatedAt)
		assert.Equal(t, now.UTC(), domain.ModifiedAt)

		require.Len(t, domain.Attributes, 2)
		assert.Equal(t, "attr-10", domain.Attributes[0].AttributeID)
		assert.Equal(t, category.AttributeRoleVariant, domain.Attributes[0].Role)
		assert.True(t, domain.Attributes[0].Required)
		assert.Equal(t, 1, domain.Attributes[0].SortOrder)
		assert.True(t, domain.Attributes[0].Filterable)
		assert.True(t, domain.Attributes[0].Searchable)

		assert.Equal(t, "attr-20", domain.Attributes[1].AttributeID)
		assert.Equal(t, category.AttributeRoleSpecification, domain.Attributes[1].Role)
	})

	t.Run("maps entity without attributes", func(t *testing.T) {
		now := time.Now().UTC()
		entity := &categoryEntity{
			ID:         "cat-456",
			Version:    1,
			Name:       "Sports",
			Enabled:    false,
			Attributes: nil,
			CreatedAt:  now,
			ModifiedAt: now,
		}

		domain := mapper.ToDomain(entity)

		require.NotNil(t, domain)
		assert.Equal(t, "cat-456", domain.ID)
		assert.Nil(t, domain.Attributes)
	})

	t.Run("converts time to UTC", func(t *testing.T) {
		loc, _ := time.LoadLocation("Europe/Berlin")
		localTime := time.Date(2024, 6, 15, 14, 30, 0, 0, loc)

		entity := &categoryEntity{
			ID:         "cat-789",
			Version:    1,
			Name:       "Test",
			Enabled:    true,
			Attributes: nil,
			CreatedAt:  localTime,
			ModifiedAt: localTime,
		}

		domain := mapper.ToDomain(entity)

		assert.Equal(t, time.UTC, domain.CreatedAt.Location())
		assert.Equal(t, time.UTC, domain.ModifiedAt.Location())
	})
}

func TestCategoryMapper_GetID(t *testing.T) {
	mapper := newCategoryMapper()

	entity := &categoryEntity{ID: "category-id-123"}

	assert.Equal(t, "category-id-123", mapper.GetID(entity))
}

func TestCategoryMapper_GetVersion(t *testing.T) {
	mapper := newCategoryMapper()

	entity := &categoryEntity{Version: 8}

	assert.Equal(t, 8, mapper.GetVersion(entity))
}

func TestCategoryMapper_SetVersion(t *testing.T) {
	mapper := newCategoryMapper()

	entity := &categoryEntity{Version: 1}

	mapper.SetVersion(entity, 15)

	assert.Equal(t, 15, entity.Version)
}

func TestCategoryMapper_RoundTrip(t *testing.T) {
	mapper := newCategoryMapper()

	t.Run("domain -> entity -> domain preserves all data", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Millisecond)
		original := category.Reconstruct(
			"cat-roundtrip",
			5,
			"Automotive",
			true,
			[]category.CategoryAttribute{
				{
					AttributeID: "attr-brand",
					Slug:        "brand",
					Role:        category.AttributeRoleVariant,
					Required:    true,
					SortOrder:   1,
					Filterable:  true,
					Searchable:  true,
				},
				{
					AttributeID: "attr-model",
					Slug:        "model",
					Role:        category.AttributeRoleSpecification,
					Required:    false,
					SortOrder:   2,
					Filterable:  false,
					Searchable:  true,
				},
			},
			now,
			now,
		)

		entity := mapper.ToEntity(original)
		restored := mapper.ToDomain(entity)

		assert.Equal(t, original.ID, restored.ID)
		assert.Equal(t, original.Version, restored.Version)
		assert.Equal(t, original.Name, restored.Name)
		assert.Equal(t, original.Enabled, restored.Enabled)
		assert.Equal(t, original.CreatedAt, restored.CreatedAt)
		assert.Equal(t, original.ModifiedAt, restored.ModifiedAt)

		require.Len(t, restored.Attributes, len(original.Attributes))
		for i, attr := range original.Attributes {
			assert.Equal(t, attr.AttributeID, restored.Attributes[i].AttributeID)
			assert.Equal(t, attr.Role, restored.Attributes[i].Role)
			assert.Equal(t, attr.Required, restored.Attributes[i].Required)
			assert.Equal(t, attr.SortOrder, restored.Attributes[i].SortOrder)
			assert.Equal(t, attr.Filterable, restored.Attributes[i].Filterable)
			assert.Equal(t, attr.Searchable, restored.Attributes[i].Searchable)
		}
	})
}

func TestMapCategoryAttributeToEntity(t *testing.T) {
	attr := category.CategoryAttribute{
		AttributeID: "attr-123",
		Slug:        "color",
		Role:        category.AttributeRoleVariant,
		Required:    true,
		SortOrder:   5,
		Filterable:  true,
		Searchable:  false,
	}

	entity := mapCategoryAttributeToEntity(attr, 0)

	assert.Equal(t, "attr-123", entity.AttributeID)
	assert.Equal(t, "variant", entity.Role)
	assert.True(t, entity.Required)
	assert.Equal(t, 5, entity.SortOrder)
	assert.True(t, entity.Filterable)
	assert.False(t, entity.Searchable)
}

func TestMapCategoryAttributeToDomain(t *testing.T) {
	entity := categoryAttributeEntity{
		AttributeID: "attr-456",
		Role:        "specification",
		Required:    false,
		SortOrder:   10,
		Filterable:  false,
		Searchable:  true,
	}

	attr := mapCategoryAttributeToDomain(entity, 0)

	assert.Equal(t, "attr-456", attr.AttributeID)
	assert.Equal(t, category.AttributeRoleSpecification, attr.Role)
	assert.False(t, attr.Required)
	assert.Equal(t, 10, attr.SortOrder)
	assert.False(t, attr.Filterable)
	assert.True(t, attr.Searchable)
}
