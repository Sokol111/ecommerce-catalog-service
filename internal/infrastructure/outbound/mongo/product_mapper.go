package mongo

import (
	"github.com/Sokol111/ecommerce-catalog-service/internal/application/product"
	"github.com/samber/lo"
)

type ProductMapper struct{}

func NewProductMapper() *ProductMapper {
	return &ProductMapper{}
}

func (m *ProductMapper) ToEntity(p *product.Product) *ProductEntity {
	return &ProductEntity{
		ID:          p.ID,
		Version:     p.Version,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Quantity:    p.Quantity,
		ImageID:     p.ImageID,
		CategoryID:  p.CategoryID,
		Enabled:     p.Enabled,
		Attributes:  m.attributesToEntities(p.Attributes),
		CreatedAt:   p.CreatedAt,
		ModifiedAt:  p.ModifiedAt,
	}
}

func (m *ProductMapper) ToDomain(e *ProductEntity) *product.Product {
	return product.Reconstruct(
		e.ID,
		e.Version,
		e.Name,
		e.Description,
		e.Price,
		e.Quantity,
		e.ImageID,
		e.CategoryID,
		e.Enabled,
		m.attributesToDomain(e.Attributes),
		e.CreatedAt.UTC(),
		e.ModifiedAt.UTC(),
	)
}

func (m *ProductMapper) GetID(e *ProductEntity) string {
	return e.ID
}

func (m *ProductMapper) GetVersion(e *ProductEntity) int64 {
	return e.Version
}

func (m *ProductMapper) SetVersion(e *ProductEntity, version int64) {
	e.Version = version
}

func (m *ProductMapper) attributesToEntities(attrs []product.AttributeValue) []ProductAttributeEntity {
	if attrs == nil {
		return nil
	}

	return lo.Map(attrs, mapProductAttributeToEntity)
}

func mapProductAttributeToEntity(attr product.AttributeValue, _ int) ProductAttributeEntity {
	return ProductAttributeEntity{
		AttributeID:      attr.AttributeID,
		AttributeSlug:    attr.AttributeSlug,
		OptionSlugValue:  attr.OptionSlugValue,
		OptionSlugValues: attr.OptionSlugValues,
		NumericValue:     attr.NumericValue,
		TextValue:        attr.TextValue,
		BooleanValue:     attr.BooleanValue,
	}
}

func (m *ProductMapper) attributesToDomain(entities []ProductAttributeEntity) []product.AttributeValue {
	if entities == nil {
		return nil
	}

	return lo.Map(entities, mapProductAttributeToDomain)
}

func mapProductAttributeToDomain(e ProductAttributeEntity, _ int) product.AttributeValue {
	return product.AttributeValue{
		AttributeID:      e.AttributeID,
		AttributeSlug:    e.AttributeSlug,
		OptionSlugValue:  e.OptionSlugValue,
		OptionSlugValues: e.OptionSlugValues,
		NumericValue:     e.NumericValue,
		TextValue:        e.TextValue,
		BooleanValue:     e.BooleanValue,
	}
}
