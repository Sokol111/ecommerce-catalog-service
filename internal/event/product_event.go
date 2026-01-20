package event

import (
	"context"
	"time"

	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/events"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/category"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/product"
	commonsevents "github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/events"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/observability/tracing"
	"github.com/google/uuid"
)

// ProductEventFactory creates product events
type ProductEventFactory interface {
	NewProductCreatedOutboxMessage(ctx context.Context, p *product.Product, attrs []*attribute.Attribute, cat *category.Category) (outbox.Message, error)
	NewProductUpdatedOutboxMessage(ctx context.Context, p *product.Product, attrs []*attribute.Attribute, cat *category.Category) (outbox.Message, error)
}

type productEventFactory struct{}

// newProductEventFactory creates a new ProductEventFactory
func newProductEventFactory() ProductEventFactory {
	return &productEventFactory{}
}

func toProductEventAttributes(productAttrs []product.ProductAttribute, attrs []*attribute.Attribute, cat *category.Category) *[]events.ProductAttribute {
	if len(productAttrs) == 0 {
		return nil
	}

	// Build attribute map for quick lookup
	attrMap := lo.KeyBy(attrs, func(a *attribute.Attribute) string {
		return a.ID
	})

	// Build category attribute map for role and sortOrder
	var catAttrMap map[string]category.CategoryAttribute
	if cat != nil {
		catAttrMap = lo.KeyBy(cat.Attributes, func(ca category.CategoryAttribute) string {
			return ca.AttributeID
		})
	}

	result := lo.Map(productAttrs, func(pAttr product.ProductAttribute, _ int) events.ProductAttribute {
		eventAttr := events.ProductAttribute{
			AttributeID:      pAttr.AttributeID,
			OptionSlugValue:  pAttr.OptionSlugValue,
			OptionSlugValues: lo.ToPtr(pAttr.OptionSlugValues),
			NumericValue:     pAttr.NumericValue,
			TextValue:        pAttr.TextValue,
			BooleanValue:     pAttr.BooleanValue,
		}

		// Enrich from Attribute
		if attr, ok := attrMap[pAttr.AttributeID]; ok {
			eventAttr.AttributeSlug = attr.Slug
			eventAttr.AttributeName = attr.Name
			eventAttr.AttributeType = string(attr.Type)
			eventAttr.AttributeUnit = attr.Unit

			// Find option details for single type
			if pAttr.OptionSlugValue != nil {
				if opt, found := lo.Find(attr.Options, func(o attribute.Option) bool {
					return o.Slug == *pAttr.OptionSlugValue
				}); found {
					eventAttr.OptionName = &opt.Name
					eventAttr.OptionColorCode = opt.ColorCode
				}
			}

			// Find option details for multiple type
			if len(pAttr.OptionSlugValues) > 0 {
				optionMap := lo.KeyBy(attr.Options, func(o attribute.Option) string {
					return o.Slug
				})
				names := lo.FilterMap(pAttr.OptionSlugValues, func(slug string, _ int) (string, bool) {
					if opt, exists := optionMap[slug]; exists {
						return opt.Name, true
					}
					return "", false
				})
				if len(names) > 0 {
					eventAttr.OptionNames = &names
				}
			}
		}

		// Enrich from CategoryAttribute
		if catAttr, ok := catAttrMap[pAttr.AttributeID]; ok {
			eventAttr.Role = string(catAttr.Role)
			eventAttr.SortOrder = catAttr.SortOrder
		}

		return eventAttr
	})
	return &result
}

func (f *productEventFactory) newProductCreatedEvent(ctx context.Context, p *product.Product, attrs []*attribute.Attribute, cat *category.Category) *events.ProductCreatedEvent {
	traceId := tracing.GetTraceID(ctx)

	return &events.ProductCreatedEvent{
		Metadata: commonsevents.EventMetadata{
			EventID:   uuid.New().String(),
			EventType: events.EventTypeProductCreated,
			Source:    "ecommerce-catalog-service",
			Timestamp: time.Now().UTC(),
			TraceID:   &traceId,
		},
		Payload: events.ProductCreatedPayload{
			ProductID:   p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    p.Quantity,
			Enabled:     p.Enabled,
			Version:     p.Version,
			ImageID:     p.ImageID,
			CategoryID:  p.CategoryID,
			CreatedAt:   p.CreatedAt,
			ModifiedAt:  p.ModifiedAt,
			Attributes:  toProductEventAttributes(p.Attributes, attrs, cat),
		},
	}
}

func (f *productEventFactory) newProductUpdatedEvent(ctx context.Context, p *product.Product, attrs []*attribute.Attribute, cat *category.Category) *events.ProductUpdatedEvent {
	traceId := tracing.GetTraceID(ctx)

	return &events.ProductUpdatedEvent{
		Metadata: commonsevents.EventMetadata{
			EventID:   uuid.New().String(),
			EventType: events.EventTypeProductUpdated,
			Source:    "ecommerce-catalog-service",
			Timestamp: time.Now().UTC(),
			TraceID:   &traceId,
		},
		Payload: events.ProductUpdatedPayload{
			ProductID:   p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    p.Quantity,
			Enabled:     p.Enabled,
			Version:     p.Version,
			ImageID:     p.ImageID,
			CategoryID:  p.CategoryID,
			CreatedAt:   p.CreatedAt,
			ModifiedAt:  p.ModifiedAt,
			Attributes:  toProductEventAttributes(p.Attributes, attrs, cat),
		},
	}
}

func (f *productEventFactory) NewProductCreatedOutboxMessage(ctx context.Context, p *product.Product, attrs []*attribute.Attribute, cat *category.Category) (outbox.Message, error) {
	e := f.newProductCreatedEvent(ctx, p, attrs, cat)

	return outbox.Message{
		Payload: e,
		EventID: e.Metadata.EventID,
		Key:     p.ID,
	}, nil
}

func (f *productEventFactory) NewProductUpdatedOutboxMessage(ctx context.Context, p *product.Product, attrs []*attribute.Attribute, cat *category.Category) (outbox.Message, error) {
	e := f.newProductUpdatedEvent(ctx, p, attrs, cat)

	return outbox.Message{
		Payload: e,
		EventID: e.Metadata.EventID,
		Key:     p.ID,
	}, nil
}
