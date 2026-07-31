package product

import (
	"context"

	"github.com/Sokol111/ecommerce-commons/pkg/kafka/outbox"
)

// ProductEventFactory creates product events
type ProductEventFactory interface {
	NewProductUpdatedOutboxMessage(ctx context.Context, p *Product) outbox.Message
	NewProductDeletedOutboxMessage(ctx context.Context, productID string) outbox.Message
}
