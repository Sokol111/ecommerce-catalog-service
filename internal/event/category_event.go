package event

import (
	"context"
	"time"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/events"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/category"
	commonsevents "github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/events"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/observability/tracing"
	"github.com/google/uuid"
)

// CategoryEventFactory creates category events
type CategoryEventFactory interface {
	NewCategoryCreatedOutboxMessage(ctx context.Context, c *category.Category) (outbox.Message, error)
	NewCategoryUpdatedOutboxMessage(ctx context.Context, c *category.Category) (outbox.Message, error)
}

type categoryEventFactory struct{}

// newCategoryEventFactory creates a new CategoryEventFactory
func newCategoryEventFactory() CategoryEventFactory {
	return &categoryEventFactory{}
}

func toCategoryEventAttributes(attrs []category.CategoryAttribute) []events.CategoryAttribute {
	if len(attrs) == 0 {
		return []events.CategoryAttribute{}
	}

	result := make([]events.CategoryAttribute, 0, len(attrs))
	for _, attr := range attrs {
		result = append(result, events.CategoryAttribute{
			AttributeID: attr.AttributeID,
			Role:        string(attr.Role),
			Required:    attr.Required,
			SortOrder:   attr.SortOrder,
			Filterable:  attr.Filterable,
			Searchable:  attr.Searchable,
		})
	}
	return result
}

func (f *categoryEventFactory) newCategoryCreatedEvent(ctx context.Context, c *category.Category) *events.CategoryCreatedEvent {
	traceId := tracing.GetTraceID(ctx)

	return &events.CategoryCreatedEvent{
		Metadata: commonsevents.EventMetadata{
			EventID:   uuid.New().String(),
			EventType: events.EventTypeCategoryCreated,
			Source:    "ecommerce-category-service",
			Timestamp: time.Now().UTC(),
			TraceID:   &traceId,
		},
		Payload: events.CategoryCreatedPayload{
			CategoryID: c.ID,
			Name:       c.Name,
			Enabled:    c.Enabled,
			Attributes: toCategoryEventAttributes(c.Attributes),
			Version:    c.Version,
			CreatedAt:  c.CreatedAt,
			ModifiedAt: c.ModifiedAt,
		},
	}
}

func (f *categoryEventFactory) newCategoryUpdatedEvent(ctx context.Context, c *category.Category) *events.CategoryUpdatedEvent {
	traceId := tracing.GetTraceID(ctx)

	return &events.CategoryUpdatedEvent{
		Metadata: commonsevents.EventMetadata{
			EventID:   uuid.New().String(),
			EventType: events.EventTypeCategoryUpdated,
			Source:    "ecommerce-category-service",
			Timestamp: time.Now().UTC(),
			TraceID:   &traceId,
		},
		Payload: events.CategoryUpdatedPayload{
			CategoryID: c.ID,
			Name:       c.Name,
			Enabled:    c.Enabled,
			Attributes: toCategoryEventAttributes(c.Attributes),
			Version:    c.Version,
			CreatedAt:  c.CreatedAt,
			ModifiedAt: c.ModifiedAt,
		},
	}
}

func (f *categoryEventFactory) NewCategoryCreatedOutboxMessage(ctx context.Context, c *category.Category) (outbox.Message, error) {
	e := f.newCategoryCreatedEvent(ctx, c)

	return outbox.Message{
		Payload: e,
		EventID: e.Metadata.EventID,
		Key:     c.ID,
	}, nil
}

func (f *categoryEventFactory) NewCategoryUpdatedOutboxMessage(ctx context.Context, c *category.Category) (outbox.Message, error) {
	e := f.newCategoryUpdatedEvent(ctx, c)

	return outbox.Message{
		Payload: e,
		EventID: e.Metadata.EventID,
		Key:     c.ID,
	}, nil
}
