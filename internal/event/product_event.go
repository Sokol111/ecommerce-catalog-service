package event

import (
	"context"
	"time"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/events"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/product"
	commonsevents "github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/events"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/observability/tracing"
	"github.com/google/uuid"
)

// ProductEventFactory creates product events
type ProductEventFactory interface {
	NewProductCreatedOutboxMessage(ctx context.Context, p *product.Product) (outbox.Message, error)
	NewProductUpdatedOutboxMessage(ctx context.Context, p *product.Product) (outbox.Message, error)
}

type productEventFactory struct{}

// newProductEventFactory creates a new ProductEventFactory
func newProductEventFactory() ProductEventFactory {
	return &productEventFactory{}
}

func toProductEventAttributes(attrs []product.ProductAttribute) *[]events.ProductAttribute {
	if len(attrs) == 0 {
		return nil
	}

	result := make([]events.ProductAttribute, len(attrs))
	for i, attr := range attrs {
		result[i] = events.ProductAttribute{
			AttributeID:      attr.AttributeID,
			OptionSlugValue:  attr.OptionSlugValue,
			OptionSlugValues: toStringSlicePtr(attr.OptionSlugValues),
			NumericValue:     attr.NumericValue,
			TextValue:        attr.TextValue,
			BooleanValue:     attr.BooleanValue,
		}
	}
	return &result
}

func toStringSlicePtr(s []string) *[]string {
	if s == nil {
		return nil
	}
	return &s
}

func (f *productEventFactory) newProductCreatedEvent(ctx context.Context, p *product.Product) *events.ProductCreatedEvent {
	traceId := tracing.GetTraceID(ctx)

	return &events.ProductCreatedEvent{
		Metadata: commonsevents.EventMetadata{
			EventID:   uuid.New().String(),
			EventType: events.EventTypeProductCreated,
			Source:    "ecommerce-product-service",
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
			Attributes:  toProductEventAttributes(p.Attributes),
		},
	}
}

func (f *productEventFactory) newProductUpdatedEvent(ctx context.Context, p *product.Product) *events.ProductUpdatedEvent {
	traceId := tracing.GetTraceID(ctx)

	return &events.ProductUpdatedEvent{
		Metadata: commonsevents.EventMetadata{
			EventID:   uuid.New().String(),
			EventType: events.EventTypeProductUpdated,
			Source:    "ecommerce-product-service",
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
			Attributes:  toProductEventAttributes(p.Attributes),
		},
	}
}

func (f *productEventFactory) NewProductCreatedOutboxMessage(ctx context.Context, p *product.Product) (outbox.Message, error) {
	e := f.newProductCreatedEvent(ctx, p)

	return outbox.Message{
		Payload: e,
		EventID: e.Metadata.EventID,
		Key:     p.ID,
	}, nil
}

func (f *productEventFactory) NewProductUpdatedOutboxMessage(ctx context.Context, p *product.Product) (outbox.Message, error) {
	e := f.newProductUpdatedEvent(ctx, p)

	return outbox.Message{
		Payload: e,
		EventID: e.Metadata.EventID,
		Key:     p.ID,
	}, nil
}
