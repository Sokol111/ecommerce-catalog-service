package fxconfig

import (
	fx_connect "github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/inbound/connect/fxconfig"
	fx_kafka "github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/outbound/kafka/fxconfig"
	fx_mongo "github.com/Sokol111/ecommerce-catalog-service/internal/infrastructure/outbound/mongo/fxconfig"
	fx_tenant_api "github.com/Sokol111/ecommerce-tenant-service-api/pkg/fxconfig"
	"go.uber.org/fx"
)

func NewInfrastructureModule() fx.Option {
	return fx.Options(
		fx_tenant_api.NewKafkaConsumerModule(),
		fx_tenant_api.NewTenantSlugsProviderModule(),
		fx_tenant_api.NewGrpcClientsModule(),
		fx_mongo.NewMongoModule(),
		fx_kafka.NewKafkaModule(),
		fx_connect.NewConnectModule(),
	)
}
