package mongo

import (
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/product"
)

type productMapper struct{}

func newProductMapper() *productMapper {
	return &productMapper{}
}

func (m *productMapper) ToEntity(p *product.Product) *productEntity {
	return &productEntity{
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

func (m *productMapper) ToDomain(e *productEntity) *product.Product {
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

func (m *productMapper) GetID(e *productEntity) string {
	return e.ID
}

func (m *productMapper) GetVersion(e *productEntity) int {
	return e.Version
}

func (m *productMapper) SetVersion(e *productEntity, version int) {
	e.Version = version
}

func (m *productMapper) attributesToEntities(attrs []product.AttributeValue) []productAttributeEntity {
	if attrs == nil {
		return nil
	}

	result := make([]productAttributeEntity, len(attrs))
	for i, attr := range attrs {
		result[i] = productAttributeEntity{
			AttributeID:      attr.AttributeID,
			OptionSlugValue:  attr.OptionSlugValue,
			OptionSlugValues: attr.OptionSlugValues,
			NumericValue:     attr.NumericValue,
			TextValue:        attr.TextValue,
			BooleanValue:     attr.BooleanValue,
		}
	}
	return result
}

func (m *productMapper) attributesToDomain(entities []productAttributeEntity) []product.AttributeValue {
	if entities == nil {
		return nil
	}

	result := make([]product.AttributeValue, len(entities))
	for i, e := range entities {
		result[i] = product.AttributeValue{
			AttributeID:      e.AttributeID,
			OptionSlugValue:  e.OptionSlugValue,
			OptionSlugValues: e.OptionSlugValues,
			NumericValue:     e.NumericValue,
			TextValue:        e.TextValue,
			BooleanValue:     e.BooleanValue,
		}
	}
	return result
}
