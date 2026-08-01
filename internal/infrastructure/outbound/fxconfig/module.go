package fxconfig

import (
	kafka "github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/outbound/kafka/fxconfig"
	mongo "github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/outbound/mongo/fxconfig"
	"go.uber.org/fx"
)

func NewOutboundInfrastructureModule() fx.Option {
	return fx.Options(
		mongo.NewMongoModule(),
		kafka.NewKafkaModule(),
	)
}
