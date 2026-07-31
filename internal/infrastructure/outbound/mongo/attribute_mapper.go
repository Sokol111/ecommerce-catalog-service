package mongo

import (
	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-catalog-service/internal/application/attribute"
)

type AttributeMapper struct{}

func NewAttributeMapper() *AttributeMapper {
	return &AttributeMapper{}
}

func (m *AttributeMapper) ToEntity(a *attribute.Attribute) *AttributeEntity {
	options := lo.Map(a.Options, func(opt attribute.Option, _ int) OptionEntity {
		return OptionEntity{
			Name:      opt.Name,
			Slug:      opt.Slug,
			ColorCode: opt.ColorCode,
			SortOrder: opt.SortOrder,
		}
	})

	return &AttributeEntity{
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

func (m *AttributeMapper) ToDomain(e *AttributeEntity) *attribute.Attribute {
	options := lo.Map(e.Options, func(opt OptionEntity, _ int) attribute.Option {
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

func (m *AttributeMapper) GetID(e *AttributeEntity) string {
	return e.ID
}

func (m *AttributeMapper) GetVersion(e *AttributeEntity) int64 {
	return e.Version
}

func (m *AttributeMapper) SetVersion(e *AttributeEntity, version int64) {
	e.Version = version
}
