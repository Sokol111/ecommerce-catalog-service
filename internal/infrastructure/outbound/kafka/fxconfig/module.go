package kafka

import (
	"github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/outbound/kafka"
	"go.uber.org/fx"
)

func NewKafkaModule() fx.Option {
	return fx.Options(
		fx.Provide(
			kafka.NewProductEventFactory,
			kafka.NewCategoryEventFactory,
			kafka.NewAttributeEventFactory,
		),
	)
}
