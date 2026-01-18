package kafka

import (
	"github.com/Sokol111/ecommerce-catalog-service-api/gen/events"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/kafka/avro/mapping"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Invoke(registerSchemas),
	)
}

func registerSchemas(tm *mapping.TypeMapping) error {
	return tm.RegisterBindings(events.SchemaBindings)
}
