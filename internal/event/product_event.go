package event

import (
	"context"

	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/events"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/product"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
)

// ProductEventFactory creates product events
type ProductEventFactory interface {
	NewProductUpdatedOutboxMessage(ctx context.Context, p *product.Product) outbox.Message
}

type productEventFactory struct{}

// newProductEventFactory creates a new ProductEventFactory
func newProductEventFactory() ProductEventFactory {
	return &productEventFactory{}
}

func toProductEventAttributes(productAttrs []product.AttributeValue) *[]events.AttributeValue {
	if len(productAttrs) == 0 {
		return nil
	}

	result := lo.Map(productAttrs, func(pAttr product.AttributeValue, _ int) events.AttributeValue {
		return events.AttributeValue{
			AttributeID:      pAttr.AttributeID,
			OptionSlugValue:  pAttr.OptionSlugValue,
			OptionSlugValues: lo.ToPtr(pAttr.OptionSlugValues),
			NumericValue:     pAttr.NumericValue,
			TextValue:        pAttr.TextValue,
			BooleanValue:     pAttr.BooleanValue,
		}
	})

	return &result
}

func (f *productEventFactory) newProductUpdatedEvent(p *product.Product) *events.ProductUpdatedEvent {
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
			Attributes:  toProductEventAttributes(p.Attributes),
		},
	}
}

func (f *productEventFactory) NewProductUpdatedOutboxMessage(ctx context.Context, p *product.Product) outbox.Message {
	return outbox.Message{
		Event: f.newProductUpdatedEvent(p),
		Key:   p.ID,
	}
}
