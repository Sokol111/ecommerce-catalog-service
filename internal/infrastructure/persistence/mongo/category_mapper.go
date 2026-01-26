package mongo

import (
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/category"
	"github.com/samber/lo"
)

type categoryMapper struct{}

func newCategoryMapper() *categoryMapper {
	return &categoryMapper{}
}

func (m *categoryMapper) ToEntity(c *category.Category) *categoryEntity {
	return &categoryEntity{
		ID:         c.ID,
		Version:    c.Version,
		Name:       c.Name,
		Enabled:    c.Enabled,
		Attributes: m.attributesToEntities(c.Attributes),
		CreatedAt:  c.CreatedAt,
		ModifiedAt: c.ModifiedAt,
	}
}

func (m *categoryMapper) ToDomain(e *categoryEntity) *category.Category {
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

func (m *categoryMapper) attributesToEntities(attrs []category.CategoryAttribute) []categoryAttributeEntity {
	if attrs == nil {
		return nil
	}

	return lo.Map(attrs, mapCategoryAttributeToEntity)
}

func mapCategoryAttributeToEntity(attr category.CategoryAttribute, _ int) categoryAttributeEntity {
	return categoryAttributeEntity{
		AttributeID: attr.AttributeID,
		Role:        string(attr.Role),
		Required:    attr.Required,
		SortOrder:   attr.SortOrder,
		Filterable:  attr.Filterable,
		Searchable:  attr.Searchable,
	}
}

func (m *categoryMapper) attributesToDomain(entities []categoryAttributeEntity) []category.CategoryAttribute {
	if entities == nil {
		return nil
	}

	return lo.Map(entities, mapCategoryAttributeToDomain)
}

func mapCategoryAttributeToDomain(attr categoryAttributeEntity, _ int) category.CategoryAttribute {
	return category.CategoryAttribute{
		AttributeID: attr.AttributeID,
		Role:        category.AttributeRole(attr.Role),
		Required:    attr.Required,
		SortOrder:   attr.SortOrder,
		Filterable:  attr.Filterable,
		Searchable:  attr.Searchable,
	}
}

func (m *categoryMapper) GetID(e *categoryEntity) string {
	return e.ID
}

func (m *categoryMapper) GetVersion(e *categoryEntity) int {
	return e.Version
}

func (m *categoryMapper) SetVersion(e *categoryEntity, version int) {
	e.Version = version
}
