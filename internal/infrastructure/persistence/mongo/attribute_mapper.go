package mongo

import (
	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
)

type attributeMapper struct{}

func newAttributeMapper() *attributeMapper {
	return &attributeMapper{}
}

func (m *attributeMapper) ToEntity(a *attribute.Attribute) *attributeEntity {
	options := lo.Map(a.Options, func(opt attribute.Option, _ int) optionEntity {
		return optionEntity{
			Name:      opt.Name,
			Slug:      opt.Slug,
			ColorCode: opt.ColorCode,
			SortOrder: opt.SortOrder,
		}
	})

	return &attributeEntity{
		ID:         a.ID,
		Version:    a.Version,
		Name:       a.Name,
		Slug:       a.Slug,
		Type:       string(a.Type),
		Unit:       a.Unit,
		Enabled:    a.Enabled,
		Options:    options,
		CreatedAt:  a.CreatedAt,
		ModifiedAt: a.ModifiedAt,
	}
}

func (m *attributeMapper) ToDomain(e *attributeEntity) *attribute.Attribute {
	options := lo.Map(e.Options, func(opt optionEntity, _ int) attribute.Option {
		return attribute.Option{
			Name:      opt.Name,
			Slug:      opt.Slug,
			ColorCode: opt.ColorCode,
			SortOrder: opt.SortOrder,
		}
	})

	return attribute.Reconstruct(
		e.ID,
		e.Version,
		e.Name,
		e.Slug,
		attribute.AttributeType(e.Type),
		e.Unit,
		e.Enabled,
		options,
		e.CreatedAt.UTC(),
		e.ModifiedAt.UTC(),
	)
}

func (m *attributeMapper) GetID(e *attributeEntity) string {
	return e.ID
}

func (m *attributeMapper) GetVersion(e *attributeEntity) int {
	return e.Version
}

func (m *attributeMapper) SetVersion(e *attributeEntity, version int) {
	e.Version = version
}
