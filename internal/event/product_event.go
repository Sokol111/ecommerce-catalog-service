package event

import (
	"context"
	"time"

	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/events"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/product"
	commonsevents "github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/events"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/observability/tracing"
	"github.com/google/uuid"
)

// ProductEventFactory creates product events
type ProductEventFactory interface {
	NewProductCreatedOutboxMessage(ctx context.Context, p *product.Product, attrs []*attribute.Attribute) (outbox.Message, error)
	NewProductUpdatedOutboxMessage(ctx context.Context, p *product.Product, attrs []*attribute.Attribute) (outbox.Message, error)
}

type productEventFactory struct{}

// newProductEventFactory creates a new ProductEventFactory
func newProductEventFactory() ProductEventFactory {
	return &productEventFactory{}
}

// toProductEventAttributes converts product attributes to event attributes
// Only immutable references (IDs, slugs) and product-specific values are included
func toProductEventAttributes(productAttrs []product.ProductAttribute, attrs []*attribute.Attribute) *[]events.ProductAttribute {
	if len(productAttrs) == 0 {
		return nil
	}

	attrMap := lo.KeyBy(attrs, func(a *attribute.Attribute) string {
		return a.ID
	})

	result := lo.Map(productAttrs, func(pAttr product.ProductAttribute, _ int) events.ProductAttribute {
		eventAttr := events.ProductAttribute{
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

func (f *productEventFactory) newProductCreatedEvent(ctx context.Context, p *product.Product, attrs []*attribute.Attribute) *events.ProductCreatedEvent {
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
			Attributes:  toProductEventAttributes(p.Attributes, attrs),
		},
	}
}

func (f *productEventFactory) newProductUpdatedEvent(ctx context.Context, p *product.Product, attrs []*attribute.Attribute) *events.ProductUpdatedEvent {
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
			Attributes:  toProductEventAttributes(p.Attributes, attrs),
		},
	}
}

func (f *productEventFactory) NewProductCreatedOutboxMessage(ctx context.Context, p *product.Product, attrs []*attribute.Attribute) (outbox.Message, error) {
	e := f.newProductCreatedEvent(ctx, p, attrs)

	return outbox.Message{
		Payload: e,
		EventID: e.Metadata.EventID,
		Key:     p.ID,
	}, nil
}

func (f *productEventFactory) NewProductUpdatedOutboxMessage(ctx context.Context, p *product.Product, attrs []*attribute.Attribute) (outbox.Message, error) {
	e := f.newProductUpdatedEvent(ctx, p, attrs)

	return outbox.Message{
		Payload: e,
		EventID: e.Metadata.EventID,
		Key:     p.ID,
	}, nil
}
