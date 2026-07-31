package mongo

import (
	"github.com/Sokol111/ecommerce-catalog-service/internal/application/category"
	"github.com/samber/lo"
)

type CategoryMapper struct{}

func NewCategoryMapper() *CategoryMapper {
	return &CategoryMapper{}
}

func (m *CategoryMapper) ToEntity(c *category.Category) *CategoryEntity {
	return &CategoryEntity{
		ID:         c.ID,
		Version:    c.Version,
		Name:       c.Name,
		Enabled:    c.Enabled,
		Attributes: m.attributesToEntities(c.Attributes),
		CreatedAt:  c.CreatedAt,
		ModifiedAt: c.ModifiedAt,
	}
}

func (m *CategoryMapper) ToDomain(e *CategoryEntity) *category.Category {
	return category.Reconstruct(
		e.ID,
		e.Version,
		e.Name,
		e.Enabled,
		m.attributesToDomain(e.Attributes),
		e.CreatedAt.UTC(),
		e.ModifiedAt.UTC(),
	)
}

func (m *CategoryMapper) attributesToEntities(attrs []category.CategoryAttribute) []CategoryAttributeEntity {
	if attrs == nil {
		return nil
	}

	return lo.Map(attrs, mapCategoryAttributeToEntity)
}

func mapCategoryAttributeToEntity(attr category.CategoryAttribute, _ int) CategoryAttributeEntity {
	return CategoryAttributeEntity{
		AttributeID: attr.AttributeID,
		Slug:        attr.Slug,
		Role:        string(attr.Role),
		SortOrder:   attr.SortOrder,
		Filterable:  attr.Filterable,
		Searchable:  attr.Searchable,
	}
}

func (m *CategoryMapper) attributesToDomain(entities []CategoryAttributeEntity) []category.CategoryAttribute {
	if entities == nil {
		return nil
	}

	return lo.Map(entities, mapCategoryAttributeToDomain)
}

func mapCategoryAttributeToDomain(attr CategoryAttributeEntity, _ int) category.CategoryAttribute {
	return category.CategoryAttribute{
		AttributeID: attr.AttributeID,
		Slug:        attr.Slug,
		Role:        category.AttributeRole(attr.Role),
		SortOrder:   attr.SortOrder,
		Filterable:  attr.Filterable,
		Searchable:  attr.Searchable,
	}
}

func (m *CategoryMapper) GetID(e *CategoryEntity) string {
	return e.ID
}

func (m *CategoryMapper) GetVersion(e *CategoryEntity) int64 {
	return e.Version
}

func (m *CategoryMapper) SetVersion(e *CategoryEntity, version int64) {
	e.Version = version
}
