package event

import (
	"context"
	"time"

	"github.com/samber/lo"

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

// toCategoryEventAttributes converts category attributes to event attributes
// Only immutable references and category-specific settings are included
func toCategoryEventAttributes(categoryAttrs []category.CategoryAttribute) []events.CategoryAttribute {
	return lo.Map(categoryAttrs, func(catAttr category.CategoryAttribute, _ int) events.CategoryAttribute {
		return events.CategoryAttribute{
			AttributeID:   catAttr.AttributeID,
			AttributeSlug: catAttr.Slug,
			Role:          string(catAttr.Role),
			Required:      catAttr.Required,
			SortOrder:     catAttr.SortOrder,
			Filterable:    catAttr.Filterable,
			Searchable:    catAttr.Searchable,
		}
	})
}

func (f *categoryEventFactory) newCategoryCreatedEvent(ctx context.Context, c *category.Category) *events.CategoryCreatedEvent {
	traceId := tracing.GetTraceID(ctx)

	return &events.CategoryCreatedEvent{
		Metadata: commonsevents.EventMetadata{
			EventID:   uuid.New().String(),
			EventType: events.EventTypeCategoryCreated,
			Source:    "ecommerce-catalog-service",
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
			Source:    "ecommerce-catalog-service",
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
