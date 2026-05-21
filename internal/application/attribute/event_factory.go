package attribute

import (
	"context"

	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
)

// AttributeEventFactory defines the port for creating attribute event outbox messages.
type AttributeEventFactory interface {
	NewAttributeUpdatedOutboxMessage(ctx context.Context, a *Attribute) outbox.Message
}
