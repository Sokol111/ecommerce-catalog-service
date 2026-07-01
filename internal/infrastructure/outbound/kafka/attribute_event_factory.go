package kafka

import (
	"context"

	eventsv1 "github.com/Sokol111/ecommerce-catalog-service-api/gen/events/catalog/v1"
	apiEvents "github.com/Sokol111/ecommerce-catalog-service-api/pkg/events"
	"github.com/Sokol111/ecommerce-catalog-service/internal/application/attribute"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type attributeEventFactory struct{}

// newAttributeEventFactory creates a new AttributeEventFactory
func newAttributeEventFactory() attribute.AttributeEventFactory {
	return &attributeEventFactory{}
}

func toAttributeType(t attribute.AttributeType) eventsv1.AttributeType {
	switch t {
	case attribute.AttributeTypeSingle:
		return eventsv1.AttributeType_ATTRIBUTE_TYPE_SINGLE
	case attribute.AttributeTypeMultiple:
		return eventsv1.AttributeType_ATTRIBUTE_TYPE_MULTIPLE
	case attribute.AttributeTypeRange:
		return eventsv1.AttributeType_ATTRIBUTE_TYPE_RANGE
	case attribute.AttributeTypeBoolean:
		return eventsv1.AttributeType_ATTRIBUTE_TYPE_BOOLEAN
	case attribute.AttributeTypeText:
		return eventsv1.AttributeType_ATTRIBUTE_TYPE_TEXT
	default:
		return eventsv1.AttributeType_ATTRIBUTE_TYPE_UNSPECIFIED
	}
}

func toEventOptions(options []attribute.Option) []*eventsv1.AttributeOption {
	return lo.Map(options, func(opt attribute.Option, _ int) *eventsv1.AttributeOption {
		return &eventsv1.AttributeOption{
			Slug:      opt.Slug,
			Name:      opt.Name,
			ColorCode: opt.ColorCode,
			SortOrder: int32(opt.SortOrder),
		}
	})
}

func (f *attributeEventFactory) newAttributeUpdatedEvent(a *attribute.Attribute) *eventsv1.AttributeUpdatedEvent {
	return &eventsv1.AttributeUpdatedEvent{
		AttributeId: a.ID,
		Slug:        a.Slug,
		Name:        a.Name,
		Type:        toAttributeType(a.Type),
		Unit:        a.Unit,
		Enabled:     a.Enabled,
		Version:     int64(a.Version),
		ModifiedAt:  timestamppb.New(a.ModifiedAt),
		Options:     toEventOptions(a.Options),
	}
}

func (f *attributeEventFactory) NewAttributeUpdatedOutboxMessage(ctx context.Context, a *attribute.Attribute) outbox.Message {
	event := f.newAttributeUpdatedEvent(a)
	return outbox.Message{
		Event: event,
		Key:   a.ID,
		Topic: apiEvents.TopicFor(event),
	}
}
