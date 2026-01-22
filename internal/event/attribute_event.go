package event

import (
	"context"
	"time"

	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-catalog-service-api/gen/events"
	"github.com/Sokol111/ecommerce-catalog-service/internal/domain/attribute"
	commonsevents "github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/events"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/observability/tracing"
	"github.com/google/uuid"
)

// AttributeEventFactory creates attribute events
type AttributeEventFactory interface {
	NewAttributeUpdatedOutboxMessage(ctx context.Context, a *attribute.Attribute) (outbox.Message, error)
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

func (f *attributeEventFactory) newAttributeUpdatedEvent(ctx context.Context, a *attribute.Attribute) *events.AttributeUpdatedEvent {
	traceID := tracing.GetTraceID(ctx)

	return &events.AttributeUpdatedEvent{
		Metadata: commonsevents.EventMetadata{
			EventID:   uuid.New().String(),
			EventType: events.EventTypeAttributeUpdated,
			Source:    "ecommerce-catalog-service",
			Timestamp: time.Now().UTC(),
			TraceID:   &traceID,
		},
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

func (f *attributeEventFactory) NewAttributeUpdatedOutboxMessage(ctx context.Context, a *attribute.Attribute) (outbox.Message, error) {
	e := f.newAttributeUpdatedEvent(ctx, a)

	return outbox.Message{
		Payload: e,
		EventID: e.Metadata.EventID,
		Key:     a.ID,
	}, nil
}
