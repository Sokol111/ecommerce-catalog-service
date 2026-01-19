package event

import (
	"context"
	"time"

	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/events"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/category"
	commonsevents "github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/events"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/observability/tracing"
	"github.com/google/uuid"
)

// CategoryEventFactory creates category events
type CategoryEventFactory interface {
	NewCategoryCreatedOutboxMessage(ctx context.Context, c *category.Category, attrs []*attribute.Attribute) (outbox.Message, error)
	NewCategoryUpdatedOutboxMessage(ctx context.Context, c *category.Category, attrs []*attribute.Attribute) (outbox.Message, error)
}

type categoryEventFactory struct{}

// newCategoryEventFactory creates a new CategoryEventFactory
func newCategoryEventFactory() CategoryEventFactory {
	return &categoryEventFactory{}
}

func toCategoryEventAttributes(categoryAttrs []category.CategoryAttribute, attrs []*attribute.Attribute) []events.CategoryAttribute {
	if len(categoryAttrs) == 0 {
		return []events.CategoryAttribute{}
	}

	attrMap := lo.KeyBy(attrs, func(attr *attribute.Attribute) string {
		return attr.ID
	})

	return lo.Map(categoryAttrs, func(catAttr category.CategoryAttribute, _ int) events.CategoryAttribute {
		attr := attrMap[catAttr.AttributeID]

		return events.CategoryAttribute{
			AttributeID:      catAttr.AttributeID,
			AttributeName:    attr.Name,
			AttributeSlug:    attr.Slug,
			AttributeType:    string(attr.Type),
			AttributeUnit:    attr.Unit,
			AttributeOptions: toEventAttributeOptions(attr.Options),
			Role:             string(catAttr.Role),
			Required:         catAttr.Required,
			SortOrder:        catAttr.SortOrder,
			Filterable:       catAttr.Filterable,
			Searchable:       catAttr.Searchable,
		}
	})
}

func toEventAttributeOptions(options []attribute.Option) []events.AttributeOption {
	if len(options) == 0 {
		return []events.AttributeOption{}
	}

	return lo.Map(options, func(opt attribute.Option, _ int) events.AttributeOption {
		return events.AttributeOption{
			Name:      opt.Name,
			Slug:      opt.Slug,
			ColorCode: opt.ColorCode,
			SortOrder: opt.SortOrder,
			Enabled:   opt.Enabled,
		}
	})
}

func (f *categoryEventFactory) newCategoryCreatedEvent(ctx context.Context, c *category.Category, attrs []*attribute.Attribute) *events.CategoryCreatedEvent {
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
			Attributes: toCategoryEventAttributes(c.Attributes, attrs),
			Version:    c.Version,
			CreatedAt:  c.CreatedAt,
			ModifiedAt: c.ModifiedAt,
		},
	}
}

func (f *categoryEventFactory) newCategoryUpdatedEvent(ctx context.Context, c *category.Category, attrs []*attribute.Attribute) *events.CategoryUpdatedEvent {
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
			Attributes: toCategoryEventAttributes(c.Attributes, attrs),
			Version:    c.Version,
			CreatedAt:  c.CreatedAt,
			ModifiedAt: c.ModifiedAt,
		},
	}
}

func (f *categoryEventFactory) NewCategoryCreatedOutboxMessage(ctx context.Context, c *category.Category, attrs []*attribute.Attribute) (outbox.Message, error) {
	e := f.newCategoryCreatedEvent(ctx, c, attrs)

	return outbox.Message{
		Payload: e,
		EventID: e.Metadata.EventID,
		Key:     c.ID,
	}, nil
}

func (f *categoryEventFactory) NewCategoryUpdatedOutboxMessage(ctx context.Context, c *category.Category, attrs []*attribute.Attribute) (outbox.Message, error) {
	e := f.newCategoryUpdatedEvent(ctx, c, attrs)

	return outbox.Message{
		Payload: e,
		EventID: e.Metadata.EventID,
		Key:     c.ID,
	}, nil
}
