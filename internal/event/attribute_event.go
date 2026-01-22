package event

import (
	"context"

	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/events"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
)

// AttributeEventFactory creates attribute events
type AttributeEventFactory interface {
	NewAttributeUpdatedOutboxMessage(ctx context.Context, a *attribute.Attribute) outbox.Message
}

type attributeEventFactory struct{}

// newAttributeEventFactory creates a new AttributeEventFactory
func newAttributeEventFactory() AttributeEventFactory {
	return &attributeEventFactory{}
}

func toEventOptions(options []attribute.Option) []events.AttributeOption {
	return lo.Map(options, func(opt attribute.Option, _ int) events.AttributeOption {
		return events.AttributeOption{
			Slug:      opt.Slug,
			Name:      opt.Name,
			ColorCode: opt.ColorCode,
			SortOrder: opt.SortOrder,
		}
	})
}

func (f *attributeEventFactory) newAttributeUpdatedEvent(a *attribute.Attribute) *events.AttributeUpdatedEvent {
	return &events.AttributeUpdatedEvent{
		// Metadata is populated automatically by outbox
		Payload: events.AttributeUpdatedPayload{
			AttributeID: a.ID,
			Slug:        a.Slug,
			Name:        a.Name,
			Type:        string(a.Type),
			Unit:        a.Unit,
			Enabled:     a.Enabled,
			Version:     a.Version,
			ModifiedAt:  a.ModifiedAt,
			Options:     toEventOptions(a.Options),
		},
	}
}

func (f *attributeEventFactory) NewAttributeUpdatedOutboxMessage(ctx context.Context, a *attribute.Attribute) outbox.Message {
	return outbox.Message{
		Event: f.newAttributeUpdatedEvent(a),
		Key:   a.ID,
	}
}
