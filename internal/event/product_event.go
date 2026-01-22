package event

import (
	"context"

	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/events"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/product"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
)

// ProductEventFactory creates product events
type ProductEventFactory interface {
	NewProductCreatedOutboxMessage(ctx context.Context, p *product.Product, attrs []*attribute.Attribute) outbox.Message
	NewProductUpdatedOutboxMessage(ctx context.Context, p *product.Product, attrs []*attribute.Attribute) outbox.Message
}

type productEventFactory struct{}

// newProductEventFactory creates a new ProductEventFactory
func newProductEventFactory() ProductEventFactory {
	return &productEventFactory{}
}

// toProductEventAttributes converts product attributes to event attributes
// Only immutable references (IDs, slugs) and product-specific values are included
func toProductEventAttributes(productAttrs []product.AttributeValue, attrs []*attribute.Attribute) *[]events.AttributeValue {
	if len(productAttrs) == 0 {
		return nil
	}

	attrMap := lo.KeyBy(attrs, func(a *attribute.Attribute) string {
		return a.ID
	})

	result := lo.Map(productAttrs, func(pAttr product.AttributeValue, _ int) events.AttributeValue {
		eventAttr := events.AttributeValue{
			AttributeID:      pAttr.AttributeID,
			OptionSlugValue:  pAttr.OptionSlugValue,
			OptionSlugValues: lo.ToPtr(pAttr.OptionSlugValues),
			NumericValue:     pAttr.NumericValue,
			TextValue:        pAttr.TextValue,
			BooleanValue:     pAttr.BooleanValue,
		}

		// Get attribute slug (immutable)
		if attr, ok := attrMap[pAttr.AttributeID]; ok {
			eventAttr.AttributeSlug = attr.Slug
		}

		return eventAttr
	})

	return &result
}

func (f *productEventFactory) newProductCreatedEvent(p *product.Product, attrs []*attribute.Attribute) *events.ProductCreatedEvent {
	return &events.ProductCreatedEvent{
		// Metadata is populated automatically by outbox
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
			Attributes:  toProductEventAttributes(p.Attributes, attrs),
		},
	}
}

func (f *productEventFactory) newProductUpdatedEvent(p *product.Product, attrs []*attribute.Attribute) *events.ProductUpdatedEvent {
	return &events.ProductUpdatedEvent{
		// Metadata is populated automatically by outbox
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
			Attributes:  toProductEventAttributes(p.Attributes, attrs),
		},
	}
}

func (f *productEventFactory) NewProductCreatedOutboxMessage(ctx context.Context, p *product.Product, attrs []*attribute.Attribute) outbox.Message {
	return outbox.Message{
		Event: f.newProductCreatedEvent(p, attrs),
		Key:   p.ID,
	}
}

func (f *productEventFactory) NewProductUpdatedOutboxMessage(ctx context.Context, p *product.Product, attrs []*attribute.Attribute) outbox.Message {
	return outbox.Message{
		Event: f.newProductUpdatedEvent(p, attrs),
		Key:   p.ID,
	}
}
