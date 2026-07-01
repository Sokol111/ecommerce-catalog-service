package kafka

import (
	"context"

	eventsv1 "github.com/Sokol111/ecommerce-catalog-service-api/gen/events/catalog/v1"
	apiEvents "github.com/Sokol111/ecommerce-catalog-service-api/pkg/events"
	"github.com/Sokol111/ecommerce-catalog-service/internal/application/product"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type productEventFactory struct{}

// newProductEventFactory creates a new ProductEventFactory
func newProductEventFactory() product.ProductEventFactory {
	return &productEventFactory{}
}

func toProductEventAttributeValue(pAttr product.AttributeValue) *eventsv1.AttributeValue {
	av := &eventsv1.AttributeValue{
		AttributeId:   pAttr.AttributeID,
		AttributeSlug: pAttr.AttributeSlug,
	}
	switch {
	case pAttr.OptionSlugValue != nil:
		av.Value = &eventsv1.AttributeValue_OptionSlugValue{OptionSlugValue: *pAttr.OptionSlugValue}
	case len(pAttr.OptionSlugValues) > 0:
		av.Value = &eventsv1.AttributeValue_OptionSlugValues{OptionSlugValues: &eventsv1.StringList{Values: pAttr.OptionSlugValues}}
	case pAttr.NumericValue != nil:
		av.Value = &eventsv1.AttributeValue_NumericValue{NumericValue: *pAttr.NumericValue}
	case pAttr.TextValue != nil:
		av.Value = &eventsv1.AttributeValue_TextValue{TextValue: *pAttr.TextValue}
	case pAttr.BooleanValue != nil:
		av.Value = &eventsv1.AttributeValue_BooleanValue{BooleanValue: *pAttr.BooleanValue}
	}
	return av
}

func toProductEventAttributes(productAttrs []product.AttributeValue) []*eventsv1.AttributeValue {
	if len(productAttrs) == 0 {
		return nil
	}
	return lo.Map(productAttrs, func(pAttr product.AttributeValue, _ int) *eventsv1.AttributeValue {
		return toProductEventAttributeValue(pAttr)
	})
}

func (f *productEventFactory) newProductUpdatedEvent(p *product.Product) *eventsv1.ProductUpdatedEvent {
	return &eventsv1.ProductUpdatedEvent{
		ProductId:   p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Quantity:    int32(p.Quantity),
		Enabled:     p.Enabled,
		Version:     int64(p.Version),
		ImageId:     p.ImageID,
		CategoryId:  p.CategoryID,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		ModifiedAt:  timestamppb.New(p.ModifiedAt),
		Attributes:  toProductEventAttributes(p.Attributes),
	}
}

func (f *productEventFactory) NewProductUpdatedOutboxMessage(ctx context.Context, p *product.Product) outbox.Message {
	event := f.newProductUpdatedEvent(p)
	return outbox.Message{
		Event: event,
		Key:   p.ID,
		Topic: apiEvents.TopicFor(event),
	}
}

func (f *productEventFactory) NewProductDeletedOutboxMessage(ctx context.Context, productID string) outbox.Message {
	event := &eventsv1.ProductDeletedEvent{
		ProductId: productID,
	}
	return outbox.Message{
		Event: event,
		Key:   productID,
		Topic: apiEvents.TopicFor(event),
	}
}
