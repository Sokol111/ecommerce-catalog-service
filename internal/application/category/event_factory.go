package category

import (
	"context"

	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
)

// CategoryEventFactory creates category events
type CategoryEventFactory interface {
	NewCategoryUpdatedOutboxMessage(ctx context.Context, c *Category) outbox.Message
}
