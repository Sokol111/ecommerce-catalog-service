package event

import (
	"context"

	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/events"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/category"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
)

// CategoryEventFactory creates category events
type CategoryEventFactory interface {
	NewCategoryCreatedOutboxMessage(ctx context.Context, c *category.Category) outbox.Message
	NewCategoryUpdatedOutboxMessage(ctx context.Context, c *category.Category) outbox.Message
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

func (f *categoryEventFactory) newCategoryCreatedEvent(c *category.Category) *events.CategoryCreatedEvent {
	return &events.CategoryCreatedEvent{
		// Metadata is populated automatically by outbox
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

func (f *categoryEventFactory) newCategoryUpdatedEvent(c *category.Category) *events.CategoryUpdatedEvent {
	return &events.CategoryUpdatedEvent{
		// Metadata is populated automatically by outbox
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

func (f *categoryEventFactory) NewCategoryCreatedOutboxMessage(ctx context.Context, c *category.Category) outbox.Message {
	return outbox.Message{
		Event: f.newCategoryCreatedEvent(c),
		Key:   c.ID,
	}
}

func (f *categoryEventFactory) NewCategoryUpdatedOutboxMessage(ctx context.Context, c *category.Category) outbox.Message {
	return outbox.Message{
		Event: f.newCategoryUpdatedEvent(c),
		Key:   c.ID,
	}
}
