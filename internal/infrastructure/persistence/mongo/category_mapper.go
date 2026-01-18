package mongo

import (
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/category"
)

type categoryMapper struct{}

func newCategoryMapper() *categoryMapper {
	return &categoryMapper{}
}

func (m *categoryMapper) ToEntity(c *category.Category) *categoryEntity {
	attributes := make([]categoryAttributeEntity, 0, len(c.Attributes))
	for _, attr := range c.Attributes {
		attributes = append(attributes, categoryAttributeEntity{
			AttributeID: attr.AttributeID,
			Role:        string(attr.Role),
			Required:    attr.Required,
			SortOrder:   attr.SortOrder,
			Filterable:  attr.Filterable,
			Searchable:  attr.Searchable,
		})
	}

	return &categoryEntity{
		ID:         c.ID,
		Version:    c.Version,
		Name:       c.Name,
		Enabled:    c.Enabled,
		Attributes: attributes,
		CreatedAt:  c.CreatedAt,
		ModifiedAt: c.ModifiedAt,
	}
}

func (m *categoryMapper) ToDomain(e *categoryEntity) *category.Category {
	attributes := make([]category.CategoryAttribute, 0, len(e.Attributes))
	for _, attr := range e.Attributes {
		attributes = append(attributes, category.CategoryAttribute{
			AttributeID: attr.AttributeID,
			Role:        category.AttributeRole(attr.Role),
			Required:    attr.Required,
			SortOrder:   attr.SortOrder,
			Filterable:  attr.Filterable,
			Searchable:  attr.Searchable,
		})
	}

	return category.Reconstruct(
		e.ID,
		e.Version,
		e.Name,
		e.Enabled,
		attributes,
		e.CreatedAt.UTC(),
		e.ModifiedAt.UTC(),
	)
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
