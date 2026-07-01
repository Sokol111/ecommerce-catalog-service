package kafka

import (
	"context"

	eventsv1 "github.com/Sokol111/ecommerce-catalog-service-api/gen/events/catalog/v1"
	apiEvents "github.com/Sokol111/ecommerce-catalog-service-api/pkg/events"
	"github.com/Sokol111/ecommerce-catalog-service/internal/application/category"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type categoryEventFactory struct{}

// newCategoryEventFactory creates a new CategoryEventFactory
func newCategoryEventFactory() category.CategoryEventFactory {
	return &categoryEventFactory{}
}

func toCategoryAttributeRole(r category.AttributeRole) eventsv1.CategoryAttributeRole {
	switch r {
	case category.AttributeRoleVariant:
		return eventsv1.CategoryAttributeRole_CATEGORY_ATTRIBUTE_ROLE_VARIANT
	case category.AttributeRoleSpecification:
		return eventsv1.CategoryAttributeRole_CATEGORY_ATTRIBUTE_ROLE_SPECIFICATION
	default:
		return eventsv1.CategoryAttributeRole_CATEGORY_ATTRIBUTE_ROLE_UNSPECIFIED
	}
}

// toCategoryEventAttributes converts category attributes to event attributes
// Only immutable references and category-specific settings are included
func toCategoryEventAttributes(categoryAttrs []category.CategoryAttribute) []*eventsv1.CategoryAttribute {
	return lo.Map(categoryAttrs, func(catAttr category.CategoryAttribute, _ int) *eventsv1.CategoryAttribute {
		return &eventsv1.CategoryAttribute{
			AttributeId:   catAttr.AttributeID,
			AttributeSlug: catAttr.Slug,
			Role:          toCategoryAttributeRole(catAttr.Role),
			SortOrder:     int32(catAttr.SortOrder),
			Filterable:    catAttr.Filterable,
			Searchable:    catAttr.Searchable,
		}
	})
}

func (f *categoryEventFactory) newCategoryUpdatedEvent(c *category.Category) *eventsv1.CategoryUpdatedEvent {
	return &eventsv1.CategoryUpdatedEvent{
		CategoryId: c.ID,
		Name:       c.Name,
		Enabled:    c.Enabled,
		Attributes: toCategoryEventAttributes(c.Attributes),
		Version:    int64(c.Version),
		CreatedAt:  timestamppb.New(c.CreatedAt),
		ModifiedAt: timestamppb.New(c.ModifiedAt),
	}
}

func (f *categoryEventFactory) NewCategoryUpdatedOutboxMessage(ctx context.Context, c *category.Category) outbox.Message {
	event := f.newCategoryUpdatedEvent(c)
	return outbox.Message{
		Event: event,
		Key:   c.ID,
		Topic: apiEvents.TopicFor(event),
	}
}
